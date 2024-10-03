package manualgooglechrome

import (
	"context"
	"fmt"
	"fuk-funding/go/fp"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
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
	"strings"
	"sync"
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

func targetListener(browserContext context.Context, log *zap.SugaredLogger, wg *sync.WaitGroup, options ChromeOptions) func(ev any) {
	loadData := func(requestId network.RequestID, requestUrl string) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			urlData, err := url.Parse(requestUrl)

			c := chromedp.FromContext(browserContext)
			body, err := network.GetResponseBody(requestId).Do(cdp.WithExecutor(browserContext, c.Target))
			if err != nil {
				log.Info(zap.Error(err))
				return
			}

			filePath, err := url.JoinPath(urlData.Host, urlData.Path)
			if err != nil {
				log.Info(zap.Error(err))
				return
			}

			if strings.HasSuffix(filePath, "/") {
				filePath = filePath + "index"
			}

			err = fp.StoreFileRecursive(path.Join(options.BaseDir, filePath), body)
			if err != nil {
				log.Info(zap.Error(err))
				return
			}
		}()
	}

	return func(ev any) {
		switch e := ev.(type) {
		case *cdproto.Message:
			//log.Println("[*cdproto.Message]", "Method: ", string(e.Method), "Result: ", string(e.Result), "Params: ", string(e.Params))
			return

		case *target.EventTargetCreated:
			//log.Println("[*target.EventTargetCreated]", "TargetInfo: ", e.TargetInfo.Type, e.TargetInfo.Title)
			return
		case *target.EventTargetInfoChanged:
			//log.Println("[*target.EventTargetInfoChanged]", "TargetInfo: ", e.TargetInfo.Type, e.TargetInfo.Title)
			return

		case *log2.EventEntryAdded:
			//log.Println("[*log2.EventEntryAdded]", "Entry: ", e.Entry)
			return

		case *runtime.EventExecutionContextCreated:
			//log.Println("[*runtime.EventExecutionContextCreated]", "Name: ", string(e.Context.AuxData))
			return
		case *runtime.EventExecutionContextsCleared:
			//log.Println("[*runtime.EventExecutionContextsCleared]")
			return
		case *runtime.EventConsoleAPICalled:
			//log.Println("[*runtime.EventConsoleAPICalled]")
			return

		case *page.EventLifecycleEvent:
			//log.Println("[*page.EventLifecycleEvent]", "Name: ", e.Name)
			return
		case *page.EventLoadEventFired:
			//log.Println("[*page.EventLoadEventFired]", "Timestamp: ", e.Timestamp)
			return
		case *page.EventFrameStoppedLoading:
			//log.Println("[*page.EventFrameStoppedLoading]", "FrameID: ", e.FrameID)
			return
		case *page.EventFrameNavigated:
			//log.Println("[*page.EventFrameNavigated]", "Frame: ", e.Frame)
			return
		case *page.EventNavigatedWithinDocument:
			//log.Println("[*page.EventNavigatedWithinDocument]", "E: ", e)
			return
		case *page.EventDomContentEventFired:
			//log.Println("[*page.EventDomContentEventFired]", "Timestamp: ", e.Timestamp)
			return
		case *page.EventFrameStartedLoading:
			//log.Println("[*page.EventFrameStartedLoading]", "FrameID: ", e.FrameID)
			return
		case *page.EventFrameResized:
			//log.Println("[*page.EventFrameResized]", "E: ", e)
			return

		case *network.EventLoadingFinished:
			//log.Println("[*network.EventLoadingFinished]", "RequestID: ", e.RequestID)
			return

		case *network.EventResponseReceived:
			urlData, err := url.Parse(e.Response.URL)
			if err != nil {
				log.Info(zap.Error(err))
				return
			}

			if fp.Contains(options.IgnoredHostsWithSubdomains, urlData.Host) || fp.SubdomainOf(options.IgnoredHostsWithSubdomains, urlData.Host) {
				return
			}

			if fp.Contains(options.IgnoreNetworkResponseTypes, e.Type) {
				return
			}

			if fp.Contains(options.IgnoreMimeType, e.Response.MimeType) {
				return
			}

			log.Debugf(
				"[*network.EventResponseReceived] Type: %s, URL: %s",
				e.Type.String(),
				e.Response.URL,
			)

			loadData(e.RequestID, e.Response.URL)
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

		// Ignored network events
		case *network.EventPolicyUpdated,
			*network.EventResourceChangedPriority,
			*network.EventRequestServedFromCache,
			*network.EventLoadingFailed:
			return

		case *dom.EventChildNodeInserted,
			*dom.EventChildNodeCountUpdated,
			*dom.EventAttributeModified,
			*dom.EventDocumentUpdated:
			return

		case *css.EventMediaQueryResultChanged,
			*css.EventStyleSheetAdded,
			*css.EventFontsUpdated,
			*css.EventStyleSheetRemoved:
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
		chromedp.NoSandbox,
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
		network.Enable(),
		chromedp.Navigate(options.InitialHref),
	); err != nil {
		return err
	}

	log.Debugf(`Navigated to %s`, options.InitialHref)
	log.Debugf(`All the information will be loaded to %s`, options.BaseDir)
	log.Debugf(`When you are done, please click Ctrl+C to stop the process`)

	// wait for the user to stop the process
	<-ctx.Done()
	wg.Wait()

	return nil
}
