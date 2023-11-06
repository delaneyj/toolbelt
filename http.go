package toolbelt

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/CAFxX/httpcompression"
	"github.com/cenkalti/backoff"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func RunHotReload(port int, onStartPath string) CtxErrFunc {
	return func(ctx context.Context) error {
		onStartPath = strings.TrimPrefix(onStartPath, "/")
		localHost := fmt.Sprintf("http://localhost:%d", port)
		localURLToLoad := fmt.Sprintf("%s/%s", localHost, onStartPath)

		// Make sure page is ready before we start
		backoff := backoff.NewExponentialBackOff()
		for {
			if _, err := http.Get(localURLToLoad); err == nil {
				break
			}

			d := backoff.NextBackOff()
			log.Printf("Server not ready. Retrying in %v", d)
			time.Sleep(d)
		}

		// Launch browser in user mode, so we can reuse the same browser session
		wsURL := launcher.NewUserMode().MustLaunch()
		browser := rod.New().ControlURL(wsURL).MustConnect().NoDefaultDevice()

		// Get the current pages
		pages, err := browser.Pages()
		if err != nil {
			return fmt.Errorf("failed to get pages: %w", err)
		}
		var page *rod.Page
		for _, p := range pages {
			info, err := p.Info()
			if err != nil {
				return fmt.Errorf("failed to get page info: %w", err)
			}

			// If we already have the page open, just reload it
			if strings.HasPrefix(info.URL, localHost) {
				p.MustActivate().MustReload()
				page = p

				break
			}
		}
		if page == nil {
			// Otherwise, open a new page
			page = browser.MustPage(localURLToLoad)
		}

		slog.Info("page loaded", "url", localURLToLoad, "page", page.TargetID)
		return nil
	}
}

func CompressMiddleware() func(next http.Handler) http.Handler {
	compress, err := httpcompression.DefaultAdapter()
	if err != nil {
		panic(err)
	}
	return compress
}

type ServerSentEventsHandler struct {
	w                   http.ResponseWriter
	flusher             http.Flusher
	usingCompression    bool
	compressionMinBytes int
	shouldLogPanics     bool
	hasPanicked         bool
}

func NewSSE(w http.ResponseWriter, r *http.Request) *ServerSentEventsHandler {
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("response writer does not support flushing")
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	flusher.Flush()

	return &ServerSentEventsHandler{
		w:                   w,
		flusher:             flusher,
		usingCompression:    len(r.Header.Get("Accept-Encoding")) > 0,
		compressionMinBytes: 256,
		shouldLogPanics:     true,
	}
}

type SSEEvent struct {
	Id                string
	Event             string
	Data              []string
	Retry             time.Duration
	SkipMinBytesCheck bool
}

type SSEEventOption func(*SSEEvent)

func WithSSEId(id string) SSEEventOption {
	return func(e *SSEEvent) {
		e.Id = id
	}
}

func WithSSEEvent(event string) SSEEventOption {
	return func(e *SSEEvent) {
		e.Event = event
	}
}

func WithSSERetry(retry time.Duration) SSEEventOption {
	return func(e *SSEEvent) {
		e.Retry = retry
	}
}

func WithSSESkipMinBytesCheck(skip bool) SSEEventOption {
	return func(e *SSEEvent) {
		e.SkipMinBytesCheck = skip
	}
}

func (sse *ServerSentEventsHandler) Send(data string, opts ...SSEEventOption) {
	sse.SendMultiData([]string{data}, opts...)
}

func (sse *ServerSentEventsHandler) SendMultiData(dataArr []string, opts ...SSEEventOption) {
	if sse.hasPanicked && len(dataArr) > 0 {
		return
	}
	defer func() {
		// Can happen if the client closes the connection or
		// other middleware panics during flush (such as compression)
		// Not ideal, but we can't do much about it
		if r := recover(); r != nil && sse.shouldLogPanics {
			sse.hasPanicked = true
			log.Printf("recovered from panic: %v", r)
		}
	}()

	evt := SSEEvent{
		Id:    fmt.Sprintf("%d", NextID()),
		Event: "",
		Data:  dataArr,
		Retry: time.Second,
	}
	for _, opt := range opts {
		opt(&evt)
	}

	totalSize := 0

	if evt.Event != "" {
		evtFmt := fmt.Sprintf("event: %s\n", evt.Event)
		eventSize, err := sse.w.Write([]byte(evtFmt))
		if err != nil {
			panic(fmt.Sprintf("failed to write event: %v", err))
		}
		totalSize += eventSize
	}
	if evt.Id != "" {
		idFmt := fmt.Sprintf("id: %s\n", evt.Id)
		idSize, err := sse.w.Write([]byte(idFmt))
		if err != nil {
			panic(fmt.Sprintf("failed to write id: %v", err))
		}
		totalSize += idSize
	}
	if evt.Retry.Milliseconds() > 0 {
		retryFmt := fmt.Sprintf("retry: %d\n", evt.Retry.Milliseconds())
		retrySize, err := sse.w.Write([]byte(retryFmt))
		if err != nil {
			panic(fmt.Sprintf("failed to write retry: %v", err))
		}
		totalSize += retrySize
	}

	newLineBuf := []byte("\n")
	lastDataIdx := len(evt.Data) - 1
	for i, d := range evt.Data {
		dataFmt := fmt.Sprintf("data: %s", d)
		dataSize, err := sse.w.Write([]byte(dataFmt))
		if err != nil {
			panic(fmt.Sprintf("failed to write data: %v", err))
		}
		totalSize += dataSize

		if i != lastDataIdx {
			if !evt.SkipMinBytesCheck {
				newlineSuffixCount := 3
				if sse.usingCompression && totalSize+newlineSuffixCount < sse.compressionMinBytes {
					bufSize := sse.compressionMinBytes - totalSize - newlineSuffixCount
					buf := bytes.Repeat([]byte(" "), bufSize)
					if _, err := sse.w.Write(buf); err != nil {
						panic(fmt.Sprintf("failed to write data: %v", err))
					}
				}
			}
		}
		sse.w.Write(newLineBuf)
	}
	sse.w.Write([]byte("\n\n"))
	sse.flusher.Flush()
}
