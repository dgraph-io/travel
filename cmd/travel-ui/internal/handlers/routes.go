package handlers

import (
	"log"
	"os"

	"github.com/dgraph-io/travel/internal/mid"
	"github.com/dgraph-io/travel/internal/platform/web"
)

// UI constructs an http.Handler with all application routes defined.
func UI(build string, shutdown chan os.Signal, log *log.Logger, apiHost string) (*web.App, error) {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	app.Handle("GET", "/", index)

	fetch := fetch{
		apiHost: apiHost,
	}
	app.Handle("GET", "/data", fetch.handler)

	return app, nil
}
