package hot

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/delaneyj/toolbelt/gomps"
	"github.com/go-chi/chi/v5"
	"github.com/valyala/fasttemplate"
)

type HotReloadOptions struct {
	URL   string
	Retry time.Duration
}

func Reload(opts *HotReloadOptions) (gomps.NODE, func(context.Context, *chi.Mux) error) {
	if opts == nil {
		opts = &HotReloadOptions{
			URL:   "/hot",
			Retry: 250 * time.Millisecond,
		}
	}

	t := fasttemplate.New(`
	console.log("hot reload script loaded")
let hotValue =  ''
const source = new EventSource('[[url]]');
source.onopen = (event) => {
	console.log("EventSource connected.");
}
source.onerror = (event) => {
	console.log("EventSource failed.");
	setTimeout(() => {
		window.location.reload()
	}, 5000)
}

`, "[[", "]]")
	s := t.ExecuteString(map[string]interface{}{
		"url": opts.URL,
	})
	hotScript := gomps.SCRIPT(gomps.TYPE("module"), gomps.RAW(s))

	hotRoute := func(setupCtx context.Context, router *chi.Mux) error {
		retryMS := int(opts.Retry / time.Millisecond)
		router.Get(opts.URL, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Content-Type", "text/event-stream")
			w.(http.Flusher).Flush()

			i := 0
			for {
				evt := fmt.Sprintf(
					"id:%X\nretry:%d\ndata:%x\n\n",
					i, retryMS, i,
				)
				w.Write([]byte(evt))
				w.(http.Flusher).Flush()
				time.Sleep(opts.Retry)
				i++
			}
		})

		return nil
	}

	return hotScript, hotRoute
}
