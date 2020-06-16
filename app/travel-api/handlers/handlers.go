// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/loader"
	"github.com/dgraph-io/travel/business/mid"
	"github.com/dgraph-io/travel/foundation/web"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, dbConfig data.DBConfig, loaderConfig loader.Config) *web.App {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register the check endpoints.
	check := check{
		build:    build,
		dbConfig: dbConfig,
	}
	app.Handle(http.MethodGet, "/v1/health", check.health)

	// Register the feed endpoints.
	feed := feed{
		log:          log,
		dbConfig:     dbConfig,
		loaderConfig: loaderConfig,
	}
	app.Handle(http.MethodPost, "/v1/feed/upload", feed.upload)

	return app
}
