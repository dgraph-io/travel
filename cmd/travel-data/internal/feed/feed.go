package feed

import (
	"context"
	"log"

	"github.com/dgraph-io/travel/cmd/travel-data/internal/places"
)

// Pull retrieves and stores the feed data for this API.
func Pull(log *log.Logger, apiKey string, dbHost string) error {
	ctx := context.Background()

	// Construct a places value so we can search and store
	// the place data we need.

	city := places.City{
		Name: "Sydney",
		Lat:  -33.865143, // This is the lat,lng for Sydney.
		Lng:  151.209900,
	}

	log.Print("feed : Pull : SetCity : Setting Sydney in DB")

	// Your new city is Sydney.
	// Creates the city node only if it doesn't exists in the database,
	// and sets the node UID of the city in p.cityUID
	// Create a new instane of places.New everytime you have a new city.
	_, err := places.New(ctx, city, apiKey, dbHost)
	if err != nil {
		return err
	}

	// Pull all the hotels for Sydney, Australia.
	// search := places.Search{
	// 	Lat:     city.Lat,
	// 	Lng:     city.Lng,
	// 	Keyword: "hotels",
	// 	Radius:  5000,
	// }

	// For now we will test with 1 place.
	// for i := 0; i < 1; i++ {

	// 	// I hate this but we need to keep this non-idiomatic error
	// 	// variable because an io.EOF error means we are done but
	// 	// we did get data back to process.
	// 	log.Printf("feed : Pull : Search : Searching for %q", search.Keyword)
	// 	places, errRet := p.Search(ctx, search)
	// 	if errRet != nil && errRet != io.EOF {
	// 		return errRet
	// 	}

	// 	log.Printf("******************** place result ********************\n\n%+v\n\n", places)

	// 	// Store the places in Dgraph.
	// 	for _, place := range places {
	// 		log.Printf("feed : Pull : Store : Adding place %q", place.Name)
	// 		if err := p.Store(ctx, log, place); err != nil {
	// 			return err
	// 		}
	// 	}

	// 	// If this was the last result, we are done.
	// 	if errRet == io.EOF {
	// 		break
	// 	}
	// }

	return nil
}
