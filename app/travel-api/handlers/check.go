package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/ready"
	"github.com/dgraph-io/travel/foundation/web"
)

type checkGroup struct {
	build     string
	gqlConfig data.GraphQLConfig
}

func (cg *checkGroup) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: cg.build,
	}

	// Wait for a second to see if the database is ready.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := ready.Validate(ctx, cg.gqlConfig.URL, 100*time.Millisecond); err != nil {

		// If the database is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.
		health.Status = "db not ready"
		return web.Respond(ctx, w, health, http.StatusInternalServerError)
	}

	health.Status = "ok"
	return web.Respond(ctx, w, health, http.StatusOK)
}
