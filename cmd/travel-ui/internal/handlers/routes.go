package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/travel/internal/mid"
	"github.com/dgraph-io/travel/internal/platform/web"
)

// UI constructs an http.Handler with all application routes defined.
func UI(build string, shutdown chan os.Signal, log *log.Logger, apiHost string) (*web.App, error) {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	app.Handle("GET", "/", index)

	fs := http.FileServer(http.Dir("assets"))
	fs = http.StripPrefix("/assets/", fs)
	f := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		fs.ServeHTTP(w, r)
		return nil
	}
	app.Handle("GET", "/assets/*", f)

	fetch := fetch{
		apiHost: apiHost,
	}
	app.Handle("GET", "/data", fetch.data)

	return app, nil
}
