package handlers

import (
	"log"
	"os"

	"github.com/dgraph-io/travel/internal/mid"
	"github.com/dgraph-io/travel/internal/platform/web"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger) *web.App {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register health check endpoint. This route is not authenticated.
	check := Check{
		build: build,
	}
	app.Handle("GET", "/v1/health", check.Health)

	return app
}
