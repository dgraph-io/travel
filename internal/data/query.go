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
{
	city(func: uid(%s)) {
		city_name
		lat
		lng
	}
}`, cityID)

	var result struct {
		City []places.City
	}
	if err := q.graphql.QueryPM(ctx, query, &result); err != nil {
		return places.City{}, errors.Wrap(err, query)
	}

	if len(result.City) == 0 {
		return places.City{}, errors.Wrap(errors.New("no city found"), query)
	}

	return result.City[0], nil
}

// Advisory returns the specified advisory from the database by the city id.
func (q *query) Advisory(ctx context.Context, cityID string) (advisory.Advisory, error) {
	query := fmt.Sprintf(`
{
	city(func: uid(%s)) {
		advisory {
			country
			country_code
			continent
			advisory_score
			advisory_last_updated
			advisory_message
			source
		}
	}
}`, cityID)

	var result struct {
		City []struct {
			Advisory advisory.Advisory
		}
	}
	if err := q.graphql.QueryPM(ctx, query, &result); err != nil {
		return advisory.Advisory{}, errors.Wrap(err, query)
	}

	if len(result.City) == 0 {
		return advisory.Advisory{}, errors.Wrap(errors.New("no advisory found"), query)
	}

	return result.City[0].Advisory, nil
}

// Weather returns the specified weather from the database by the city id.
func (q *query) Weather(ctx context.Context, cityID string) (weather.Weather, error) {
	query := fmt.Sprintf(`
{
	city(func: uid(%s)) {
		weather {
			weather_id
			city_name
			visibility
			description
			temp
			feels_like
			temp_min
			temp_max
			pressure
			humidity
			wind_speed
			wind_direction
			sunrise
			sunset
		}
	}
}`, cityID)

	var result struct {
		City []struct {
			Weather weather.Weather
		}
	}
	if err := q.graphql.QueryPM(ctx, query, &result); err != nil {
		return weather.Weather{}, errors.Wrap(err, query)
	}

	if len(result.City) == 0 {
		return weather.Weather{}, errors.Wrap(errors.New("no weather found"), query)
	}

	return result.City[0].Weather, nil
}

// Places returns the collection of palces from the database by the city id.
func (q *query) Places(ctx context.Context, cityID string) ([]places.Place, error) {
	query := fmt.Sprintf(`
{
	city(func: uid(%s)) {
		places {
			place_id
			city_name
			name
			address
			lat
			lng
			location_type
			avg_user_rating
			no_user_rating
			gmaps_url
			photo_id
		}
	}
}`, cityID)

	var result struct {
		City []struct {
			Places []places.Place
		}
	}
	if err := q.graphql.QueryPM(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, query)
	}

	if len(result.City) == 0 {
		return nil, errors.Wrap(errors.New("no places found"), query)
	}

	return result.City[0].Places, nil
}
