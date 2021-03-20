// Package weather provides support for managing weather data in the database.
package weather

import (
	"context"
	"fmt"
	"log"

	"github.com/ardanlabs/graphql"
	"github.com/dgraph-io/travel/business/data"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("weather not found")
)

// Store manages the set of API's for city access.
type Store struct {
	log *log.Logger
	gql *graphql.GraphQL
}

// NewStore constructs a weather store for api access.
func NewStore(log *log.Logger, gql *graphql.GraphQL) Store {
	return Store{
		log: log,
		gql: gql,
	}
}

// Replace replaces a weather in the database and connects it
// to the specified city.
func (s Store) Replace(ctx context.Context, traceID string, wth Weather) (Weather, error) {
	if wth.ID != "" {
		return Weather{}, errors.New("weather contains id")
	}
	if wth.City.ID == "" {
		return Weather{}, errors.New("cityid not provided")
	}

	if oldWth, err := s.QueryByCity(ctx, traceID, wth.City.ID); err == nil {
		if err := s.delete(ctx, traceID, oldWth.ID); err != nil {
			if err != ErrNotFound {
				return Weather{}, errors.Wrap(err, "deleting weather from database")
			}
		}
	}

	return s.add(ctx, traceID, wth)
}

// QueryByCity returns the specified weather from the database by the city id.
func (s Store) QueryByCity(ctx context.Context, traceID string, cityID string) (Weather, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		weather {
			id
			city {
				id
			}
			city_name
			description
			feels_like
			humidity
			pressure
			sunrise
			sunset
			temp
			temp_min
			temp_max
			visibility
			wind_direction
			wind_speed
		}
	}
}`, cityID)

	s.log.Printf("%s: %s: %s", traceID, "weather.QueryByID", data.Log(query))

	var result struct {
		GetCity struct {
			Weather Weather `json:"weather"`
		} `json:"getCity"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return Weather{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.Weather.ID == "" {
		return Weather{}, ErrNotFound
	}

	return result.GetCity.Weather, nil
}

// =============================================================================

func (s Store) delete(ctx context.Context, traceID string, wthID string) error {
	var result result
	mutation := fmt.Sprintf(`
	mutation {
		resp: deleteWeather(filter: { id: [%q] })
		%s
	}`, wthID, result.document())

	s.log.Printf("%s: %s: %s", traceID, "weather.Delete", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete weather")
	}

	if result.Resp.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.Resp.NumUids, result.Resp.Msg)
		return errors.New(msg)
	}

	return nil
}

func (s Store) add(ctx context.Context, traceID string, wth Weather) (Weather, error) {
	var result id
	mutation := fmt.Sprintf(`
	mutation {
		resp: addWeather(input: [{
			city: {
				id: %q
			}
			city_name: %q
			description: %q
			feels_like: %f
			humidity: %d
			pressure: %d
			sunrise: %d
			sunset: %d
			temp: %f
			temp_min: %f
			temp_max: %f
			visibility: %q
			wind_direction: %d
			wind_speed: %f
		}])
		%s
	}`, wth.City.ID, wth.CityName, wth.Desc, wth.FeelsLike, wth.Humidity,
		wth.Pressure, wth.Sunrise, wth.Sunset, wth.Temp,
		wth.MinTemp, wth.MaxTemp, wth.Visibility, wth.WindDirection,
		wth.WindSpeed, result.document())

	s.log.Printf("%s: %s: %s", traceID, "weather.Add", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, &result); err != nil {
		return Weather{}, errors.Wrap(err, "failed to add weather")
	}

	if len(result.Resp.Entities) != 1 {
		return Weather{}, errors.New("advisory id not returned")
	}

	wth.ID = result.Resp.Entities[0].ID
	return wth, nil
}
