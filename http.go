package toolbelt

import (
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
	Data              string
	Retry             time.Duration
	SkipMinBytesCheck bool
}

type SSEEventOption func(*SSEEvent)

func SSEEventId(id string) SSEEventOption {
	return func(e *SSEEvent) {
		e.Id = id
	}
}

func SSEEventEvent(event string) SSEEventOption {
	return func(e *SSEEvent) {
		e.Event = event
	}
}

func SSEEventRetry(retry time.Duration) SSEEventOption {
	return func(e *SSEEvent) {
		e.Retry = retry
	}
}

func SSEEventSkipMinBytesCheck(skip bool) SSEEventOption {
	return func(e *SSEEvent) {
		e.SkipMinBytesCheck = skip
	}
}

func (sse *ServerSentEventsHandler) Send(data string, opts ...SSEEventOption) {
	if sse.hasPanicked {
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
		Data:  data,
		Retry: time.Second,
	}
	for _, opt := range opts {
		opt(&evt)
	}

	prefixSB := strings.Builder{}
	if evt.Event != "" {
		prefixSB.WriteString(fmt.Sprintf("event: %s\n", evt.Event))
	}
	if evt.Id != "" {
		prefixSB.WriteString(fmt.Sprintf("id: %s\n", evt.Id))
	}
	prefixSB.WriteString(fmt.Sprintf("data: %s", evt.Data))

	suffixSB := strings.Builder{}
	if evt.Retry.Milliseconds() > 0 {
		suffixSB.WriteString(fmt.Sprintf("\nretry: %d", evt.Retry.Milliseconds()))
	}
	suffixSB.WriteString("\n\n")

	prefix := prefixSB.String()
	suffix := suffixSB.String()
	length := len(prefix) + len(suffix)

	sb := strings.Builder{}
	sb.WriteString(prefix)
	if evt.SkipMinBytesCheck {
		if sse.usingCompression && length < sse.compressionMinBytes {
			buf := make([]byte, sse.compressionMinBytes-length)
			for i := range buf {
				buf[i] = ' '
			}
			sb.Write(buf)
		}
	}
	sb.WriteString(suffix)
	eventFormatted := sb.String()
	sse.w.Write([]byte(eventFormatted))
	sse.flusher.Flush()
}
