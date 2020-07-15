// Package ready provides support for waiting for the database to be
// ready to receive commands.
package ready

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Validate checks if the DB is ready to receive requests. It will attempt
// a check between each retry interval specified. The context holds the
// total amount of time Readiness will wait to validate the DB is healthy.
func Validate(ctx context.Context, url string, retryInterval time.Duration) error {
	var t *time.Timer

	// We will try until the context timeout has expired.
	for {

		// If there is no error, then report health.
		if err := checkDB(ctx, url); err == nil {
			return nil
		}

		// Check if the timeout has expired.
		if ctx.Err() != nil {
			return errors.Wrap(ctx.Err(), "timed out")
		}

		// Create the timer if one doesn't exist.
		if t == nil {
			t = time.NewTimer(retryInterval)
		}

		// Wait before we try again or timeout.
		select {
		case <-ctx.Done():
			t.Stop()
			return errors.Wrap(ctx.Err(), "timed out")
		case <-t.C:
			t.Reset(retryInterval)
		}
	}
}

// checkDB attempts to validate if the database is ready.
func checkDB(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	url = fmt.Sprintf("%s/health", strings.TrimRight(url, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}

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
}
