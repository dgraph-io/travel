package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Readiness checks if the DB is ready to receive requests. It will attempt
// a check between each retryDelay duration specified. The context holds the
// total amount of time Readiness will wait to validate the DB is healthy.
func Readiness(ctx context.Context, apiHost string, retryDelay time.Duration) error {

	// Create a timer to pump the iteration of the loop.
	t := time.NewTimer(retryDelay)
	defer t.Stop()

	// We will try until the retryDelay time has exipired.
	for {

		// Define and execute a function to perform the health check call.
		err := func() error {

			// Construct a request to perform the health call.
			req, err := http.NewRequest(http.MethodGet, "http://"+apiHost+"/health", nil)
			if err != nil {
				return err
			}

			// The actual call to the database should happen within 100 milliseconds.
			ctx, cancel := context.WithTimeout(req.Context(), 100*time.Millisecond)
			defer cancel()
			req = req.WithContext(ctx)

			// Perform the health check.
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Check the status code to see if we bother to check
			// the response.
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("%s", resp.Status)
			}

			// Capture the response and decode.
			var result []struct {
				Status string
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return err
			}

			switch {
			case len(result) == 0:
				return errors.New("unknown status")
			case result[0].Status != "healthy":
				return fmt.Errorf("%s", result[0].Status)
			}

			return nil
		}()

		// If there is no error, then report health.
		if err == nil {
			return nil
		}

		// Wait before we try again.
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "timed out")
		case <-t.C:
			t.Reset(retryDelay)
		}
	}
}
