package manualgooglechrome

import (
	"github.com/chromedp/cdproto/debugger"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"sync"
	"time"
)

type ChromeRequest struct {
	Id           network.RequestID
	wg           *sync.WaitGroup
	releaseTimer *time.Timer

	scriptId *runtime.ScriptID

	MimeType string

	RequestUrl      string
	RequestHeader   network.Headers
	ResponseHeaders network.Headers
	ResponseStatus  int64

	SourceMapURL string

	DocumentUrl string
	Type        network.ResourceType
}

func (cr *ChromeRequest) Wait() { cr.wg.Wait() }
func (cr *ChromeRequest) Release() {
	cr.wg.Done()
	go func() {
		if cr.Type != network.ResourceTypeScript || cr.scriptId == nil {
			cr.wg.Done()
			return
		}

		cr.releaseTimer = time.NewTimer(3 * time.Second)
		<-cr.releaseTimer.C
		cr.wg.Done()
	}()
}

func (cr *ChromeRequest) RegisterScriptId(id runtime.ScriptID, sourceMapURL string) {
	cr.scriptId = &id
	cr.SourceMapURL = sourceMapURL

	if cr.releaseTimer != nil {
		cr.releaseTimer.Stop()
	}
}

func NewChromeRequest(id network.RequestID) *ChromeRequest {
	cr := &ChromeRequest{
		Id: id,
		wg: &sync.WaitGroup{},
	}
	cr.wg.Add(2)
	return cr
}

type ChromeNetworkStorage struct {
	rw       sync.RWMutex
	Requests map[network.RequestID]*ChromeRequest
	Scripts  map[string]network.RequestID
}

func NewChromeNetworkStorage() *ChromeNetworkStorage {
	return &ChromeNetworkStorage{
		Requests: make(map[network.RequestID]*ChromeRequest),
		Scripts:  make(map[string]network.RequestID),
	}
}

type ReactOnResult struct {
	val bool
	req *ChromeRequest
}

func (r ReactOnResult) IsReacted() bool {
	return r.val
}

func (r ReactOnResult) Request() *ChromeRequest {
	return r.req
}

func (cns *ChromeNetworkStorage) ReactOn(ev any) (rr ReactOnResult) {
	cns.rw.Lock()
	defer cns.rw.Unlock()

	rr.val = true

	switch e := ev.(type) {
	case *network.EventRequestWillBeSentExtraInfo,
		*network.EventResponseReceivedExtraInfo,
		*network.EventResponseReceivedEarlyHints,
		*network.EventRequestServedFromCache,
		*network.EventWebSocketFrameSent,
		*network.EventWebSocketClosed,
		*network.EventWebSocketFrameReceived,
		*network.EventWebSocketFrameError,
		*network.EventWebSocketWillSendHandshakeRequest,
		*network.EventWebSocketHandshakeResponseReceived,
		*network.EventPolicyUpdated,
		*network.EventWebSocketCreated,
		*network.EventEventSourceMessageReceived:
		return

	case *network.EventResponseReceived:
		if req, ok := cns.Requests[e.RequestID]; ok {
			req.ResponseHeaders = e.Response.Headers
			req.MimeType = e.Response.MimeType
			req.ResponseStatus = e.Response.Status
		}
		return

	case *network.EventDataReceived:
		return

	case *debugger.EventScriptFailedToParse:
		return
	case *debugger.EventScriptParsed:
		if e.URL == "" || e.SourceMapURL == "" {
			return
		}

		if reqId, ok := cns.Scripts[e.URL]; ok {
			if req, ok := cns.Requests[reqId]; ok {
				req.RegisterScriptId(e.ScriptID, e.SourceMapURL)
			}
		}

		return

	case *network.EventRequestWillBeSent:
		crx := NewChromeRequest(e.RequestID)
		crx.RequestUrl = e.Request.URL
		crx.RequestHeader = e.Request.Headers
		crx.DocumentUrl = e.DocumentURL
		crx.Type = e.Type
		cns.Requests[e.RequestID] = crx
		cns.Scripts[e.Request.URL] = e.RequestID
		rr.req = crx
		return

	case *network.EventLoadingFinished:
		if req, ok := cns.Requests[e.RequestID]; ok {
			req.Release()
		}
		return

	case *network.EventLoadingFailed:
		if req, ok := cns.Requests[e.RequestID]; ok {
			req.Release()
		}
		return
	}

	rr.val = false
	return
}
