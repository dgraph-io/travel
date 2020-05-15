package data

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

type mutateWeather struct {
	marshal weatherMarshal
}

var mutWeather mutateWeather

func (mutateWeather) add(ctx context.Context, graphql *graphql.GraphQL, weather Weather) (Weather, error) {
	if weather.ID != "" {
		return Weather{}, errors.New("weather contains id")
	}

	mutation, result := mutWeather.marshal.add(weather)
	if err := graphql.Mutate(ctx, mutation, &result); err != nil {
		return Weather{}, errors.Wrap(err, "failed to add weather")
	}

	if len(result.AddWeather.Weather) != 1 {
		return Weather{}, errors.New("advisory id not returned")
	}

	weather.ID = result.AddWeather.Weather[0].ID
	return weather, nil
}

func (mutateWeather) updateCity(ctx context.Context, graphql *graphql.GraphQL, cityID string, weather Weather) error {
	if weather.ID == "" {
		return errors.New("weather missing id")
	}

	mutation, result := mutWeather.marshal.updCity(cityID, weather)
	err := graphql.Mutate(ctx, mutation, &result)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

func (mutateWeather) delete(ctx context.Context, query query, graphql *graphql.GraphQL, cityID string) error {
	weather, err := query.Weather(ctx, cityID)
	if err != nil {
		return err
	}

	mutation, result := mutWeather.marshal.delete(weather.ID)
	if err := graphql.Mutate(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete weather")
	}

	if result.DeleteWeather.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteWeather.NumUids, result.DeleteWeather.Msg)
		return errors.New(msg)
	}

	return nil
}

type weatherMarshal struct{}

func (weatherMarshal) add(weather Weather) (string, weatherIDResult) {
	var result weatherIDResult
	mutation := fmt.Sprintf(`
mutation {
	addWeather(input: [{
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
}`, weather.CityName, weather.Desc, weather.FeelsLike, weather.Humidity,
		weather.Pressure, weather.Sunrise, weather.Sunset, weather.Temp,
		weather.MinTemp, weather.MaxTemp, weather.Visibility, weather.WindDirection,
		weather.WindSpeed, result.graphql())

	return mutation, result
}

func (weatherMarshal) updCity(cityID string, weather Weather) (string, cityIDResult) {
	var result cityIDResult
	mutation := fmt.Sprintf(`
mutation {
	updateCity(input: {
		filter: {
		  id: [%q]
		},
		set: {
			weather: {
				id: %q
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
			}
		}
	})
	%s
}`, cityID, weather.ID, weather.CityName, weather.Desc, weather.FeelsLike, weather.Humidity,
		weather.Pressure, weather.Sunrise, weather.Sunset, weather.Temp,
		weather.MinTemp, weather.MaxTemp, weather.Visibility, weather.WindDirection,
		weather.WindSpeed, result.marshal())

	return mutation, result
}

func (weatherMarshal) delete(weatherID string) (string, deleteWeatherResult) {
	var result deleteWeatherResult
	mutation := fmt.Sprintf(`
mutation {
	deleteWeather(filter: { id: [%q] })
	%s
}`, weatherID, result.graphql())

	return mutation, result
}

type weatherIDResult struct {
	AddWeather struct {
		Weather []struct {
			ID string `json:"id"`
		} `json:"weather"`
	} `json:"addWeather"`
}

func (weatherIDResult) graphql() string {
	return `{
		weather {
			id
		}
	}`
}

type deleteWeatherResult struct {
	DeleteWeather struct {
		Msg     string
		NumUids int
	} `json:"deleteWeather"`
}

func (deleteWeatherResult) graphql() string {
	return `{
		msg,
		numUids,
	}`
}
