package feed

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/feeds/advisory"
	"github.com/dgraph-io/travel/internal/feeds/places"
	"github.com/dgraph-io/travel/internal/feeds/weather"
	"github.com/pkg/errors"
	"googlemaps.github.io/maps"
)

// ErrFailed is used to report the program failed back to main
// so the correct error code is returned.
var ErrFailed = errors.New("feed failed")

// Dgraph represents the IP and Ports we need to talk to the
// server for the different functions we need to perform.
type Dgraph struct {
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

	log.Println("feed : Work : Wait for the database is ready ...")
	err := data.Readiness(ctx, dgraph.APIHost, 5*time.Second)
	if err != nil {
		log.Printf("feed : Work : Readiness : ERROR : %+v", err)
		return ErrFailed
	}

	db, err := data.NewDB(dgraph.APIHost)
	if err != nil {
		log.Printf("feed : Work : New Data : ERROR : %+v", err)
		return ErrFailed
	}

	if err := db.Schema.Create(ctx); err != nil {
		log.Printf("feed : Work : Create Schema : ERROR : %+v", err)
		return ErrFailed
	}

	city, err := storeCity(ctx, log, db, search.CityName, search.Lat, search.Lng)
	if err != nil {
		log.Printf("feed : Work : Store City : ERROR : %+v", err)
		return ErrFailed
	}

	if err := storeWeather(ctx, log, db, city.ID, keys.WeatherKey, search.Lat, search.Lng); err != nil {
		log.Printf("feed : Work : Store Weather : ERROR : %+v", err)
		return ErrFailed
	}

	if err := storeAdvisory(ctx, log, db, city.ID, search.CountryCode); err != nil {
		log.Printf("feed : Work : Store Advisory : ERROR : %+v", err)
		return ErrFailed
	}

	if err := storePlaces(ctx, log, db, city, keys.MapKey, search.Keyword, search.Radius); err != nil {
		log.Printf("feed : Work : Store Advisory : ERROR : %+v", err)
		return ErrFailed
	}

	return nil
}

// storeCity add the specified city into the database.
func storeCity(ctx context.Context, log *log.Logger, db *data.DB, name string, lat float64, lng float64) (data.City, error) {
	city := data.City{
		Name: name,
		Lat:  lat,
		Lng:  lng,
	}
	city, err := db.Store.City(ctx, city)
	if err != nil {
		return data.City{}, errors.Wrap(err, "storing city")
	}

	log.Printf("feed : Work : Location : CityID[%s] Name[%s] Lat[%f] Lng[%f]", city.ID, name, lat, lng)

	return city, nil
}

// storeWeather pulls weather information and stores it for the specified city.
func storeWeather(ctx context.Context, log *log.Logger, db *data.DB, cityID string, apiKey string, lat float64, lng float64) error {
	weather, err := weather.Search(ctx, apiKey, lat, lng)
	if err != nil {
		return errors.Wrap(err, "searching weather")
	}

	log.Printf("feed : Work : Search Weather : Result : %+v", weather)

	if err := db.Store.Weather(ctx, cityID, marshal.Weather(weather)); err != nil {
		return errors.Wrap(err, "storing weather")
	}

	return nil
}

// storeAdvisory pulls advisory information and stores it for the specified city.
func storeAdvisory(ctx context.Context, log *log.Logger, db *data.DB, cityID string, countryCode string) error {
	advisory, err := advisory.Search(ctx, countryCode)
	if err != nil {
		return errors.Wrap(err, "searching advisory")
	}

	log.Printf("feed : Work : Search Advisory : Result : %+v", advisory)

	if err := db.Store.Advisory(ctx, cityID, marshal.Advisory(advisory)); err != nil {
		return errors.Wrap(err, "storing advisory")
	}

	return nil
}

// storePlaces pulls place information and stores it for the specified city.
func storePlaces(ctx context.Context, log *log.Logger, db *data.DB, city data.City, apiKey string, keyword string, radius uint) error {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return errors.Wrap(err, "creating map client")
	}

	filter := places.Filter{
		Name:    city.Name,
		Lat:     city.Lat,
		Lng:     city.Lng,
		Keyword: keyword,
		Radius:  radius,
	}
	log.Printf("feed : Work : Search Places : filter[%+v]", filter)

	// For now we will test with 40 places.
	for i := 0; i < 2; i++ {

		places, errRet := places.Search(ctx, client, filter)
		if errRet != nil && errRet != io.EOF {
			return errors.Wrap(err, "searching places")
		}

		for _, place := range places {
			log.Printf("feed : Work : Search Places : %s", place.Name)
		}

		if err := db.Store.Places(ctx, city.ID, marshal.Places(places)); err != nil {
			return errors.Wrap(err, "storing places")
		}

		if errRet == io.EOF {
			break
		}
	}

	return nil
}
