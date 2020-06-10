package data

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// AddCity adds a new city to the database. If the city already exists
// this function will fail but the found city is returned. If the city is
// being added, the city with the id from the database is returned.
func (m mutate) AddCity(ctx context.Context, cty City) (City, error) {
	if city, err := m.query.CityByName(ctx, cty.Name); err == nil {
		return city, ErrCityExists
	}

	cty, err := city.add(ctx, m.graphql, cty)
	if err != nil {
		return City{}, errors.Wrap(err, "adding city to database")
	}

	return cty, nil
}

// =============================================================================

type cty struct {
	prepare ctyPrepare
}

var city cty

func (c cty) add(ctx context.Context, graphql *graphql.GraphQL, city City) (City, error) {
	if city.ID != "" {
		return City{}, errors.New("city contains id")
	}

	mutation, result := c.prepare.add(city)
	if err := graphql.Query(ctx, mutation, &result); err != nil {
		return City{}, errors.Wrap(err, "failed to add city")
	}

	if len(result.AddCity.City) != 1 {
		return City{}, errors.New("city id not returned")
	}

	city.ID = result.AddCity.City[0].ID
	return city, nil
}

// =============================================================================

type ctyPrepare struct{}

func (ctyPrepare) add(city City) (string, cityAddResult) {
	var result cityAddResult
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

type cityAddResult struct {
	AddCity struct {
		City []struct {
			ID string `json:"id"`
		} `json:"city"`
	} `json:"addCity"`
}

func (cityAddResult) document() string {
	return `{
		city {
			id
		}
	}`
}

type cityUpdateResult struct {
	UpdateCity struct {
		City []struct {
			ID string `json:"id"`
		} `json:"city"`
	} `json:"updateCity"`
}

func (cityUpdateResult) document() string {
	return `{
		city {
			id
		}
	}`
}
