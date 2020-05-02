package data

import (
	"context"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

// Error variables to indicate entities exsit.
var (
	ErrCityExists  = errors.New("city exists")
	ErrPlaceExists = errors.New("place exists")
)

type mutate struct {
	query   query
	graphql *graphql.GraphQL
}

// AddCity first checks to validate the specified city doesn't exists in
// the database. If it doesn't, then the city is added to the database.
// It will return a new City with the city ID from the database.
func (m *mutate) AddCity(ctx context.Context, city City) (City, error) {
	if mutCity.exists(ctx, m.query, city) {
		return City{}, ErrCityExists
	}

	city, err := mutCity.add(ctx, m.graphql, city)
	if err != nil {
		return City{}, errors.New("adding city to database")
	}

	return city, nil
}

// AddPlace adds a new place to the database and connects it to the specified
// city. If the place already exists (by name), the function will return
// an error ErrPlaceExists.
func (m *mutate) AddPlace(ctx context.Context, cityID string, place Place) (Place, error) {
	if mutPlace.exists(ctx, m.query, place) {
		return Place{}, ErrPlaceExists
	}

	place, err := mutPlace.add(ctx, m.graphql, place)
	if err != nil {
		return Place{}, errors.New("adding place to database")
	}

	if err := mutPlace.updateCity(ctx, m.graphql, cityID, place); err != nil {
		return Place{}, errors.Wrap(err, "adding place to city in database")
	}

	return place, nil
}

// ReplaceAdvisory add a new advisory to the database and connects it
// to the specified city.
func (m *mutate) ReplaceAdvisory(ctx context.Context, cityID string, advisory Advisory) (Advisory, error) {
	if err := mutAdvisory.delete(ctx, m.query, m.graphql, cityID); err != nil {
		if err != ErrAdvisoryNotFound {
			return Advisory{}, errors.Wrap(err, "deleting advisory from database")
		}
	}

	advisory, err := mutAdvisory.add(ctx, m.graphql, advisory)
	if err != nil {
		return Advisory{}, errors.Wrap(err, "adding advisory to database")
	}

	if err := mutAdvisory.updateCity(ctx, m.graphql, cityID, advisory); err != nil {
		return Advisory{}, errors.Wrap(err, "replace advisory in city")
	}

	return advisory, nil
}

// ReplaceWeather add a new weather to the database and connects it
// to the specified city.
func (m *mutate) ReplaceWeather(ctx context.Context, cityID string, weather Weather) (Weather, error) {
	if err := mutWeather.delete(ctx, m.query, m.graphql, cityID); err != nil {
		if err != ErrWeatherNotFound {
			return Weather{}, errors.Wrap(err, "deleting weather from database")
		}
	}

	weather, err := mutWeather.add(ctx, m.graphql, weather)
	if err != nil {
		return Weather{}, errors.Wrap(err, "adding weather to database")
	}

	if err := mutWeather.updateCity(ctx, m.graphql, cityID, weather); err != nil {
		return Weather{}, errors.Wrap(err, "replace weather in city")
	}

	return weather, nil
}
