package feed

import (
	"context"
	"log"

	"github.com/dgraph-io/travel/cmd/feeds/sydney/internal/places"
)

// Pull retrieves and stores the feed data for this API.
func Pull(log *log.Logger, apiKey string, dbHost string) error {
	ctx := context.Background()

	loc := places.Location{
		Lat:     -33.865143,
		Lng:     151.209900,
		Keyword: "Sydney",
		Radius:  5000,
	}
	result, err := places.Retrieve(ctx, apiKey, loc)
	if err != nil {
		return err
	}

	log.Print("place result\n", string(result))

	if err := places.Store(ctx, dbHost, result); err != nil {
		return err
	}

	return nil
}
