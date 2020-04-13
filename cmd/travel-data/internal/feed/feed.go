package feed

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/dgraph-io/travel/cmd/travel-data/internal/places"
)

// ErrFailed is used to report the program failed back to main
// so the correct error code is returned.
var ErrFailed = errors.New("feed failed")

// Pull retrieves and stores the feed data for this API.
func Pull(log *log.Logger, cityName, mapsKey, weatherKey, dbHost string) error {
	ctx := context.Background()

	// Construct a client value so we can search and store
	// the city data we need.
	client, err := places.NewClient(ctx, mapsKey, weatherKey, dbHost)
	if err != nil {
		log.Print("feed : Pull : NewClient : ERROR : ", err)
		return ErrFailed
	}

	// Construct a city so we can perform work against that city.
	city, err := places.NewCity(ctx, client, cityName, -33.865143, 151.209900)
	if err != nil {
		log.Print("feed : Pull : NewCity : ERROR : ", err)
		return ErrFailed
	}

	log.Printf("feed : Pull : SetCity : Set %q with ID %q in DB", city.Name, city.ID)

	// // Pull all the hotels for the configured city.
	// search := places.Search{
	// 	Keyword: "hotels",
	// 	Radius:  5000,
	// }

	// // For now we will test with 1 place.
	// for i := 0; i < 1; i++ {

	// 	// I hate this but we need to keep this non-idiomatic error
	// 	// variable because an io.EOF error means we are done but
	// 	// we did get data back to process.
	// 	log.Printf("feed : Pull : Search : Searching for %q", search.Keyword)
	// 	places, errRet := city.Search(ctx, search)
	// 	if errRet != nil && errRet != io.EOF {
	// 		log.Print("feed : Pull : Search : ERROR : ", errRet)
	// 		return ErrFailed
	// 	}
	// 	log.Printf("feed : Pull : Search : Result\n%+v", places)

	// 	// Store each individual place in Dgraph.
	// 	log.Printf("feed : Pull : Store : Adding %d Places", len(places))
	// 	for _, place := range places {
	// 		if err := city.Store(ctx, log, place); err != nil {
	// 			log.Printf("feed : Pull : Store : ERROR : %s : %+v", err, place)
	// 			continue
	// 		}
	// 		log.Printf("feed : Pull : Store : Success : %+v", place)
	// 	}

	// 	// If this was the last result, we are done.
	// 	if errRet == io.EOF {
	// 		break
	// 	}
	// }

	return nil
}
