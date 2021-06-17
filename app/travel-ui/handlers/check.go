package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgraph-io/travel/business/data"
)

type checkGroup struct {
	build     string
	gqlConfig data.GraphQLConfig
}

func (cg *checkGroup) readiness(w http.ResponseWriter, r *http.Request) {
	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: cg.build,
	}

	// Wait for a second to see if the database is ready.
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := data.Validate(ctx, cg.gqlConfig.URL, 100*time.Millisecond); err != nil {

		// If the database is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.
		health.Status = "db not ready"
		if err := response(w, http.StatusInternalServerError, health); err != nil {
			log.Println("liveness", "ERROR", err)
		}
	}

	health.Status = "ok"
	if err := response(w, http.StatusOK, health); err != nil {
		log.Println("liveness", "ERROR", err)
	}
}

func response(w http.ResponseWriter, statusCode int, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
