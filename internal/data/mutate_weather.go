package data

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// ReplaceWeather replaces a weather in the database and connects it
// to the specified city.
func (m mutate) ReplaceWeather(ctx context.Context, cityID string, wth Weather) (Weather, error) {
	if err := weather.delete(ctx, m.query, m.graphql, cityID); err != nil {
		if err != ErrWeatherNotFound {
			return Weather{}, errors.Wrap(err, "deleting weather from database")
		}
	}

	wth, err := weather.add(ctx, m.graphql, wth)
	if err != nil {
		return Weather{}, errors.Wrap(err, "adding weather to database")
	}

	if err := weather.updateCity(ctx, m.graphql, cityID, wth); err != nil {
		return Weather{}, errors.Wrap(err, "replace weather in city")
	}

	return wth, nil
}

// =============================================================================

type wth struct {
	prepare weatherPrepare
}

var weather wth

func (w wth) add(ctx context.Context, graphql *graphql.GraphQL, weather Weather) (Weather, error) {
	if weather.ID != "" {
		return Weather{}, errors.New("weather contains id")
	}

	mutation, result := w.prepare.add(weather)
	if err := graphql.Query(ctx, mutation, &result); err != nil {
		return Weather{}, errors.Wrap(err, "failed to add weather")
	}

	if len(result.AddWeather.Weather) != 1 {
		return Weather{}, errors.New("advisory id not returned")
	}

	weather.ID = result.AddWeather.Weather[0].ID
	return weather, nil
}

func (w wth) updateCity(ctx context.Context, graphql *graphql.GraphQL, cityID string, weather Weather) error {
	if weather.ID == "" {
		return errors.New("weather missing id")
	}

	mutation, result := w.prepare.updateCity(cityID, weather)
	err := graphql.Query(ctx, mutation, &result)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

func (w wth) delete(ctx context.Context, query query, graphql *graphql.GraphQL, cityID string) error {
	weather, err := query.Weather(ctx, cityID)
	if err != nil {
		return err
	}

	mutation, result := w.prepare.delete(weather.ID)
	if err := graphql.Query(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete weather")
	}

	if result.DeleteWeather.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteWeather.NumUids, result.DeleteWeather.Msg)
		return errors.New(msg)
	}

	return nil
}

// =============================================================================

type weatherPrepare struct{}

func (weatherPrepare) add(weather Weather) (string, weatherAddResult) {
	var result weatherAddResult
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
		weather.WindSpeed, result.document())

	return mutation, result
}

func (weatherPrepare) updateCity(cityID string, weather Weather) (string, cityUpdateResult) {
	var result cityUpdateResult
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
		weather.WindSpeed, result.document())

	return mutation, result
}

func (weatherPrepare) delete(weatherID string) (string, weatherDeleteResult) {
	var result weatherDeleteResult
	mutation := fmt.Sprintf(`
mutation {
	deleteWeather(filter: { id: [%q] })
	%s
}`, weatherID, result.document())

	return mutation, result
}

type weatherAddResult struct {
	AddWeather struct {
		Weather []struct {
			ID string `json:"id"`
		} `json:"weather"`
	} `json:"addWeather"`
}

func (weatherAddResult) document() string {
	return `{
		weather {
			id
		}
	}`
}

type weatherDeleteResult struct {
	DeleteWeather struct {
		Msg     string
		NumUids int
	} `json:"deleteWeather"`
}

func (weatherDeleteResult) document() string {
	return `{
		msg,
		numUids,
	}`
}
