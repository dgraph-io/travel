// Package loader provides support for update new and old city information.
package loader

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/advisory"
	"github.com/dgraph-io/travel/business/data/city"
	"github.com/dgraph-io/travel/business/data/place"
	"github.com/dgraph-io/travel/business/data/ready"
	"github.com/dgraph-io/travel/business/data/schema"
	"github.com/dgraph-io/travel/business/data/weather"
	advisoryfeed "github.com/dgraph-io/travel/business/feeds/advisory"
	placesfeed "github.com/dgraph-io/travel/business/feeds/places"
	weatherfeed "github.com/dgraph-io/travel/business/feeds/weather"
	"github.com/pkg/errors"
	"googlemaps.github.io/maps"
)

// Search represents a city and its coordinates. All fields must be
// populated for a Search to be successful.
type Search struct {
	CityName    string
	CountryCode string
	Lat         float64
	Lng         float64
}

// Config defines the set of mandatory settings.
type Config struct {
	Filter Filter
	Keys   Keys
	URL    URL
}

// Filter represents search related refinements.
type Filter struct {
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

// UpdateSchema creates/updates the schema for the database.
func UpdateSchema(gqlConfig data.GraphQLConfig, schemaConfig schema.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := ready.Validate(ctx, gqlConfig.URL, 5*time.Second)
	if err != nil {
		return errors.Wrapf(err, "waiting for database to be ready")
	}

	gql := data.NewGraphQL(gqlConfig)

	schema, err := schema.New(gql, schemaConfig)
	if err != nil {
		return errors.Wrapf(err, "preparing schema")
	}

	if err := schema.Create(ctx); err != nil {
		return errors.Wrapf(err, "creating schema")
	}

	return nil
}

// UpdateData retrieves and stores the feed data for this API.
func UpdateData(log *log.Logger, gqlConfig data.GraphQLConfig, traceID string, config Config, search Search) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	gql := data.NewGraphQL(gqlConfig)
	loader := newLoader(log, gql)

	cty, err := loader.addCity(ctx, traceID, search.CityName, search.Lat, search.Lng)
	if err != nil {
		return errors.Wrapf(err, "adding city")
	}

	if err := loader.replaceWeather(ctx, traceID, config.Keys.WeatherKey, config.URL.Weather, cty.ID, cty.Lat, cty.Lng); err != nil {
		return errors.Wrapf(err, "replacing weather")
	}

	if err := loader.replaceAdvisory(ctx, traceID, config.URL.Advisory, cty.ID, search.CountryCode); err != nil {
		return errors.Wrapf(err, "replacing advisory")
	}

	if err := loader.addPlaces(ctx, traceID, config.Keys.MapKey, cty, config.Filter.Categories, config.Filter.Radius); err != nil {
		return errors.Wrapf(err, "adding places")
	}

	return nil
}

type loader struct {
	log      *log.Logger
	gql      *graphql.GraphQL
	advisory advisory.Advisory
	city     city.City
	place    place.Place
	weather  weather.Weather
}

func newLoader(log *log.Logger, gql *graphql.GraphQL) loader {
	return loader{
		log:      log,
		gql:      gql,
		advisory: advisory.New(log, gql),
		city:     city.New(log, gql),
		place:    place.New(log, gql),
		weather:  weather.New(log, gql),
	}
}

// addCity add the specified city into the database.
func (l loader) addCity(ctx context.Context, traceID string, name string, lat float64, lng float64) (city.Info, error) {
	newCity := city.Info{
		Name: name,
		Lat:  lat,
		Lng:  lng,
	}
	newCity, err := l.city.Add(ctx, traceID, newCity)
	if err != nil && err != city.ErrExists {
		return city.Info{}, errors.Wrapf(err, "adding city: %s", name)
	}

	if err == city.ErrExists {
		log.Printf("feed: Work: City Existed: ID: %s Name: %s Lat: %f Lng: %f", newCity.ID, name, lat, lng)
	} else {
		log.Printf("feed: Work: Added City: ID: %s Name: %s Lat: %f Lng: %f", newCity.ID, name, lat, lng)
	}

	return newCity, nil
}

// replaceWeather pulls weather information and updates it for the specified city.
func (l loader) replaceWeather(ctx context.Context, traceID string, apiKey string, url string, cityID string, lat float64, lng float64) error {
	feedData, err := weatherfeed.Search(ctx, apiKey, url, lat, lng)
	if err != nil {
		return errors.Wrap(err, "searching weather")
	}

	newWeather := marshalWeather(feedData, cityID)
	newWeather, err = l.weather.Replace(ctx, traceID, newWeather)
	if err != nil {
		return errors.Wrap(err, "storing weather")
	}

	log.Printf("feed: Work: Replaced Weather: ID: %s Desc: %s", newWeather.ID, newWeather.Desc)
	return nil
}

// replaceAdvisory pulls advisory information and updates it for the specified city.
func (l loader) replaceAdvisory(ctx context.Context, traceID string, url string, cityID string, countryCode string) error {
	feedData, err := advisoryfeed.Search(ctx, url, countryCode)
	if err != nil {
		return errors.Wrap(err, "searching advisory")
	}

	newAdvisory := marshalAdvisory(feedData, cityID)
	newAdvisory, err = l.advisory.Replace(ctx, traceID, newAdvisory)
	if err != nil {
		return errors.Wrap(err, "replacing advisory")
	}

	log.Printf("feed: Work: Replaced Advisory: ID: %s Message: %s", newAdvisory.ID, newAdvisory.Message)
	return nil
}

// addPlaces pulls place information and adds new places to the specified city.
func (l loader) addPlaces(ctx context.Context, traceID string, apiKey string, cty city.Info, categories []string, radius uint) error {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return errors.Wrap(err, "creating map client")
	}

	for _, category := range categories {
		filter := placesfeed.Filter{
			Name:    cty.Name,
			Lat:     cty.Lat,
			Lng:     cty.Lng,
			Keyword: category,
			Radius:  radius,
		}
		log.Printf("feed: Work: Search Places: filter: %v]", filter)

		// Only store up to the first 20 places.
		for i := 0; i < 1; i++ {
			feedList, errRet := placesfeed.Search(ctx, client, &filter)
			if errRet != nil && errRet != io.EOF {
				return errors.Wrap(err, "searching places")
			}

			for _, feedData := range feedList {
				newPlace, err := l.place.Add(ctx, traceID, marshalPlace(feedData, cty.ID, category))
				if err != nil && err != place.ErrExists {
					return errors.Wrapf(err, "adding place: %s", newPlace.Name)
				}

				if err == place.ErrExists {
					log.Printf("feed: Work: Place Existed: ID: %s Name: %s", newPlace.ID, newPlace.Name)
				} else {
					log.Printf("feed: Work: Added Place: ID: %s Name: %s", newPlace.ID, newPlace.Name)
				}
			}

			if errRet == io.EOF {
				break
			}
		}
	}

	return nil
}
