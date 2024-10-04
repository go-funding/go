package main

import (
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"fuk-funding/go/ctx"
)

// Simple checker
var _ = CommandRunnable(WatchSourcemapCommand{})

type WatchSourcemapCommand struct{}

func (wsc WatchSourcemapCommand) CommandData() *cli.Command {
	return &cli.Command{
		Name:  "watch-sourcemap",
		Usage: "Watch for .js.map files and run sourcemapper",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "dir",
				Usage:    "Base directory to watch for .js.map files",
				Required: true,
			},
			&cli.DurationFlag{
				Name:  "interval",
				Usage: "Polling interval for checking new .js.map files",
				Value: 5 * time.Second,
			},
		},
	}
}

func (wsc WatchSourcemapCommand) Run(appCtx *ctx.Context, cliCtx *cli.Context) error {
	log := appCtx.Logger.Named("watch-sourcemap")
	baseDir := cliCtx.String("dir")
	interval := cliCtx.Duration("interval")

	processedFiles := make(map[string]bool)

	for {
		pattern := filepath.Join(baseDir, "**", "*.js.map")
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			log.Error("Error globbing for .js.map files", zap.Error(err))
			continue
		}

		for _, match := range matches {
			if !processedFiles[match] {
				log.Info("New .js.map file detected", zap.String("file", match))
				err := runSourcemapper(match, log)
				if err != nil {
					log.Error("Error running sourcemapper", zap.Error(err))
				} else {
					processedFiles[match] = true
				}
			}
		}

		time.Sleep(interval)
	}
}

func runSourcemapper(mapFile string, log *zap.SugaredLogger) error {
	baseDir := filepath.Dir(mapFile)
	fileName := filepath.Base(mapFile)
	fileNameWithoutExt := strings.TrimSuffix(fileName, ".js.map")
	outputDir := filepath.Join(baseDir, fileNameWithoutExt)

	cmd := exec.Command("sourcemapper",
		"-output", "test",
		"-url", mapFile,
		"-output", outputDir,
	)

	log.Info("Running sourcemapper", zap.String("command", cmd.String()))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sourcemapper error: %w\nOutput: %s", err, output)
	}

	log.Info("Sourcemapper completed successfully", zap.String("output", string(output)))
	return nil
}
