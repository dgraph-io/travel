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

type city struct {
	CountryCode string
	Name        string
	Lat         float64
	Lng         float64
}

// These are the currently cities supported.
var cities = map[string]city{
	"miami":   {"US", "miami", 25.7617, -80.1918},
	"newyork": {"US", "new york", 40.730610, -73.935242},
	"sydney":  {"AU", "sydney", -33.865143, 151.209900},
}

// ErrFailed is used to report the program failed back to main
// so the correct error code is returned.
var ErrFailed = errors.New("feed failed")

// Search represents a city and its coordinates. All fields must be
// populated for a Search to be successful.
type Search struct {
	CityName   string
	Categories []string
	Radius     uint
}

// Keys represents the set of keys needed for the different API's
// that are used to retrieve data.
type Keys struct {
	MapKey     string
	WeatherKey string
}

// URL represents the set of url's needed for the different API's
// that are used to retrieve data.
type URL struct {
	Advisory string
	Weather  string
}

// Work retrieves and stores the feed data for this API.
func Work(log *log.Logger, dgraph data.Dgraph, search Search, keys Keys, url URL) error {
	ctx := context.Background()

	searchCity, ok := cities[search.CityName]
	if !ok {
		log.Print("feed: Work: city selection: ERROR: not a suppored city")
		return ErrFailed
	}

	log.Println("feed: Work: Wait for the database is ready ...")
	err := data.Readiness(ctx, dgraph.APIHostInside, 5*time.Second)
	if err != nil {
		log.Printf("feed: Work: Readiness: ERROR: %v", err)
		return ErrFailed
	}

	db, err := data.NewDB(dgraph)
	if err != nil {
		log.Printf("feed: Work: New Data: ERROR: %v", err)
		return ErrFailed
	}

	if err := db.Schema.Create(ctx); err != nil {
		log.Printf("feed: Work: Create Schema: ERROR: %v", err)
		return ErrFailed
	}

	city, err := addCity(ctx, log, db, searchCity.Name, searchCity.Lat, searchCity.Lng)
	if err != nil {
		log.Printf("feed: Work: Add City: ERROR: %v", err)
		return ErrFailed
	}

	if err := replaceWeather(ctx, log, db, keys.WeatherKey, url.Weather, city.ID, city.Lat, city.Lng); err != nil {
		log.Printf("feed: Work: Replace Weather: ERROR: %v", err)
		return ErrFailed
	}

	if err := replaceAdvisory(ctx, log, db, url.Advisory, city.ID, searchCity.CountryCode); err != nil {
		log.Printf("feed: Work: Replace Advisory: ERROR: %v", err)
		return ErrFailed
	}

	if err := addPlaces(ctx, log, db, keys.MapKey, city, search.Categories, search.Radius); err != nil {
		log.Printf("feed: Work: Add Place: ERROR: %v", err)
		return ErrFailed
	}

	return nil
}

// addCity add the specified city into the database.
func addCity(ctx context.Context, log *log.Logger, db *data.DB, name string, lat float64, lng float64) (data.City, error) {
	city := data.City{
		Name: name,
		Lat:  lat,
		Lng:  lng,
	}
	city, err := db.Mutate.AddCity(ctx, city)
	if err != nil && err != data.ErrCityExists {
		return data.City{}, errors.Wrapf(err, "adding city: %s", name)
	}

	if err == data.ErrCityExists {
		log.Printf("feed: Work: City Existed: ID: %s Name: %s Lat: %f Lng: %f", city.ID, name, lat, lng)
	} else {
		log.Printf("feed: Work: Added City: ID: %s Name: %s Lat: %f Lng: %f", city.ID, name, lat, lng)
	}

	return city, nil
}

// replaceWeather pulls weather information and updates it for the specified city.
func replaceWeather(ctx context.Context, log *log.Logger, db *data.DB, apiKey string, url string, cityID string, lat float64, lng float64) error {
	weather, err := weather.Search(ctx, apiKey, url, lat, lng)
	if err != nil {
		return errors.Wrap(err, "searching weather")
	}

	updWeather := marshal.Weather(weather)
	updWeather, err = db.Mutate.ReplaceWeather(ctx, cityID, updWeather)
	if err != nil {
		return errors.Wrap(err, "storing weather")
	}

	log.Printf("feed: Work: Replaced Weather: ID: %s Desc: %s", updWeather.ID, updWeather.Desc)
	return nil
}

// replaceAdvisory pulls advisory information and updates it for the specified city.
func replaceAdvisory(ctx context.Context, log *log.Logger, db *data.DB, url string, cityID string, countryCode string) error {
	advisory, err := advisory.Search(ctx, url, countryCode)
	if err != nil {
		return errors.Wrap(err, "searching advisory")
	}

	updAdvisory := marshal.Advisory(advisory)
	updAdvisory, err = db.Mutate.ReplaceAdvisory(ctx, cityID, updAdvisory)
	if err != nil {
		return errors.Wrap(err, "replacing advisory")
	}

	log.Printf("feed: Work: Replaced Advisory: ID: %s Message: %s", updAdvisory.ID, updAdvisory.Message)
	return nil
}

// addPlaces pulls place information and adds new places to the specified city.
func addPlaces(ctx context.Context, log *log.Logger, db *data.DB, apiKey string, city data.City, categories []string, radius uint) error {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return errors.Wrap(err, "creating map client")
	}

	for _, category := range categories {
		filter := places.Filter{
			Name:    city.Name,
			Lat:     city.Lat,
			Lng:     city.Lng,
			Keyword: category,
			Radius:  radius,
		}
		log.Printf("feed: Work: Search Places: filter: %v]", filter)

		// Only store up to the first 20 places.
		for i := 0; i < 1; i++ {
			places, errRet := places.Search(ctx, client, &filter)
			if errRet != nil && errRet != io.EOF {
				return errors.Wrap(err, "searching places")
			}

			for _, place := range places {
				place, err := db.Mutate.AddPlace(ctx, marshal.Place(place, city.ID, category))
				if err != nil && err != data.ErrPlaceExists {
					return errors.Wrapf(err, "adding place: %s", place.Name)
				}

				if err == data.ErrPlaceExists {
					log.Printf("feed: Work: Place Existed: ID: %s Name: %s", place.ID, place.Name)
				} else {
					log.Printf("feed: Work: Added Place: ID: %s Name: %s", place.ID, place.Name)
				}
			}

			if errRet == io.EOF {
				break
			}
		}
	}

	return nil
}
