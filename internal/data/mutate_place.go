package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type mutatePlace struct {
	marshal placeMarshal
}

var mutPlace mutatePlace

func (mutatePlace) exists(ctx context.Context, query query, place Place) bool {
	if _, err := query.PlaceByName(ctx, place.Name); err != nil {
		return false
	}
	return true
}

func (mutatePlace) add(ctx context.Context, graphql *graphql.GraphQL, place Place) (Place, error) {
	if place.ID != "" {
		return Place{}, errors.New("place contains id")
	}

	for i := range place.LocationType {
		if !strings.HasPrefix(place.LocationType[i], `"`) {
			place.LocationType[i] = fmt.Sprintf(`"%s"`, place.LocationType[i])
		}
	}

	mutation, result := mutPlace.marshal.add(place)
	if err := graphql.Mutate(ctx, mutation, &result); err != nil {
		return Place{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.AddPlace.Place) != 1 {
		return Place{}, errors.New("place id not returned")
	}

	place.ID = result.AddPlace.Place[0].ID
	return place, nil
}

func (mutatePlace) updateCity(ctx context.Context, graphql *graphql.GraphQL, cityID string, place Place) error {
	if place.ID == "" {
		return errors.New("place missing id")
	}

	mutation, result := mutPlace.marshal.updCity(cityID, place)
	err := graphql.Mutate(ctx, mutation, &result)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

type placeMarshal struct{}

func (placeMarshal) add(place Place) (string, placeIDResult) {
	var result placeIDResult
	mutation := fmt.Sprintf(`
mutation {
	addPlace(input: [{
		address: %q,
		avg_user_rating: %f,
		city_name: %q,
		gmaps_url: %q,
		lat: %f,
		lng: %f,
		location_type: [%q],
		name: %q,
		no_user_rating: %d,
		place_id: %q,
		photo_id: %q
	}])
	%s
}`, place.Address, place.AvgUserRating, place.CityName, place.GmapsURL,
		place.Lat, place.Lng, strings.Join(place.LocationType, ","), place.Name,
		place.NumberOfRatings, place.PlaceID, place.PhotoReferenceID,
		result.graphql())

	return mutation, result
}

func (placeMarshal) updCity(cityID string, place Place) (string, cityIDResult) {
	var result cityIDResult
	mutation := fmt.Sprintf(`
mutation {
	updateCity(input: {
		filter: {
		  id: [%q]
		},
		set: {
			places: [{
				id: %q,
				address: %q,
				avg_user_rating: %f,
				city_name: %q,
				gmaps_url: %q,
				lat: %f,
				lng: %f,
				location_type: [%q],
				name: %q,
				no_user_rating: %d,
				place_id: %q,
				photo_id: %q
			}]
		}
	})
	%s
}`, cityID, place.ID, place.Address, place.AvgUserRating, place.CityName, place.GmapsURL,
		place.Lat, place.Lng, strings.Join(place.LocationType, ","), place.Name,
		place.NumberOfRatings, place.PlaceID, place.PhotoReferenceID, result.graphql())

	return mutation, result
}

type placeIDResult struct {
	AddPlace struct {
		Place []struct {
			ID string `json:"id"`
		} `json:"place"`
	} `json:"addPlace"`
}

func (placeIDResult) graphql() string {
	return `{
		place {
			id
		}
	}`
}
