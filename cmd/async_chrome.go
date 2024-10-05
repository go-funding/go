package main

import (
	"bufio"
	"context"
	"fmt"
	"fuk-funding/go/config"
	"fuk-funding/go/ctx"
	"fuk-funding/go/engine/manualgooglechrome"
	"fuk-funding/go/utils"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"os"
	"sync"
	"time"
)

// Simple checker
var _ = CommandRunnable(AsyncChromeCommand{})

type AsyncChromeCommand struct{}

func (acc AsyncChromeCommand) CommandData() *cli.Command {
	return &cli.Command{
		Name:  "async-chrome",
		Usage: "Run chrome loader asynchronously",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Usage:    "Path to file containing URLs (one per line)",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "concurrent",
				Usage: "Number of concurrent page loads",
				Value: 10,
			},
			&cli.DurationFlag{
				Name:  "timeout",
				Usage: "Timeout for each page load",
				Value: 2 * time.Second,
			},
			&cli.BoolFlag{
				Name:  "not-mono",
				Usage: "Not mono",
			},
		},
	}
}

func (acc AsyncChromeCommand) Run(appCtx *ctx.Context, cliCtx *cli.Context) error {
	log := appCtx.Logger.Named("async-chrome")
	filePath := cliCtx.String("file")
	concurrent := cliCtx.Int("concurrent")
	timeout := cliCtx.Duration("timeout")

	parsedUrls, err := readURLsFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read URLs from file: %w", err)
	}

	sem := semaphore.NewWeighted(int64(concurrent))

	threadIDs := make(chan int, concurrent)
	for i := 0; i < concurrent; i++ {
		threadIDs <- i
	}

	var wg sync.WaitGroup

	for i, parsedUrl := range parsedUrls {
		if err := sem.Acquire(cliCtx.Context, 1); err != nil {
			log.Error("Failed to acquire semaphore", zap.Error(err))
			continue
		}

		wg.Add(1)
		go func(parsedUrl string) {
			threadID := <-threadIDs
			defer func() {
				threadIDs <- threadID
				sem.Release(1)
				wg.Done()
			}()

			if parsedUrl == "" {
				log.Warn("Empty URL")
				return
			}

			var baseDir = fmt.Sprintf(`./output/mono`)
			if cliCtx.Bool("not-mono") {
				baseDir = fmt.Sprintf(`./output/%v`, time.Now().UnixNano())
			}

			dirName := utils.UrlDirname(utils.MustUrlParse(parsedUrl))
			dirPath := fmt.Sprintf(`./%s/%s`, baseDir, dirName)
			if _, err := os.Stat(dirPath); err == nil {
				log.Infof("URL is already loaded %s in %s", parsedUrl, dirPath)
				return
			}

			if err := os.Mkdir(dirPath, 0755); err != nil {
				log.Error("Failed to create directory", dirPath, err)
				return
			}

			if err := loadURL(parsedUrl, timeout, log, threadID, baseDir); err != nil {
				log.Error("Failed to load URL", parsedUrl, err)
			}

			log.Infof("DONE URL %d/%d %s", i+1, len(parsedUrls), parsedUrl)
		}(parsedUrl)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	return nil
}

func readURLsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func loadURL(url string, timeout time.Duration, log *zap.SugaredLogger, threadId int, baseDir string) error {
	// Add userId to userDataDit
	userDir := fmt.Sprintf("@user-group/user-%d", threadId)
	userDataDir := fmt.Sprintf("./output/%s", userDir)

	err := manualgooglechrome.Run(context.Background(), log, manualgooglechrome.ChromeOptions{
		InitialHref: url,
		UserDataDir: userDataDir,
		BaseDir:     baseDir,
		Headless:    true,
		Timeout:     timeout,

		IgnoreMimeType:             config.IgnoredMimeTypes,
		IgnoredHostsWithSubdomains: config.IgnoredHostsWithSubdomains,
		IgnoreNetworkResponseTypes: config.IgnoredNetworkResponseTypes,
	})
	if err != nil {
		log.Error(`chrome run`, zap.Error(err))
	}
	return nil
}
