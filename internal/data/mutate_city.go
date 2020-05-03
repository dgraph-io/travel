package data

import (
	"context"
	"fmt"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type mutateCity struct {
	marshal cityMarshal
}

var mutCity mutateCity

func (mutateCity) add(ctx context.Context, graphql *graphql.GraphQL, city City) (City, error) {
	if city.ID != "" {
		return City{}, errors.New("city contains id")
	}

	mutation, result := mutCity.marshal.add(city)
	if err := graphql.Mutate(ctx, mutation, &result); err != nil {
		return City{}, errors.Wrap(err, "failed to add city")
	}

	if len(result.AddCity.City) != 1 {
		return City{}, errors.New("city id not returned")
	}

	city.ID = result.AddCity.City[0].ID
	return city, nil
}

type cityMarshal struct{}

func (cityMarshal) add(city City) (string, cityIDResult) {
	var result cityIDResult
	mutation := fmt.Sprintf(`
	mutation {
		addCity(input: [
			{name: %q, lat: %f, lng: %f}
		])
		%s
	}`, city.Name, city.Lat, city.Lng, result.graphql())

	return mutation, result
}

type cityIDResult struct {
	AddCity struct {
		City []struct {
			ID string `json:"id"`
		} `json:"city"`
	} `json:"addCity"`
}

func (cityIDResult) graphql() string {
	return `{
		city {
			id
		}
	}`
}
