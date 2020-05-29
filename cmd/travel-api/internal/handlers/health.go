package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/web"
)

type check struct {
	build  string
	dgraph data.Dgraph
}

func (c *check) health(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: c.build,
	}

	// Wait for a second to see if the database is ready.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := data.Readiness(ctx, c.dgraph.URL, 100*time.Millisecond); err != nil {

		// If the database is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.
		health.Status = "db not ready"
		return web.Respond(ctx, w, health, http.StatusInternalServerError)
	}

	health.Status = "ok"
	return web.Respond(ctx, w, health, http.StatusOK)
}
