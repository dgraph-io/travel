package data

import (
	"context"
	"fmt"

	"github.com/dgraph-io/travel/internal/advisory"
	"github.com/dgraph-io/travel/internal/places"
	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/dgraph-io/travel/internal/weather"
	"github.com/pkg/errors"
)

type query struct {
	graphql *graphql.GraphQL
}

// City returns the specified city from the database by the city id.
func (q *query) City(ctx context.Context, cityID string) (places.City, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		name
		lat
		lng
	}
}`, cityID)

	var result struct {
		GetCity struct {
			places.City
		} `json:"getCity"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return places.City{}, errors.Wrap(err, "query failed")
	}

	return result.GetCity.City, nil
}

// Advisory returns the specified advisory from the database by the city id.
func (q *query) Advisory(ctx context.Context, cityID string) (advisory.Advisory, error) {
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
			Advisory advisory.Advisory `json:"advisory"`
		} `json:"getCity"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return advisory.Advisory{}, errors.Wrap(err, "query failed")
	}

	return result.GetCity.Advisory, nil
}

// Weather returns the specified weather from the database by the city id.
func (q *query) Weather(ctx context.Context, cityID string) (weather.Weather, error) {
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
			Weather weather.Weather `json:"weather"`
		} `json:"getCity"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return weather.Weather{}, errors.Wrap(err, "query failed")
	}

	return result.GetCity.Weather, nil
}

// Places returns the collection of palces from the database by the city id.
func (q *query) Places(ctx context.Context, cityID string) ([]places.Place, error) {
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
			Places []places.Place `json:"places"`
		} `json:"getCity"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	return result.GetCity.Places, nil
}
