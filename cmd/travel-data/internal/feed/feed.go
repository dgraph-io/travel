package feed

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/places"
	"googlemaps.github.io/maps"
)

// ErrFailed is used to report the program failed back to main
// so the correct error code is returned.
var ErrFailed = errors.New("feed failed")

// Keys represents the set of keys needed for the different API's
// that are used to retrieve data.
type Keys struct {
	MapKey     string
	WeatherKey string
}

// Search represents a city and its coordinates. All fields must be
// populated for a Search to be successful.
type Search struct {
	Name    string
	Lat     float64
	Lng     float64
	Keyword string
	Radius  uint
}

// Work retrieves and stores the feed data for this API.
func Work(log *log.Logger, search Search, keys Keys, dbHost string) error {
	ctx := context.Background()

	// Construct a Data value for working with the database.
	data, err := data.New(dbHost)
	if err != nil {
		log.Print("feed : Pull : New Data : ERROR : ", err)
		return ErrFailed
	}

	// Validate the schema in the database before we start.
	if err := data.ValidateSchema(ctx); err != nil {
		log.Print("feed : Pull : ValidateSchema : ERROR : ", err)
		return ErrFailed
	}

	// Construct a Google maps client so we can search and store
	// the city data we need.
	client, err := maps.NewClient(maps.WithAPIKey(keys.MapKey))
	if err != nil {
		log.Print("feed : Pull : NewClient : ERROR : ", err)
		return ErrFailed
	}

	// Construct a city so we can perform searches against that city.
	city := places.City{
		Name: search.Name,
		Lat:  search.Lat,
		Lng:  search.Lng,
	}

	// Validate this city is in the database or add it.
	cityID, err := data.ValidateCity(ctx, city)
	if err != nil {
		log.Print("feed : Pull : ValidateCity : ERROR : ", err)
		return ErrFailed
	}

	// Construct a Filter to narrow down the places we want.
	filter := places.Filter{
		Keyword: search.Keyword,
		Radius:  search.Radius,
	}

	log.Printf("feed : Pull : Search : cityID[%s] city[%+v] filter[%+v]", cityID, city, filter)

	// For now we will test with 1 place.
	for i := 0; i < 1; i++ {

		// I hate this but we need to keep this non-idiomatic error
		// variable because an io.EOF error means we are done but
		// we did get data back to process.
		places, errRet := city.Search(ctx, client, &filter)
		if errRet != nil && errRet != io.EOF {
			log.Print("feed : Pull : Search : ERROR : ", errRet)
			return ErrFailed
		}
		log.Printf("feed : Pull : Search : Result\n%+v", places)

		// Store each individual place in the database.
		log.Printf("feed : Pull : Store : Adding %d Places", len(places))
		for _, place := range places {
			if err := data.StorePlace(ctx, log, cityID, place); err != nil {
				log.Printf("feed : Pull : Store : ERROR : %s : %+v", err, place)
				continue
			}
			log.Printf("feed : Pull : Store : Success : %+v", place)
		}

		// If this was the last result, we are done.
		if errRet == io.EOF {
			break
		}
	}

	return nil
}
