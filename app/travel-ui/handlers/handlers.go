// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/mid"
	"github.com/dgraph-io/travel/foundation/web"
	"github.com/pkg/errors"
)

// UI constructs an http.Handler with all application routes defined.
func UI(build string, shutdown chan os.Signal, log *log.Logger, gqlConfig data.GraphQLConfig, browserEndpoint string, mapsKey string) (*web.App, error) {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register the index page for the website.
	ig, err := newIndex(gqlConfig, browserEndpoint, mapsKey)
	if err != nil {
		return nil, errors.Wrap(err, "loading index template")
	}
	app.Handle(http.MethodGet, "/", ig.handler)

	// Register the assets.
	fs := http.FileServer(http.Dir("assets"))
	fs = http.StripPrefix("/assets/", fs)
	f := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		fs.ServeHTTP(w, r)
		return nil
	}
	app.Handle(http.MethodGet, "/assets/*", f)

	// Register health check endpoint.
	cg := checkGroup{
		build:     build,
		gqlConfig: gqlConfig,
	}
	app.HandleDebug(http.MethodGet, "/readiness", cg.readiness)

	// Register data load endpoint.
	fg := fetchGroup{
		log:       log,
		gqlConfig: gqlConfig,
	}
	app.Handle(http.MethodGet, "/data/:city", fg.data)

	return app, nil
}
