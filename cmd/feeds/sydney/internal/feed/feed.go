package feed

import (
	"context"
)

// Pull retrieves and stores the feed data for this API.
func Pull(apiKey string, dbHost string) error {
	ctx := context.Background()

	loc := location{
		lat:     -33.865143,
		lng:     151.209900,
		keyword: "Sydney",
		radius:  5000,
	}
	jsonData, err := retrieveLocation(ctx, apiKey, loc)
	if err != nil {
		return err
	}

	if err := storeLocation(ctx, dbHost, jsonData); err != nil {
		return err
	}

	return nil
}
