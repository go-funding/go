package manualgooglechrome

import (
	"context"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/inspector"
	log2 "github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

func proceedDomEvents(ev any) (err error, found bool) {
	switch ev.(type) {
	case *dom.EventChildNodeInserted,
		*dom.EventChildNodeCountUpdated,
		*dom.EventAttributeModified,
		*dom.EventDocumentUpdated,
		*dom.EventChildNodeRemoved,
		*dom.EventSetChildNodes,
		*dom.EventShadowRootPopped,
		*dom.EventShadowRootPushed,
		*dom.EventInlineStyleInvalidated,
		*dom.EventPseudoElementAdded,
		*dom.EventPseudoElementRemoved:
		return nil, true
	}
	return nil, false
}

func proceedNetworkEvents(ev any) (err error, found bool) {
	switch ev.(type) {
	case *network.EventResourceChangedPriority:
		return nil, true
	}
	return nil, false
}

func proceedCssEvents(ev any) (err error, found bool) {
	switch ev.(type) {
	case *css.EventMediaQueryResultChanged,
		*css.EventStyleSheetAdded,
		*css.EventFontsUpdated,
		*css.EventStyleSheetRemoved,
		*css.EventStyleSheetChanged:
		return nil, true
	}
	return nil, false
}

func proceedTargetEvents(ev any) (err error, found bool) {
	switch ev.(type) {
	case *target.EventTargetCreated,
		*target.EventAttachedToTarget,
		*target.EventTargetInfoChanged,
		*target.EventDetachedFromTarget,
		*target.EventTargetDestroyed:
		return nil, true
	}
	return nil, false
}

func proceedRuntimeEvents(ev any) (err error, found bool) {
	switch ev.(type) {
	case *runtime.EventExecutionContextCreated,
		*runtime.EventExecutionContextsCleared,
		*runtime.EventExceptionThrown,
		*runtime.EventConsoleAPICalled,
		*runtime.EventExecutionContextDestroyed:
		return nil, true
	}
	return nil, false
}

func proceedPageEvents(ev any) (err error, found bool) {
	switch ev.(type) {
	case *page.EventFrameDetached,
		*page.EventFrameRequestedNavigation,
		*page.EventLifecycleEvent,
		*page.EventLoadEventFired,
		*page.EventFrameStoppedLoading,
		*page.EventFrameNavigated,
		*page.EventNavigatedWithinDocument,
		*page.EventDomContentEventFired,
		*page.EventFrameStartedLoading,
		*page.EventFrameResized,
		*page.EventFrameAttached:
		return nil, true
	}
	return nil, false
}

func proceedOtherEvents(browserContext context.Context, log *zap.SugaredLogger) func(any) (err error, found bool) {
	return func(ev any) (err error, found bool) {
		switch e := ev.(type) {
		case *inspector.EventDetached:
			return

		case *network.EventWebSocketFrameSent,
			*network.EventWebSocketCreated:
			return

		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(browserContext)
				err := fetch.ContinueRequest(e.RequestID).Do(cdp.WithExecutor(browserContext, c.Browser))
				if err != nil {
					log.Error(err)
					return
				}
			}()
			return nil, true

		case *cdproto.Message:
			if err != nil {
				return err, true
			}

			if len(e.Result) == 2 && string(e.Result) == "{}" {
				return nil, true
			}

			return nil, true

		case *log2.EventEntryAdded:
			return nil, true
		}

		return nil, false
	}
}
