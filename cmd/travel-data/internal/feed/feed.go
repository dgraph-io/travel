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

	// Construct a client value so we can search and store
	// the city data we need.
	client, err := places.NewClient(ctx, apiKey, dbHost)
	if err != nil {
		return err
	}

	// Construct a city so we can perform work against that city.
	city, err := places.NewCity(ctx, client, "Sydney", -33.865143, 151.209900)
	if err != nil {
		return err
	}

	log.Printf("feed : Pull : SetCity : Set %q with ID %q in DB", city.Name, city.ID)

	// Pull all the hotels for Sydney, Australia.
	search := places.Search{
		Keyword: "hotels",
		Radius:  5000,
	}

	// For now we will test with 1 place.
	for i := 0; i < 1; i++ {

		// I hate this but we need to keep this non-idiomatic error
		// variable because an io.EOF error means we are done but
		// we did get data back to process.
		log.Printf("feed : Pull : Search : Searching for %q", search.Keyword)
		places, errRet := city.Search(ctx, search)
		if errRet != nil && errRet != io.EOF {
			return errRet
		}

		log.Printf("******************** place result ********************\n\n%+v\n\n", places)

		// Store the places in Dgraph.
		for _, place := range places {
			log.Printf("feed : Pull : Store : Adding place %q", place.Name)
			if err := city.Store(ctx, log, place); err != nil {
				return err
			}
		}

		// If this was the last result, we are done.
		if errRet == io.EOF {
			break
		}
	}

	return nil
}
