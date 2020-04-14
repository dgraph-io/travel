package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Readiness checks if Dgraph is ready to receive requests. It will attempt
// to check the server for readiness over the specified amount of rety
// duration.
func Readiness(apiHost string, retryDelay time.Duration) error {

	// We will only attempt the ready call five times. Once right away and
	// then we will respect the retryDelay provided.
	const attempts = 5
	delay := retryDelay / attempts

	// We will try until the retryDelay time has exipired.
	for i := 1; i <= attempts; i++ {

		// After the first attempt, wait for before we try again.
		if i > 1 {
			time.Sleep(delay)
		}

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

		// Check for errors resulting in the call.
		switch {
		case err != nil && i < attempts:
			continue
		case err != nil && i == attempts:
			return fmt.Errorf("not healthy : %s", err)
		}

		// We are health so break and return.
		break
	}

	return nil
}
