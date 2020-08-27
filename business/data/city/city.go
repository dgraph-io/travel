// Package city provides support for managing city data in the database.
package city

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrExists   = errors.New("city exists")
	ErrNotFound = errors.New("city not found")
)

// Add adds a new city to the database. If the city already exists
// this function will fail but the found city is returned. If the city is
// being added, the city with the id from the database is returned.
func Add(ctx context.Context, gql *graphql.GraphQL, city City) (City, error) {
	if city, err := QueryByName(ctx, gql, city.Name); err == nil {
		return city, ErrExists
	}

	city, err := add(ctx, gql, city)
	if err != nil {
		return City{}, errors.Wrap(err, "adding city to database")
	}

	return city, nil
}

// QueryByID returns the specified city from the database by the city id.
func QueryByID(ctx context.Context, gql *graphql.GraphQL, cityID string) (City, error) {
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
	if err := gql.Query(ctx, query, &result); err != nil {
		return City{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.ID == "" {
		return City{}, ErrNotFound
	}

	return result.GetCity, nil
}

// QueryByName returns the specified city from the database by the city name.
func QueryByName(ctx context.Context, gql *graphql.GraphQL, name string) (City, error) {
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
	if err := gql.Query(ctx, query, &result); err != nil {
		return City{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryCity) != 1 {
		return City{}, ErrNotFound
	}

	return result.QueryCity[0].City, nil
}

// QueryNames returns the list of city names currently loaded in the database.
func QueryNames(ctx context.Context, gql *graphql.GraphQL) ([]string, error) {
	query := `
	query {
		queryCity(filter: { }) {
			name
		}
	}`

	var result struct {
		QueryCity []struct {
			City
		} `json:"queryCity"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if len(result.QueryCity) != 1 {
		return nil, ErrNotFound
	}

	cities := make([]string, len(result.QueryCity))
	for i, city := range result.QueryCity {
		cities[i] = city.Name
	}

	return cities, nil
}

// =============================================================================

func add(ctx context.Context, gql *graphql.GraphQL, city City) (City, error) {
	if city.ID != "" {
		return City{}, errors.New("city contains id")
	}

	mutation, result := prepareAdd(city)
	if err := gql.Query(ctx, mutation, &result); err != nil {
		return City{}, errors.Wrap(err, "failed to add city")
	}

	if len(result.AddCity.City) != 1 {
		return City{}, errors.New("city id not returned")
	}

	city.ID = result.AddCity.City[0].ID
	return city, nil
}

// =============================================================================

func prepareAdd(city City) (string, addResult) {
	var result addResult
	mutation := fmt.Sprintf(`
	mutation {
		addCity(input: [{
			name: %q
			lat: %f
			lng: %f
		}])
		%s
	}`, city.Name, city.Lat, city.Lng, result.document())

	return mutation, result
}
