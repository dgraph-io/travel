package data

import (
	"context"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrCityExists  = errors.New("city exists")
	ErrPlaceExists = errors.New("place exists")
)

type mutate struct {
	query   query
	graphql *graphql.GraphQL
}

// DropAll will remove all the data from the database.
func (m *mutate) DropAll(ctx context.Context) error {
	var response struct {
		Code    string
		Message string
	}
	if err := m.graphql.DropAll(ctx, &response); err != nil {
		return errors.Wrap(err, "drop failed")
	}
	if response.Code != "Success" {
		return errors.New(response.Message)
	}

	return nil
}

// AddCity add a new city to the database. If the city already exists
// this function will fail but the found city is returned. If the city is
// being added, the city with the id from the database is returned.
func (m *mutate) AddCity(ctx context.Context, city City) (City, error) {
	if city, err := m.query.CityByName(ctx, city.Name); err == nil {
		return city, ErrCityExists
	}

	city, err := mutCity.add(ctx, m.graphql, city)
	if err != nil {
		return City{}, errors.New("adding city to database")
	}

	return city, nil
}

// AddPlace add a new place to the database. If the place already exists
// this function will fail but the found place is returned. If the city is
// being added, the city with the id from the database is returned.
func (m *mutate) AddPlace(ctx context.Context, cityID string, place Place) (Place, error) {
	if place, err := m.query.PlaceByName(ctx, place.Name); err == nil {
		return place, ErrPlaceExists
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
