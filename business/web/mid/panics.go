package mid

import (
	"context"
	"log"
	"net/http"

	"github.com/dgraph-io/travel/business/sys/metrics"
	"github.com/dgraph-io/travel/foundation/web"
	"github.com/pkg/errors"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics(log *log.Logger) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Wrap this handler around the next one provided.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					err = errors.Errorf("PANIC: %v", rec)
					metrics.AddPanics(ctx)
				}
			}()

			// Call the next handler and set its return value in the err variable.
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
