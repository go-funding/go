package manualgooglechrome

import (
	"context"
	"fmt"
	"fuk-funding/go/config"
	"fuk-funding/go/fp"
	"fuk-funding/go/utils"
	"fuk-funding/go/utils/printer"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/debugger"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"
)

type ChromeOptions struct {
	InitialHref string
	UserDataDir string
	BaseDir     string
	Headless    bool
	Timeout     time.Duration

	IgnoreMimeType             []string
	IgnoredHostsWithSubdomains []string
	IgnoreNetworkResponseTypes []network.ResourceType
}

func storeFileRecursive(log *zap.SugaredLogger, baseDir string, urlData *url.URL, suffix string, content []byte) (err error) {
	hostDirName := utils.UrlDirname(urlData)

	filePath, err := url.JoinPath(hostDirName, urlData.Path)
	if err != nil {
		log.Info(err)
		return
	}

	if strings.HasSuffix(filePath, "/") {
		filePath = filePath + "index" + suffix
	} else {
		filePath = filePath + suffix
	}

	err = fp.StoreFileRecursive(path.Join(baseDir, filePath), content)
	if err != nil {
		log.Error(err)
		return
	}

	return
}

func targetListener(browserContext context.Context, log *zap.SugaredLogger, wg *sync.WaitGroup, options ChromeOptions) func(ev any) {
	networkStorage := NewChromeNetworkStorage()

	loadData := func(request *ChromeRequest) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := chromedp.FromContext(browserContext)

			var getResponse func(int, int, time.Duration) ([]byte, error)
			getResponse = func(maxRetries int, currentRetry int, sleepDuration time.Duration) ([]byte, error) {
				body, err := network.GetResponseBody(request.Id).Do(cdp.WithExecutor(browserContext, c.Target))
				if err != nil {
					if currentRetry > maxRetries {
						return nil, err
					}

					if v, ok := err.(*cdproto.Error); ok {
						if v.Code == -32000 {
							time.Sleep(sleepDuration)
							return getResponse(maxRetries, currentRetry+1, sleepDuration*2)
						}
					}
				}

				return body, err
			}

			saveResource(log, ResourceInformation{
				Url: request.RequestUrl,
				GetContent: func() (body []byte, err error) {
					return getResponse(3, 0, 100*time.Millisecond)
				},
				Type:       request.Type,
				StorageDir: options.BaseDir,
			}, options)

			if request.Type != network.ResourceTypeScript {
				return
			}

			if request.scriptId != nil && request.SourceMapURL != "" {
				// Get the source map
				// Get the source map
				sourceMapURL := request.SourceMapURL
				if !strings.HasPrefix(sourceMapURL, "http") {
					baseURL, err := url.Parse(request.RequestUrl)
					if err != nil {
						log.Error("Failed to parse script URL", err)
						return
					}
					sourceMapURL = baseURL.ResolveReference(&url.URL{Path: sourceMapURL}).String()
				}

				sourceMapBody, err := getSourceMap(browserContext, sourceMapURL)
				if err != nil {
					log.Warn("Failed to get source map using CDP, trying HTTP fallback: ", err)
					sourceMapBody, err = getSourceMapHTTP(sourceMapURL)
					if err != nil {
						log.Error("Failed to get source map using HTTP fallback: ", err)
						return
					}
				}

				saveResource(log, ResourceInformation{
					Url: sourceMapURL,
					GetContent: func() (body []byte, err error) {
						return sourceMapBody, nil
					},
					Type:       request.Type,
					StorageDir: options.BaseDir,
				}, options)
			}
		}()
	}

	return func(ev any) {
		if reaction := networkStorage.ReactOn(ev); reaction.IsReacted() {
			if reaction.Request() != nil {
				go func() {
					// Wait for the request to be finished
					reaction.Request().Wait()
					loadData(reaction.Request())
				}()
			}
			return
		}

		// Print event type
		processors := []func(any) (error, bool){
			proceedNetworkEvents,
			proceedDomEvents,
			proceedCssEvents,
			proceedTargetEvents,
			proceedRuntimeEvents,
			proceedPageEvents,
			proceedOtherEvents(browserContext, log),
		}

		for _, processor := range processors {
			var err error
			var found bool

			err, found = processor(ev)
			if found || err != nil {
				if err != nil {
					log.Error(err)
				}
				return
			}
		}

		switch e := ev.(type) {
		default:
			spew.Dump(e)
			fmt.Println("Unmapped", reflect.TypeOf(e))
		}
	}
}

