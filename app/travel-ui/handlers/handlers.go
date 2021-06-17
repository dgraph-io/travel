// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"context"
	"expvar"
	"log"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/sys/metrics"
	"github.com/dgraph-io/travel/business/web/mid"
	"github.com/dgraph-io/travel/foundation/web"
	"github.com/pkg/errors"
)

// DebugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux. Using the
// DefaultServerMux would be a security risk since a dependency could inject a
// handler into our service without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// DebugMux registers all the debug standard library routes and then custom
// debug application routes for the service. This bypassing the use of the
// DefaultServerMux. Using the DefaultServerMux would be a security risk since
// a dependency could inject a handler into our service without us knowing it.
func DebugMux(build string, gqlConfig data.GraphQLConfig) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register the check endpoints.
	cg := checkGroup{
		build:     build,
		gqlConfig: gqlConfig,
	}
	mux.HandleFunc("/debug/readiness", cg.readiness)

	return mux
}

// UIMux constructs an http.Handler with all application routes defined.
func UIMux(build string, shutdown chan os.Signal, log *log.Logger, metrics *metrics.Metrics, gqlConfig data.GraphQLConfig, browserEndpoint string, mapsKey string) (*web.App, error) {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(metrics), mid.Panics(log))

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

	// Register data load endpoint.
	fg := fetchGroup{
		log:       log,
		gqlConfig: gqlConfig,
	}
	app.Handle(http.MethodGet, "/data/:city", fg.data)

	return app, nil
}
