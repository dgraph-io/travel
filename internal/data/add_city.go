package data

import (
	"context"
	"fmt"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type _addCity struct{}

var addCity _addCity

func (_addCity) exists(ctx context.Context, query query, city City) bool {
	_, err := query.CityByName(ctx, city.Name)
	if err != nil {
		return false
	}
	return true
}

func (_addCity) add(ctx context.Context, graphql *graphql.GraphQL, city City) (City, error) {
	if city.ID != "" {
		return City{}, errors.New("city contains id")
	}

	var result struct {
		AddCity struct {
			City []struct {
				ID string `json:"id"`
			} `json:"city"`
		} `json:"addCity"`
	}

	if err := graphql.Mutate(ctx, addCity.marshalAdd(city), &result); err != nil {
		return City{}, errors.Wrap(err, "failed to add city")
	}

	if len(result.AddCity.City) != 1 {
		return City{}, errors.New("city id not returned")
	}

	city.ID = result.AddCity.City[0].ID
	return city, nil
}

func (_addCity) marshalAdd(city City) string {
	return fmt.Sprintf(`
	mutation {
		addCity(input: [
			{name: %q, lat: %f, lng: %f}
		])
		{
			city {
				id
			}
		}
	}`, city.Name, city.Lat, city.Lng)
}
