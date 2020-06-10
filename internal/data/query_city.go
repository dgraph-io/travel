package data

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// City returns the specified city from the database by the city id.
func (q query) City(ctx context.Context, cityID string) (City, error) {
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
func (q query) CityByName(ctx context.Context, name string) (City, error) {
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
