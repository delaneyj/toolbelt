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
	compress, errr := httpcompression.DefaultAdapter()
	if errr != nil {
		panic(errr)
	}
	return compress
}
