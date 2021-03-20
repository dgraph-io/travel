// Package city provides support for managing city data in the database.
package city

import (
	"context"
	"fmt"
	"log"

	"github.com/ardanlabs/graphql"
	"github.com/dgraph-io/travel/business/data"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("city not found")
)

// Store manages the set of API's for city access.
type Store struct {
	log *log.Logger
	gql *graphql.GraphQL
}

// NewStore constructs a city store for api access.
func NewStore(log *log.Logger, gql *graphql.GraphQL) Store {
	return Store{
		log: log,
		gql: gql,
	}
}

// Upsert adds a new city to the database if it doesn't already exist by name.
// If the city already exists in the database, the function will return an City
// value with the existing id.
func (s Store) Upsert(ctx context.Context, traceID string, cty City) (City, error) {
	if cty.ID != "" {
		return City{}, errors.New("city contains id")
	}

	return s.upsert(ctx, traceID, cty)
}

// QueryByID returns the specified city from the database by the city id.
func (s Store) QueryByID(ctx context.Context, traceID string, cityID string) (City, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		id
		name
		lat
		lng
	}
}`, cityID)

	s.log.Printf("%s: %s: %s", traceID, "city.QueryByID", data.Log(query))

	var result struct {
		GetCity City `json:"getCity"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return City{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.ID == "" {
		return City{}, ErrNotFound
	}

	return result.GetCity, nil
}

// QueryByName returns the specified city from the database by the city name.
func (s Store) QueryByName(ctx context.Context, traceID string, name string) (City, error) {
	query := fmt.Sprintf(`
query {
	queryCity(filter: { name: { eq: %q } }) {
		id
		name
		lat
		lng
	}
}`, name)

	s.log.Printf("%s: %s: %s", traceID, "city.QueryByName", data.Log(query))

	var result struct {
		QueryCity []struct {
			City
		} `json:"queryCity"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return City{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryCity) != 1 {
		return City{}, ErrNotFound
	}

	return result.QueryCity[0].City, nil
}

// QueryNames returns the list of city names currently loaded in the database.
func (s Store) QueryNames(ctx context.Context, traceID string) ([]string, error) {
	query := `
	query {
		queryCity(filter: { }) {
			name
		}
	}`

	s.log.Printf("%s: %s: %s", traceID, "city.QueryNames", data.Log(query))

	var result struct {
		QueryCity []struct {
			City
		} `json:"queryCity"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if len(result.QueryCity) != 1 {
		return nil, ErrNotFound
	}

	cities := make([]string, len(result.QueryCity))
	for i, cty := range result.QueryCity {
		cities[i] = cty.Name
	}

	return cities, nil
}

// =============================================================================

func (s Store) upsert(ctx context.Context, traceID string, cty City) (City, error) {
	var result id
	mutation := fmt.Sprintf(`
	mutation {
		resp: addCity(input: [{
			name: %q
			lat: %f
			lng: %f
		}], upsert: true)
		%s
	}`, cty.Name, cty.Lat, cty.Lng, result.document())

	s.log.Printf("%s: %s: %s", traceID, "city.Upsert", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, &result); err != nil {
		return City{}, errors.Wrap(err, "failed to upsert city")
	}

	if len(result.Resp.Entities) != 1 {
		return City{}, errors.New("city id not returned")
	}

	cty.ID = result.Resp.Entities[0].ID
	return cty, nil
}
