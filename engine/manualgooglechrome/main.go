package manualgooglechrome

import (
	"context"
	"fmt"
	"fuk-funding/go/fp"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/fetch"
	log2 "github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
	"net/url"
	"path"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"
)

type ChromeOptions struct {
	InitialHref string
	UserDataDir string
	BaseDir     string
	Headless    bool

	IgnoreMimeType             []string
	IgnoredHostsWithSubdomains []string
	IgnoreNetworkResponseTypes []network.ResourceType
}

func storeFileRecursive(log *zap.SugaredLogger, baseDir string, urlData *url.URL, suffix string, content []byte) (err error) {
	hostDirName := urlData.Host
	{
		result := strings.Split(hostDirName, ".")
		slices.Reverse(result)
		hostDirName = strings.Join(result, ".")
	}

	filePath, err := url.JoinPath(hostDirName, urlData.Path)
	if err != nil {
		log.Info(zap.Error(err))
		return
	}

	if strings.HasSuffix(filePath, "/") {
		filePath = filePath + "index" + suffix
	} else {
		filePath = filePath + suffix
	}

	err = fp.StoreFileRecursive(path.Join(baseDir, filePath), content)
	if err != nil {
		log.Info(zap.Error(err))
		return
	}

	return
}

func targetListener(browserContext context.Context, log *zap.SugaredLogger, wg *sync.WaitGroup, options ChromeOptions) func(ev any) {
	loadData := func(requestId network.RequestID, requestUrl string, suffix string) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			urlData, err := url.Parse(requestUrl)

			c := chromedp.FromContext(browserContext)

			var getResponse func(int, int, time.Duration) ([]byte, error)
			getResponse = func(maxRetries int, currentRetry int, sleepDuration time.Duration) ([]byte, error) {
				body, err := network.GetResponseBody(requestId).Do(cdp.WithExecutor(browserContext, c.Target))
				if err != nil {
					if currentRetry > maxRetries {
						log.Info(zap.Error(err), zap.String("RequestID", string(requestId)), zap.String("URL", requestUrl))
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

			body, err := getResponse(20, 0, 100*time.Millisecond)
			if err != nil {
				log.Info(zap.Error(err), zap.String("RequestID", string(requestId)), zap.String("URL", requestUrl))
				return
			}

			err = storeFileRecursive(log, options.BaseDir, urlData, suffix, body)
			if err != nil {
				log.Error(zap.Error(err))
				return
			}
		}()
	}

	return func(ev any) {
		switch e := ev.(type) {
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(browserContext)
				// note the executor should be "Browser" here
				fetch.ContinueRequest(e.RequestID).Do(cdp.WithExecutor(browserContext, c.Browser))
			}()

		case *cdproto.Message:
			return

		case *target.EventTargetCreated,
			*target.EventAttachedToTarget,
			*target.EventTargetInfoChanged,
			*target.EventDetachedFromTarget,
			*target.EventTargetDestroyed:
			return

		case *log2.EventEntryAdded:
			return

		case *runtime.EventExecutionContextCreated:
			return
		case *runtime.EventExecutionContextsCleared:
			return
		case *runtime.EventConsoleAPICalled:
			return
		case *runtime.EventExecutionContextDestroyed:
			return

		case *page.EventFrameDetached,
			*page.EventFrameRequestedNavigation:
			return
		case *page.EventLifecycleEvent:
			return
		case *page.EventLoadEventFired:
			return
		case *page.EventFrameStoppedLoading:
			return
		case *page.EventFrameNavigated:
			return
		case *page.EventNavigatedWithinDocument:
			return
		case *page.EventDomContentEventFired:
			return
		case *page.EventFrameStartedLoading:
			return
		case *page.EventFrameResized, *page.EventFrameAttached:
			return

		case *network.EventLoadingFinished:
			return

		case *network.EventLoadingFailed:
			return

		case *network.EventResponseReceived:
			urlData, err := url.Parse(e.Response.URL)
			if err != nil {
				log.Info(zap.Error(err))
				return
			}

			if fp.Contains(options.IgnoredHostsWithSubdomains, urlData.Host) ||
				fp.SubdomainOf(options.IgnoredHostsWithSubdomains, urlData.Host) ||
				fp.Contains(options.IgnoreNetworkResponseTypes, e.Type) ||
				fp.Contains(options.IgnoreMimeType, e.Response.MimeType) {
				return
			}

			log.Debugf(
				"[*network.EventResponseReceived] Type: %s, URL: %s",
				e.Type.String(),
				e.Response.URL,
			)

			var suffix string
			switch e.Type {
			case network.ResourceTypeDocument:
				suffix = ".html"
			}

			loadData(e.RequestID, e.Response.URL, suffix)
			return
		case *network.EventRequestWillBeSentExtraInfo:
			//log.Println("[*network.EventRequestWillBeSentExtraInfo]", "RequestID: ", e.RequestID)
			return
		case *network.EventDataReceived:
			//log.Println("[*network.EventDataReceived]", "RequestID: ", e.RequestID)
			return
		case *network.EventResponseReceivedExtraInfo:
			//log.Println("[*network.EventResponseReceivedExtraInfo]", "RequestID: ", e.RequestID)
			return
		case *network.EventRequestWillBeSent:
			//log.Println("[*network.EventRequestWillBeSent]", "RequestID: ", e.RequestID)
			return

		case *network.EventPolicyUpdated,
			*network.EventResourceChangedPriority,
			*network.EventRequestServedFromCache:
			return

		case *dom.EventChildNodeInserted,
			*dom.EventChildNodeCountUpdated,
			*dom.EventAttributeModified,
			*dom.EventDocumentUpdated,
			*dom.EventChildNodeRemoved,
			*dom.EventPseudoElementRemoved,
			*dom.EventSetChildNodes:
			return

		case *css.EventMediaQueryResultChanged,
			*css.EventStyleSheetAdded,
			*css.EventFontsUpdated,
			*css.EventStyleSheetRemoved,
			*css.EventStyleSheetChanged:
			return
		default:
			spew.Dump(e)
			fmt.Println("Unmapped", reflect.TypeOf(e))
		}
	}
}

