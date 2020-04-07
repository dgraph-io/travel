package feed

import (
	"context"
	"io"
	"log"

	"github.com/dgraph-io/travel/cmd/travel-data/internal/places"
)

// Pull retrieves and stores the feed data for this API.
func Pull(log *log.Logger, apiKey string, dbHost string) error {
	ctx := context.Background()

	// Construct a places value so we can search and store
	// the place data we need.
	p, err := places.New(apiKey, dbHost)
	if err != nil {
		return err
	}

	// Pull all the hotels for Sydney, Australia.
	loc := places.Location{
		Lat:     -33.865143, // This is the lat,lng for Sydney.
		Lng:     151.209900,
		Keyword: "hotels",
		Radius:  5000,
	}

		// I hate this but we need to keep this non-idiomatic error
		// variable because an io.EOF error means we are done but
		// we did get data back to process.
		result, errRet := p.Retrieve(ctx, &loc)
		if errRet != nil && errRet != io.EOF {
			return errRet
		}
		log.Printf("******************** place result ********************\n\n%s\n\n", string(result))

		// Store the results in Dgraph.
		// if err := p.Store(ctx, result); err != nil {
		// 	return err
		// }

		// If this was the last result, we are done.
		if errRet == io.EOF {
			break
		}
	
	return nil
}
