package data

import (
	"context"
	"fmt"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type _mutWeather struct{}

var mutWeather _mutWeather

func (_mutWeather) add(ctx context.Context, graphql *graphql.GraphQL, weather Weather) (Weather, error) {
	if weather.ID != "" {
		return Weather{}, errors.New("weather contains id")
	}

	var result struct {
		AddWeather struct {
			Weather []struct {
				ID string `json:"id"`
			} `json:"weather"`
		} `json:"addWeather"`
	}

	if err := graphql.Mutate(ctx, mutWeather.marshalAdd(weather), &result); err != nil {
		return Weather{}, errors.Wrap(err, "failed to add weather")
	}

	if len(result.AddWeather.Weather) != 1 {
		return Weather{}, errors.New("advisory id not returned")
	}

	weather.ID = result.AddWeather.Weather[0].ID
	return weather, nil
}

func (_mutWeather) updateCity(ctx context.Context, graphql *graphql.GraphQL, cityID string, weather Weather) error {
	if weather.ID == "" {
		return errors.New("weather missing id")
	}

	err := graphql.Mutate(ctx, mutWeather.marshalUpdCity(cityID, weather), nil)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

func (_mutWeather) delete(ctx context.Context, query query, graphql *graphql.GraphQL, cityID string) error {
	weather, err := query.Weather(ctx, cityID)
	if err != nil {
		return err
	}

	var result struct {
		DeleteWeather struct {
			Msg     string
			NumUids int
		} `json:"deleteWeather"`
	}
	if err := graphql.Mutate(ctx, mutWeather.marshalDelete(weather.ID), &result); err != nil {
		return errors.Wrap(err, "failed to delete weather")
	}

	if result.DeleteWeather.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteWeather.NumUids, result.DeleteWeather.Msg)
		return errors.New(msg)
	}

	return nil
}

func (_mutWeather) marshalAdd(weather Weather) string {
	return fmt.Sprintf(`
mutation {
	addWeather(input: [{
		city_name: %q,
		description: %q,
		feels_like: %f,
		humidity: %d,
		pressure: %d,
		sunrise: %d,
		sunset: %d,
		temp: %f,
		temp_min: %f,
		temp_max: %f,
		visibility: %q,
		wind_direction: %d,
		wind_speed: %f
	}])
	{
		weather {
			id
		}
	}
}`, weather.CityName, weather.Desc, weather.FeelsLike, weather.Humidity,
		weather.Pressure, weather.Sunrise, weather.Sunset, weather.Temp,
		weather.MinTemp, weather.MaxTemp, weather.Visibility, weather.WindDirection,
		weather.WindSpeed)
}

func (_mutWeather) marshalUpdCity(cityID string, weather Weather) string {
	mutation := fmt.Sprintf(`
mutation {
	updateCity(input: {
		filter: {
		  id: [%q]
		},
		set: {
			weather: {
				id: %q,
				city_name: %q,
				description: %q,
				feels_like: %f,
				humidity: %d,
				pressure: %d,
				sunrise: %d,
				sunset: %d,
				temp: %f,
				temp_min: %f,
				temp_max: %f,
				visibility: %q,
				wind_direction: %d,
				wind_speed: %f
			}
		}
	})
	{
		city {
			id
		}
	}
}`, cityID, weather.ID, weather.CityName, weather.Desc, weather.FeelsLike, weather.Humidity,
		weather.Pressure, weather.Sunrise, weather.Sunset, weather.Temp,
		weather.MinTemp, weather.MaxTemp, weather.Visibility, weather.WindDirection,
		weather.WindSpeed)

	return mutation
}

func (_mutWeather) marshalDelete(weatherID string) string {
	return fmt.Sprintf(`
mutation {
	deleteWeather(filter: { id: [%q] })
	{
		msg,
		numUids,
	}
}`, weatherID)
}