func Run(ctx context.Context, baseLog *zap.SugaredLogger, options ChromeOptions) error {
	log := baseLog.Named(`[Chrome browser]`)

	allocCtx, cancel := chromedp.NewExecAllocator(
		ctx,
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("auto-open-devtools-for-tabs", true),
		chromedp.Flag("headless", options.Headless),
		chromedp.Flag("user-data-dir", options.UserDataDir),
	)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(s string, i ...interface{}) {
		log.Logf(zap.DebugLevel, s, i...)
	}))
	defer cancel()

	var wg sync.WaitGroup
	chromedp.ListenTarget(taskCtx, targetListener(taskCtx, log, &wg, options))

	// ensure that the browser process is started
	if err := chromedp.Run(
		taskCtx,
		chromedp.Navigate(options.InitialHref),
		chromedp.WaitReady("body"),
		chromedp.Reload(),
	); err != nil {
		return err
	}

	log.Debugf(`Navigated to %s`, options.InitialHref)
	log.Debugf(`All the information will be loaded to %s`, options.BaseDir)
	log.Debugf(`When you are done, please click Ctrl+C to stop the process`)

	go func() {
		// Download all sources every 10 seconds
		for {
			select {
			case <-time.After(10 * time.Second):
				err := downloadAllSources(taskCtx, log, options)
				if err != nil {
					log.Error(zap.Error(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// wait for the user to stop the process
	<-ctx.Done()
	wg.Wait()

	return nil
}

var now = time.Now().UnixNano()

func downloadAllSources(ctx context.Context, log *zap.SugaredLogger, options ChromeOptions) error {
	c := chromedp.FromContext(ctx)
	runCtx := cdp.WithExecutor(ctx, c.Target)

	log.Debugf(`Downloading all sources`)
	sources, err := page.GetResourceTree().Do(runCtx)
	if err != nil {
		return err
	}

	spew.Dump(sources)

	frameUrlParsed, err := url.Parse(sources.Frame.URL)
	if err != nil {
		return err
	}

	resultDir := fmt.Sprintf(`%s/sources/%v/%v`, options.BaseDir, frameUrlParsed.Host, now)
	for _, resource := range sources.Resources {
		log.Info(zap.String("Resource URL", resource.URL))
		content, err := page.GetResourceContent(sources.Frame.ID, resource.URL).Do(runCtx)
		if err != nil {
			log.Info(zap.Error(err))
			continue
		}

		urlData, err := url.Parse(resource.URL)
		if err != nil {
			log.Info(zap.Error(err))
			continue
		}

		err = storeFileRecursive(log, resultDir, urlData, "", content)
		if err != nil {
			log.Error(zap.Error(err))
			continue
		}
	}

	return nil
}
