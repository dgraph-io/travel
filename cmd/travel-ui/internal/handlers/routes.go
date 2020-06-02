package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/mid"
	"github.com/dgraph-io/travel/internal/platform/web"
	"github.com/pkg/errors"
)

// These are the currently cities supported. To be replaced by a query.
var cities = []string{"miami", "new york", "sydney"}

// UI constructs an http.Handler with all application routes defined.
func UI(build string, shutdown chan os.Signal, log *log.Logger, dgraph data.Dgraph, browserEndpoint string, mapsKey string) (*web.App, error) {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register the index page for the website.
	index, err := newIndex(dgraph, browserEndpoint, cities, mapsKey)
	if err != nil {
		return nil, errors.Wrap(err, "loading index template")
	}
	app.Handle("GET", "/", index.handler)

	// Register the assets.
	fs := http.FileServer(http.Dir("assets"))
	fs = http.StripPrefix("/assets/", fs)
	f := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		fs.ServeHTTP(w, r)
		return nil
	}
	app.Handle("GET", "/assets/*", f)

	// Register health check endpoint.
	check := check{
		build:  build,
		dgraph: dgraph,
	}
	app.Handle(http.MethodGet, "/health", check.health)

	// Register data load endpoint.
	fetch := fetch{
		dgraph: dgraph,
	}
	app.Handle("GET", "/data/:city", fetch.data)

	return app, nil
}
