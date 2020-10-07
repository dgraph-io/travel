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

// City manages the set of API's for city access.
type City struct {
	gql *graphql.GraphQL
}

// New constructs a City for api access.
func New(gql *graphql.GraphQL) City {
	return City{
		gql: gql,
	}
}

// Add adds a new city to the database. If the city already exists
// this function will fail but the found city is returned. If the city is
// being added, the city with the id from the database is returned.
func (c City) Add(ctx context.Context, cty Info) (Info, error) {
	if cty, err := c.QueryByName(ctx, cty.Name); err == nil {
		return cty, ErrExists
	}

	cty, err := c.add(ctx, cty)
	if err != nil {
		return Info{}, errors.Wrap(err, "adding city to database")
	}

	return cty, nil
}

// QueryByID returns the specified city from the database by the city id.
func (c City) QueryByID(ctx context.Context, cityID string) (Info, error) {
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
func (c City) QueryByName(ctx context.Context, name string) (Info, error) {
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
func (c City) QueryNames(ctx context.Context) ([]string, error) {
	query := `
	query {
		queryCity(filter: { }) {
			name
		}
	}`

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

func (c City) add(ctx context.Context, cty Info) (Info, error) {
	if cty.ID != "" {
		return Info{}, errors.New("city contains id")
	}

	mutation, result := prepareAdd(cty)
	if err := c.gql.Query(ctx, mutation, &result); err != nil {
		return Info{}, errors.Wrap(err, "failed to add city")
	}

	if len(result.AddCity.City) != 1 {
		return Info{}, errors.New("city id not returned")
	}

	cty.ID = result.AddCity.City[0].ID
	return cty, nil
}

// =============================================================================

func prepareAdd(cty Info) (string, addResult) {
	var result addResult
	mutation := fmt.Sprintf(`
	mutation {
		addCity(input: [{
			name: %q
			lat: %f
			lng: %f
		}])
		%s
	}`, cty.Name, cty.Lat, cty.Lng, result.document())

	return mutation, result
}
