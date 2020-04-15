package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Readiness checks if Dgraph is ready to receive requests. It will attempt
// to check the server for readiness every 1/2 second through out the
// specified amount of rety delay.
func Readiness(apiHost string, retryDelay time.Duration) error {

	// Minimal wait between attempts will be 1/2 second.
	delay := 500 * time.Millisecond

	// Calculate the total number of attempts we need to satisfy
	// the retry delay.
	attempts := int(retryDelay) / int(delay)

	// We will try until the retryDelay time has exipired.
	var err error
	for i := 1; i <= attempts; i++ {

		// After the first attempt, wait for before we try again.
		if i > 1 {
			time.Sleep(delay)
		}

		// Define and execute a function to perform the health check call.
		err = func() error {

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
	}

	return errors.Wrap(err, "not healthy")
}
