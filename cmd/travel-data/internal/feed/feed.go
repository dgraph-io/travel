package feed

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/dgraph-io/travel/internal/advisory"
	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/places"
	"github.com/dgraph-io/travel/internal/weather"
	"googlemaps.github.io/maps"
)

// ErrFailed is used to report the program failed back to main
// so the correct error code is returned.
var ErrFailed = errors.New("feed failed")

// Dgraph represents the IP and Ports we need to talk to the
// server for the different functions we need to perform.
type Dgraph struct {
	DBHost  string
	APIHost string
}

// Search represents a city and its coordinates. All fields must be
// populated for a Search to be successful.
type Search struct {
	CountryCode string
	CityName    string
	Lat         float64
	Lng         float64
	Keyword     string
	Radius      uint
}

// Keys represents the set of keys needed for the different API's
// that are used to retrieve data.
type Keys struct {
	MapKey     string
	WeatherKey string
}

// Work retrieves and stores the feed data for this API.
func Work(log *log.Logger, dgraph Dgraph, search Search, keys Keys) error {
	ctx := context.Background()

	// Construct a Data value for working with the database.
	db, err := data.NewDB(dgraph.DBHost, dgraph.APIHost)
	if err != nil {
		log.Printf("feed : Work : New Data : ERROR : %+v", err)
		return ErrFailed
	}

	// Validate the schema in the database before we start.
	if err := db.Validate.Schema(ctx); err != nil {
		log.Printf("feed : Work : ValidateSchema : ERROR : %+v", err)
		return ErrFailed
	}

	// Construct a Google maps client so we can search and store
	// the city data we need.
	client, err := maps.NewClient(maps.WithAPIKey(keys.MapKey))
	if err != nil {
		log.Printf("feed : Work : NewClient : ERROR : %+v", err)
		return ErrFailed
	}

	// Construct a city so we can perform searches against that city.
	city := places.City{
		Name: search.CityName,
		Lat:  search.Lat,
		Lng:  search.Lng,
	}

	// Validate this city is in the database or add it.
	cityID, err := db.Validate.City(ctx, city)
	if err != nil {
		log.Printf("feed : Work : ValidateCity : ERROR : %+v", err)
		return ErrFailed
	}

	log.Printf("feed : Work : Location : CityID[%s] Name[%s] Lat[%f] Lng[%f]", cityID, search.CityName, search.Lat, search.Lng)

	// Pull the weather for the city being specified.
	weather, err := weather.Search(ctx, keys.WeatherKey, search.Lat, search.Lng)
	if err != nil {
		log.Printf("feed : Work : Search Weather : ERROR : %+v", err)
		return ErrFailed
	}
	log.Printf("feed : Work : Search Weather : Result : %+v", weather)

	// Store the weather for the specified city.
	if err := db.Store.Weather(ctx, cityID, weather); err != nil {
		log.Printf("feed : Work : Store Weather : ERROR : %+v", err)
		return ErrFailed
	}

	// Pull the travel advisory for Australia.
	advisory, err := advisory.Search(ctx, search.CountryCode)
	if err != nil {
		log.Print("feed : Work : Search Weather : ERROR : ", err)
		return ErrFailed
	}
	log.Printf("feed : Work : Search Advisory : Result : %+v", advisory)

	// Store the travel advisory information.
	if err := db.Store.Advisory(ctx, cityID, advisory); err != nil {
		log.Print("feed : Work : Store Weather : ERROR : ", err)
		return ErrFailed
	}

	// Construct a Filter to narrow down the places we want.
	filter := places.Filter{
		Keyword: search.Keyword,
		Radius:  search.Radius,
	}

	log.Printf("feed : Work : Search Places : filter[%+v]", filter)

	// For now we will test with 1 place.
	for i := 0; i < 1; i++ {

		// Search for a collection of pages. Each new call to Search will
		// bring back a new page until io.EOF is reached.
		places, errRet := city.Search(ctx, client, &filter)
		if errRet != nil && errRet != io.EOF {
			log.Printf("feed : Work : Search Places : ERROR : %+v", errRet)
			return ErrFailed
		}
		log.Printf("feed : Work : Search Places : Result\n%+v", places)

		// Store each individual place in the database.
		log.Printf("feed : Work : Store : Adding %d Places", len(places))
		for _, place := range places {
			if err := db.Store.Place(ctx, cityID, place); err != nil {
				log.Printf("feed : Work : Store Place : ERROR : %v : %+v", place, err)
				continue
			}
			log.Printf("feed : Work : Store Place : Success : %+v", place)
		}

		// If this was the last result, we are done.
		if errRet == io.EOF {
			break
		}
	}

	return nil
}
