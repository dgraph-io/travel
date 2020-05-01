package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type _addPlace struct{}

var addPlace _addPlace

func (_addPlace) exists(ctx context.Context, query query, place Place) bool {
	_, err := query.PlaceByName(ctx, place.Name)
	if err != nil {
		return false
	}
	return true
}

func (_addPlace) add(ctx context.Context, graphql *graphql.GraphQL, place Place) (Place, error) {
	if place.ID != "" {
		return Place{}, errors.New("place contains id")
	}

	for i := range place.LocationType {
		if !strings.HasPrefix(place.LocationType[i], `"`) {
			place.LocationType[i] = fmt.Sprintf(`"%s"`, place.LocationType[i])
		}
	}

	var result struct {
		AddPlace struct {
			Place []struct {
				ID string `json:"id"`
			} `json:"place"`
		} `json:"addPlace"`
	}

	if err := graphql.Mutate(ctx, addPlace.marshalAdd(place), &result); err != nil {
		return Place{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.AddPlace.Place) != 1 {
		return Place{}, errors.New("place id not returned")
	}

	place.ID = result.AddPlace.Place[0].ID
	return place, nil
}

func (_addPlace) updateCity(ctx context.Context, graphql *graphql.GraphQL, cityID string, place Place) error {
	if place.ID == "" {
		return errors.New("place missing id")
	}

	err := graphql.Mutate(ctx, addPlace.marshalUpdCity(cityID, place), nil)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

func (_addPlace) marshalAdd(place Place) string {
	return fmt.Sprintf(`
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
	{
		place {
			id
		}
	}
}`, place.Address, place.AvgUserRating, place.CityName, place.GmapsURL,
		place.Lat, place.Lng, strings.Join(place.LocationType, ","), place.Name,
		place.NumberOfRatings, place.PlaceID, place.PhotoReferenceID)
}

func (_addPlace) marshalUpdCity(cityID string, place Place) string {
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
	{
		city {
			id
		}
	}
}`, cityID, place.ID, place.Address, place.AvgUserRating, place.CityName, place.GmapsURL,
		place.Lat, place.Lng, strings.Join(place.LocationType, ","), place.Name,
		place.NumberOfRatings, place.PlaceID, place.PhotoReferenceID)

	return mutation
}