func Run(ctx context.Context, baseLog *zap.SugaredLogger, options ChromeOptions) error {
	log := baseLog.Named(`browser[chrome]`)

	log.Debugf(`All the information will be loaded to %s`, color.RedString(options.BaseDir))

	allocCtx, cancel := chromedp.NewExecAllocator(
		ctx,
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,

		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("auto-open-devtools-for-tabs", true),
		chromedp.Flag("headless", options.Headless),
		chromedp.Flag("user-data-dir", options.UserDataDir),

		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-software-rasterizer", true),
	)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(s string, i ...interface{}) {
		log.Logf(zap.DebugLevel, s, i...)
	}))
	defer cancel()

	var wg sync.WaitGroup
	chromedp.ListenTarget(taskCtx, targetListener(taskCtx, log, &wg, options))

	if err := chromedp.Run(
		taskCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := debugger.Enable().Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	); err != nil {
		return err
	}

	if err := chromedp.Run(
		taskCtx,
		network.Enable(),
		chromedp.Navigate(options.InitialHref),
		chromedp.WaitReady("body"),
		chromedp.Reload(),
	); err != nil {
		return err
	}

	if options.Timeout > 0 {
		time.Sleep(options.Timeout)
	} else {
		// wait for the user to stop the process
		log.Debugf(`When you are done, please click %s to stop the process`, color.HiBlueString(`Ctrl+C`))
		<-ctx.Done()
	}

	wg.Wait()

	return nil
}

type ResourceInformation struct {
	Url        string
	GetContent func() ([]byte, error)
	Type       network.ResourceType
	StorageDir string
}

func cropUrl(url string) string {
	if len(url) > 200 {
		return url[:50] + "..."
	}

	return url
}

func saveResource(log *zap.SugaredLogger, info ResourceInformation, options ChromeOptions) {
	if strings.HasPrefix(info.Url, "data:") {
		mimeType := strings.Split(info.Url, ";")[0][5:]
		if fp.Contains(options.IgnoreMimeType, mimeType) {
			return

			log.Debugf(
				"%s mime type %s for URL %s",
				color.WhiteString("Ignore resource"),
				mimeType,
				cropUrl(info.Url),
			)
			return
		}
	}

	if fp.Contains(options.IgnoreNetworkResponseTypes, info.Type) {
		return

		log.Debugf(
			"%s type %s for URL %s",
			color.WhiteString("Ignore resource"),
			printer.PadStringsToLength(10, info.Type.String()),
			cropUrl(info.Url),
		)
		return
	}

	// config.IgnoredPathPatterns []*regexp.Regexp
	for _, pathPattern := range config.IgnoredPathPatterns {
		if pathPattern.MatchString(info.Url) {
			return

			log.Debugf(
				"%s URL %s",
				color.WhiteString("Ignore resource"),
				cropUrl(info.Url),
			)
			return
		}
	}

	urlData, err := url.Parse(info.Url)
	if err != nil {
		log.Error(err)
		return
	}

	if fp.Contains(options.IgnoredHostsWithSubdomains, urlData.Host) ||
		fp.SubdomainOf(options.IgnoredHostsWithSubdomains, urlData.Host) {
		return
	}

	content, err := info.GetContent()
	if err != nil {
		log.Error(err)
		return
	}

	err = storeFileRecursive(log, options.BaseDir, urlData, "", content)
	if err != nil {
		log.Error("failed to store file: ", err)
		return
	}

	log.Debugf("%s %s", color.GreenString("ResponseSaved"), cropUrl(info.Url))
}

func getSourceMap(ctx context.Context, sourceMapURL string) ([]byte, error) {
	var responseBody []byte
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			c := chromedp.FromContext(ctx)
			params := network.GetResponseBody(network.RequestID(sourceMapURL))
			body, err := params.Do(cdp.WithExecutor(ctx, c.Target))
			if err != nil {
				return err
			}
			responseBody = body
			return nil
		}),
	)
	return responseBody, err
}

func getSourceMapHTTP(sourceMapURL string) ([]byte, error) {
	resp, err := http.Get(sourceMapURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
