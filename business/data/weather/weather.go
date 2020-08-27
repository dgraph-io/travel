// Package weather provides support for managing weather data in the database.
package weather

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("weather not found")
)

// Replace replaces a weather in the database and connects it
// to the specified city.
func Replace(ctx context.Context, gql *graphql.GraphQL, weather Weather) (Weather, error) {
	if weather.ID != "" {
		return Weather{}, errors.New("weather contains id")
	}
	if weather.City.ID == "" {
		return Weather{}, errors.New("cityid not provided")
	}

	if err := delete(ctx, gql, weather.City.ID); err != nil {
		if err != ErrNotFound {
			return Weather{}, errors.Wrap(err, "deleting weather from database")
		}
	}

	weather, err := add(ctx, gql, weather)
	if err != nil {
		return Weather{}, errors.Wrap(err, "adding weather to database")
	}

	return weather, nil
}

// QueryByCity returns the specified weather from the database by the city id.
func QueryByCity(ctx context.Context, gql *graphql.GraphQL, cityID string) (Weather, error) {
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

	var result struct {
		GetCity struct {
			Weather Weather `json:"weather"`
		} `json:"getCity"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return Weather{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.Weather.ID == "" {
		return Weather{}, ErrNotFound
	}

	return result.GetCity.Weather, nil
}

// =============================================================================

func add(ctx context.Context, gql *graphql.GraphQL, weather Weather) (Weather, error) {
	mutation, result := prepareAdd(weather)
	if err := gql.Query(ctx, mutation, &result); err != nil {
		return Weather{}, errors.Wrap(err, "failed to add weather")
	}

	if len(result.AddWeather.Weather) != 1 {
		return Weather{}, errors.New("advisory id not returned")
	}

	weather.ID = result.AddWeather.Weather[0].ID
	return weather, nil
}

func delete(ctx context.Context, gql *graphql.GraphQL, cityID string) error {
	weather, err := QueryByCity(ctx, gql, cityID)
	if err != nil {
		return err
	}

	mutation, result := prepareDelete(weather.ID)
	if err := gql.Query(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete weather")
	}

	if result.DeleteWeather.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteWeather.NumUids, result.DeleteWeather.Msg)
		return errors.New(msg)
	}

	return nil
}

// =============================================================================

func prepareAdd(weather Weather) (string, addResult) {
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
}`, weather.City.ID, weather.CityName, weather.Desc, weather.FeelsLike, weather.Humidity,
		weather.Pressure, weather.Sunrise, weather.Sunset, weather.Temp,
		weather.MinTemp, weather.MaxTemp, weather.Visibility, weather.WindDirection,
		weather.WindSpeed, result.document())

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
