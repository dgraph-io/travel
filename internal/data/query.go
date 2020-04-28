package data

import (
	"context"
	"fmt"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

// ErrCityNotFound is returned when a city is not found.
var ErrCityNotFound = errors.New("city not found")

type query struct {
	graphql *graphql.GraphQL
}

// City returns the specified city from the database by the city id.
func (q *query) City(ctx context.Context, cityID string) (City, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		id
		name
		lat
		lng
	}
}`, cityID)

	var result struct {
		GetCity struct {
			City
		} `json:"getCity"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return City{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.City.ID == "" {
		return City{}, ErrCityNotFound
	}

	return result.GetCity.City, nil
}

// CityByName returns the specified city from the database by the city name.
func (q *query) CityByName(ctx context.Context, name string) (City, error) {
	query := fmt.Sprintf(`
query {
	queryCity(filter: {	name: {	eq: %q } }) {
		id
		name
		lat
		lng
	}
}`, name)

	var result struct {
		QueryCity []struct {
			City
		} `json:"queryCity"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return City{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryCity) != 1 {
		return City{}, ErrCityNotFound
	}

	return result.QueryCity[0].City, nil
}

// Advisory returns the specified advisory from the database by the city id.
func (q *query) Advisory(ctx context.Context, cityID string) (Advisory, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		advisory {
			continent
			country
			country_code
			last_updated
			message
			score
			source
		}
	}
}`, cityID)

	var result struct {
		GetCity struct {
			Advisory Advisory `json:"advisory"`
		} `json:"getCity"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return Advisory{}, errors.Wrap(err, "query failed")
	}

	return result.GetCity.Advisory, nil
}

// Weather returns the specified weather from the database by the city id.
func (q *query) Weather(ctx context.Context, cityID string) (Weather, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		weather {
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

	return result.GetCity.Weather, nil
}

// Places returns the collection of palces from the database by the city id.
func (q *query) Places(ctx context.Context, cityID string) ([]Place, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		places {
			address,
			avg_user_rating,
			city_name,
			gmaps_url,
			lat,
			lng,
			location_type,
			name,
			no_user_rating,
			place_id,
			photo_id
		}
	}
}`, cityID)

	var result struct {
		GetCity struct {
			Places []Place `json:"places"`
		} `json:"getCity"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	return result.GetCity.Places, nil
}
