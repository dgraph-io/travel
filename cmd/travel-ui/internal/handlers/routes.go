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
func UI(build string, shutdown chan os.Signal, log *log.Logger, dgraph data.Dgraph) (*web.App, error) {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	index, err := newIndex(dgraph, cities)
	if err != nil {
		return nil, errors.Wrap(err, "loading index template")
	}
	app.Handle("GET", "/", index.handler)

	// Set the route to load assets.
	fs := http.FileServer(http.Dir("assets"))
	fs = http.StripPrefix("/assets/", fs)
	f := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		fs.ServeHTTP(w, r)
		return nil
	}
	app.Handle("GET", "/assets/*", f)

	// Set the route to load data for the graph.
	fetch := fetch{
		dgraph: dgraph,
	}
	app.Handle("GET", "/data/:city", fetch.data)

	return app, nil
}
