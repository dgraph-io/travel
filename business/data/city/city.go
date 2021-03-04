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

// City manages the set of API's for city access.
type City struct {
	log *log.Logger
	gql *graphql.GraphQL
}

// New constructs a City for api access.
func New(log *log.Logger, gql *graphql.GraphQL) City {
	return City{
		log: log,
		gql: gql,
	}
}

// Upsert adds a new city to the database if it doesn't already exist by name.
// If the city already exists in the database, the function will return an Info
// value with the existing id.
func (c City) Upsert(ctx context.Context, traceID string, cty Info) (Info, error) {
	if cty.ID != "" {
		return Info{}, errors.New("city contains id")
	}

	return c.upsert(ctx, traceID, cty)
}

// QueryByID returns the specified city from the database by the city id.
func (c City) QueryByID(ctx context.Context, traceID string, cityID string) (Info, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		id
		name
		lat
		lng
	}
}`, cityID)

	c.log.Printf("%s: %s: %s", traceID, "city.QueryByID", data.Log(query))

	var result struct {
		GetCity Info `json:"getCity"`
	}
	if err := c.gql.Query(ctx, query, &result); err != nil {
		return Info{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.ID == "" {
		return Info{}, ErrNotFound
	}

	return result.GetCity, nil
}

// QueryByName returns the specified city from the database by the city name.
func (c City) QueryByName(ctx context.Context, traceID string, name string) (Info, error) {
	query := fmt.Sprintf(`
query {
	queryCity(filter: { name: { eq: %q } }) {
		id
		name
		lat
		lng
	}
}`, name)

	c.log.Printf("%s: %s: %s", traceID, "city.QueryByName", data.Log(query))

	var result struct {
		QueryCity []struct {
			Info
		} `json:"queryCity"`
	}
	if err := c.gql.Query(ctx, query, &result); err != nil {
		return Info{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryCity) != 1 {
		return Info{}, ErrNotFound
	}

	return result.QueryCity[0].Info, nil
}

// QueryNames returns the list of city names currently loaded in the database.
func (c City) QueryNames(ctx context.Context, traceID string) ([]string, error) {
	query := `
	query {
		queryCity(filter: { }) {
			name
		}
	}`

	c.log.Printf("%s: %s: %s", traceID, "city.QueryNames", data.Log(query))

	var result struct {
		QueryCity []struct {
			Info
		} `json:"queryCity"`
	}
	if err := c.gql.Query(ctx, query, &result); err != nil {
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

func (c City) upsert(ctx context.Context, traceID string, cty Info) (Info, error) {
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

	c.log.Printf("%s: %s: %s", traceID, "city.Upsert", data.Log(mutation))

	if err := c.gql.Query(ctx, mutation, &result); err != nil {
		return Info{}, errors.Wrap(err, "failed to upsert city")
	}

	if len(result.Resp.Entities) != 1 {
		return Info{}, errors.New("city id not returned")
	}

	cty.ID = result.Resp.Entities[0].ID
	return cty, nil
}
