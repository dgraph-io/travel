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

	city, err := addCity(ctx, log, db, search.CityName, search.Lat, search.Lng)
	if err != nil {
		log.Printf("feed : Work : Add City : ERROR : %v", err)
		return ErrFailed
	}

	if err := replaceWeather(ctx, log, db, city.ID, keys.WeatherKey, search.Lat, search.Lng); err != nil {
		log.Printf("feed : Work : Replace Weather : ERROR : %v", err)
		return ErrFailed
	}

	if err := replaceAdvisory(ctx, log, db, city.ID, search.CountryCode); err != nil {
		log.Printf("feed : Work : Replace Advisory : ERROR : %v", err)
		return ErrFailed
	}

	if err := addPlaces(ctx, log, db, city, keys.MapKey, search.Keyword, search.Radius); err != nil {
		log.Printf("feed : Work : Add Place : ERROR : %v", err)
		return ErrFailed
	}

	return nil
}

// addCity add the specified city into the database.
func addCity(ctx context.Context, log *log.Logger, db *data.DB, name string, lat float64, lng float64) (data.City, error) {
	if city, err := db.Query.CityByName(ctx, name); err == nil {
		log.Printf("feed : Work : City Exists : CityID[%s] Name[%s] Lat[%f] Lng[%f]", city.ID, name, lat, lng)
		return city, nil
	}

	city := data.City{
		Name: name,
		Lat:  lat,
		Lng:  lng,
	}
	city, err := db.Mutate.AddCity(ctx, city)
	if err != nil {
		return data.City{}, errors.Wrapf(err, "adding city: %s", name)
	}

	log.Printf("feed : Work : Added City : CityID[%s] Name[%s] Lat[%f] Lng[%f]", city.ID, name, lat, lng)
	return city, nil
}

// replaceWeather pulls weather information and updates it for the specified city.
func replaceWeather(ctx context.Context, log *log.Logger, db *data.DB, cityID string, apiKey string, lat float64, lng float64) error {
	weather, err := weather.Search(ctx, apiKey, lat, lng)
	if err != nil {
		return errors.Wrap(err, "searching weather")
	}

	updWeather := marshal.Weather(weather)
	updWeather, err = db.Mutate.ReplaceWeather(ctx, cityID, updWeather)
	if err != nil {
		return errors.Wrap(err, "storing weather")
	}

	log.Printf("feed : Work : Replaced Weather : %s:%s\n", updWeather.ID, updWeather.Desc)
	return nil
}

// replaceAdvisory pulls advisory information and updates it for the specified city.
func replaceAdvisory(ctx context.Context, log *log.Logger, db *data.DB, cityID string, countryCode string) error {
	advisory, err := advisory.Search(ctx, countryCode)
	if err != nil {
		return errors.Wrap(err, "searching advisory")
	}

	updAdvisory := marshal.Advisory(advisory)
	updAdvisory, err = db.Mutate.ReplaceAdvisory(ctx, cityID, updAdvisory)
	if err != nil {
		return errors.Wrap(err, "replacing advisory")
	}

	log.Printf("feed : Work : Replaced Advisory : %s:%s\n", updAdvisory.ID, updAdvisory.Message)
	return nil
}

// addPlaces pulls place information and adds new places to the specified city.
func addPlaces(ctx context.Context, log *log.Logger, db *data.DB, city data.City, apiKey string, keyword string, radius uint) error {
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

	// Only add up to the first 100 places.
	for i := 0; i < 5; i++ {
		places, errRet := places.Search(ctx, client, &filter)
		if errRet != nil && errRet != io.EOF {
			return errors.Wrap(err, "searching places")
		}

		for _, place := range places {
			if place, err := db.Query.PlaceByName(ctx, place.Name); err == nil {
				log.Printf("feed : Work : Place Exists : PlaceID[%s] Name[%s]\n", place.ID, place.Name)
				continue
			}

			place, err := db.Mutate.AddPlace(ctx, city.ID, marshal.Place(place))
			if err != nil {
				return errors.Wrapf(err, "adding place: %s", place.Name)
			}
			log.Printf("feed : Work : Added Place : PlaceID[%s] Name[%s]\n", place.ID, place.Name)
		}

		if errRet == io.EOF {
			break
		}
	}

	return nil
}
