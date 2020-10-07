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

// Weather manages the set of API's for city access.
type Weather struct {
	log *log.Logger
	gql *graphql.GraphQL
}

// New constructs a Weather for api access.
func New(log *log.Logger, gql *graphql.GraphQL) Weather {
	return Weather{
		log: log,
		gql: gql,
	}
}

// Replace replaces a weather in the database and connects it
// to the specified city.
func (w Weather) Replace(ctx context.Context, traceID string, wth Info) (Info, error) {
	if wth.ID != "" {
		return Info{}, errors.New("weather contains id")
	}
	if wth.City.ID == "" {
		return Info{}, errors.New("cityid not provided")
	}

	if err := w.delete(ctx, traceID, wth.City.ID); err != nil {
		if err != ErrNotFound {
			return Info{}, errors.Wrap(err, "deleting weather from database")
		}
	}

	wth, err := w.add(ctx, traceID, wth)
	if err != nil {
		return Info{}, errors.Wrap(err, "adding weather to database")
	}

	return wth, nil
}

// QueryByCity returns the specified weather from the database by the city id.
func (w Weather) QueryByCity(ctx context.Context, traceID string, cityID string) (Info, error) {
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

	w.log.Printf("%s: %s: %s", traceID, "weather.QueryByID", data.Log(query))

	var result struct {
		GetCity struct {
			Weather Info `json:"weather"`
		} `json:"getCity"`
	}
	if err := w.gql.Query(ctx, query, &result); err != nil {
		return Info{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.Weather.ID == "" {
		return Info{}, ErrNotFound
	}

	return result.GetCity.Weather, nil
}

// =============================================================================

func (w Weather) add(ctx context.Context, traceID string, wth Info) (Info, error) {
	mutation, result := prepareAdd(wth)
	w.log.Printf("%s: %s: %s", traceID, "weather.Add", data.Log(mutation))

	if err := w.gql.Query(ctx, mutation, &result); err != nil {
		return Info{}, errors.Wrap(err, "failed to add weather")
	}

	if len(result.AddWeather.Weather) != 1 {
		return Info{}, errors.New("advisory id not returned")
	}

	wth.ID = result.AddWeather.Weather[0].ID
	return wth, nil
}

func (w Weather) delete(ctx context.Context, traceID string, cityID string) error {
	wth, err := w.QueryByCity(ctx, traceID, cityID)
	if err != nil {
		return err
	}

	mutation, result := prepareDelete(wth.ID)
	w.log.Printf("%s: %s: %s", traceID, "weather.Delete", data.Log(mutation))

	if err := w.gql.Query(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete weather")
	}

	if result.DeleteWeather.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteWeather.NumUids, result.DeleteWeather.Msg)
		return errors.New(msg)
	}

	return nil
}

// =============================================================================

func prepareAdd(wth Info) (string, addResult) {
	var result addResult
	mutation := fmt.Sprintf(`
mutation {
	addWeather(input: [{
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

	return mutation, result
}

func prepareDelete(weatherID string) (string, deleteResult) {
	var result deleteResult
	mutation := fmt.Sprintf(`
mutation {
	deleteWeather(filter: { id: [%q] })
	%s
}`, weatherID, result.document())

	return mutation, result
}
