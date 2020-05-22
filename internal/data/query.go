package data

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// Not found errors.
var (
	ErrCityNotFound     = errors.New("city not found")
	ErrPlaceNotFound    = errors.New("place not found")
	ErrAdvisoryNotFound = errors.New("advisory not found")
	ErrWeatherNotFound  = errors.New("weather not found")
)

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
		GetCity City `json:"getCity"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return City{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.ID == "" {
		return City{}, ErrCityNotFound
	}

	return result.GetCity, nil
}

// CityByName returns the specified city from the database by the city name.
func (q *query) CityByName(ctx context.Context, name string) (City, error) {
	query := fmt.Sprintf(`
query {
	queryCity(filter: { name: { eq: %q } }) {
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
			id
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

	if result.GetCity.Advisory.ID == "" {
		return Advisory{}, ErrAdvisoryNotFound
	}

	return result.GetCity.Advisory, nil
}

// Weather returns the specified weather from the database by the city id.
func (q *query) Weather(ctx context.Context, cityID string) (Weather, error) {
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

// Place returns the collection of places from the database by the place id.
func (q *query) Place(ctx context.Context, placeID string) (Place, error) {
	query := fmt.Sprintf(`
query {
	getPlace(id: %q) {
		id
		address
		avg_user_rating
		category
		city {
			id
		}
		city_name
		gmaps_url
		lat
		lng
		location_type
		name
		no_user_rating
		place_id
		photo_id
	}
}`, placeID)

	var result struct {
		GetPlace struct {
			Place
		} `json:"getPlace"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return Place{}, errors.Wrap(err, "query failed")
	}

	if result.GetPlace.Place.ID == "" {
		return Place{}, ErrPlaceNotFound
	}

	return result.GetPlace.Place, nil
}

// PlaceByName returns the collection of places from the database
// by the place name.
func (q *query) PlaceByName(ctx context.Context, name string) (Place, error) {
	query := fmt.Sprintf(`
query {
	queryPlace(filter: { name: { eq: %q } }) {
		id
		address
		avg_user_rating
		category
		city {
			id
		}
		city_name
		gmaps_url
		lat
		lng
		location_type
		name
		no_user_rating
		place_id
		photo_id
	}
}`, name)

	var result struct {
		QueryPlace []Place `json:"queryPlace"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return Place{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return Place{}, ErrPlaceNotFound
	}

	return result.QueryPlace[0], nil
}

// PlaceByCategory returns the collection of places from the database
// by the cagtegory name.
func (q *query) PlaceByCategory(ctx context.Context, category string) ([]Place, error) {
	query := fmt.Sprintf(`
query {
	queryPlace(filter: { category: { eq: %q } }) {
		id
		address
		avg_user_rating
		category
		city {
			id
		}
		city_name
		gmaps_url
		lat
		lng
		location_type
		name
		no_user_rating
		place_id
		photo_id
	}
}`, category)

	var result struct {
		QueryPlace []Place `json:"queryPlace"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return nil, ErrPlaceNotFound
	}

	return result.QueryPlace, nil
}

// Places returns the collection of places from the database by the city id.
func (q *query) Places(ctx context.Context, cityID string) ([]Place, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		places {
			id
			address
			avg_user_rating
			category
			city {
				id
			}
			city_name
			gmaps_url
			lat
			lng
			location_type
			name
			no_user_rating
			place_id
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
