package data

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// Weather returns the specified weather from the database by the city id.
func (q query) Weather(ctx context.Context, cityID string) (Weather, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		weather {
			id
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
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return Weather{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.Weather.ID == "" {
		return Weather{}, ErrWeatherNotFound
	}

	return result.GetCity.Weather, nil
}
