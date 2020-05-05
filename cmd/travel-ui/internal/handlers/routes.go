package handlers

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/travel/internal/mid"
	"github.com/dgraph-io/travel/internal/platform/web"
	"github.com/pkg/errors"
)

// UI constructs an http.Handler with all application routes defined.
func UI(build string, shutdown chan os.Signal, log *log.Logger) (*web.App, error) {
	if err := loadTemplate("index", "index.html"); err != nil {
		return nil, errors.Wrap(err, "unable to load template")
	}

	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	fs := http.FileServer(http.Dir("static"))
	h := http.StripPrefix("/static/", fs)
	static := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		h.ServeHTTP(w, r)
		return nil
	}
	app.Handle("GET", "/static/*", static)

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		markup := executeTemplate("index", nil)
		io.WriteString(w, string(markup))
		return nil
	}
	app.Handle("GET", "/", handler)

	return app, nil
}
