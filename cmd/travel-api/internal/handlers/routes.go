package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/mid"
	"github.com/dgraph-io/travel/internal/platform/web"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, dbConfig data.DBConfig, emailConfig EmailConfig) *web.App {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register health check endpoint.
	check := check{
		build:    build,
		dbConfig: dbConfig,
	}
	app.Handle(http.MethodGet, "/v1/health", check.health)

	// Register the email endpoint.
	email := email{
		EmailConfig: emailConfig,
	}
	app.Handle(http.MethodPost, "/v1/email", email.send)

	return app
}
