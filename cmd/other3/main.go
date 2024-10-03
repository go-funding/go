package main

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
	"log"
	"net/url"
	"os"
	"path"
	"reflect"
	"sync"
	"time"
)

var UNIQUE_ID = time.Now().UnixNano()

var ignoredHosts = []string{
	"google-analytics.com",
	"maps.googleapis.com",
	"fonts.gstatic.com",
	"googletagmanager.com",
}

var ignoreNetworkResponseTypes = []network.ResourceType{
	"Stylesheet",
	"Font",
	"Image",
}

var ignoredMimeType = []string{
	"image/vnd.microsoft.icon",
}

func storeFile(relativePathDeep string, contents []byte) error {
	fullPath := path.Join("output", fmt.Sprint(UNIQUE_ID), relativePathDeep)

	if err := os.MkdirAll(path.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	if err := os.WriteFile(fullPath, contents, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func main() {
	allocCtx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.Flag("auto-open-devtools-for-tabs", true),
	)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Create a WaitGroup to wait for all tabs to be processed
	var wg sync.WaitGroup

	loadData := func(requestId network.RequestID, requestUrl string) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			urlData, err := url.Parse(requestUrl)

			c := chromedp.FromContext(ctx)
			body, err := network.GetResponseBody(requestId).Do(cdp.WithExecutor(ctx, c.Target))
			if err != nil {
				log.Println(err)
				return
			}

			filePath, err := url.JoinPath(urlData.Host, urlData.Path)
			if err != nil {
				log.Println(err)
				return
			}

			err = storeFile(filePath, body)
			if err != nil {
				return
			}
		}()
	}

	chromedp.ListenTarget(ctx, func(ev any) {
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
				log.Println(err)
				return
			}

			if fp.Contains(ignoredHosts, urlData.Host) || fp.SubdomainOf(ignoredHosts, urlData.Host) {
				return
			}

			if fp.Contains(ignoreNetworkResponseTypes, e.Type) {
				return
			}

			if fp.Contains(ignoredMimeType, e.Response.MimeType) {
				return
			}

			log.Println(
				"[*network.EventResponseReceived]",
				fmt.Sprintf("Type: %s", e.Type),
				fmt.Sprintf("URL: %s", e.Response.URL),
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
			log.Println(e)
			fmt.Println("Unmapped", reflect.TypeOf(e))
		}
	})

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://app.csquared.global/"),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Browser opened. You can now interact with the page and open new tabs.")
	log.Println("Close a tab when you're ready to download its resources.")

	time.Sleep(5 * time.Second)
	wg.Wait()

	log.Println("All tabs processed. Exiting.")
}
