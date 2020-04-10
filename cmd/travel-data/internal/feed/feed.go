package feed

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/dgraph-io/travel/cmd/travel-data/internal/places"
)

// ErrFailed is used to report the program failed back to main
// so the correct error code is returned.
var ErrFailed = errors.New("feed failed")

// Pull retrieves and stores the feed data for this API.
func Pull(log *log.Logger, cityName string, apiKey string, dbHost string) error {
	ctx := context.Background()

	// Construct a client value so we can search and store
	// the city data we need.
	client, err := places.NewClient(ctx, apiKey, dbHost)
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

	// Pull all the hotels for the configured city.
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
			log.Print("feed : Pull : Search : ERROR : ", errRet)
			return ErrFailed
		}
		log.Printf("feed : Pull : Search : Result\n%+v", places)

		// Store each individual place in Dgraph.
		log.Printf("feed : Pull : Store : Adding %d Places", len(places))
		var out strings.Builder
		for _, place := range places {
			out.WriteString(fmt.Sprintf("%+v\n", place))
			if err := city.Store(ctx, log, place); err != nil {
				log.Print("feed : Pull : Store : ERROR : ", err)
				return ErrFailed
			}
		}
		log.Printf("feed : Pull : Store : Places Added\n%+v", out.String())

		// If this was the last result, we are done.
		if errRet == io.EOF {
			break
		}
	}

	return nil
}
