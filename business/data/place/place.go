// Package place provides support for managing place data in the database.
package place

import (
	"context"
	"fmt"
	"strings"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrExists   = errors.New("place exists")
	ErrNotFound = errors.New("place not found")
)

// Add adds a new place to the database. If the place already exists
// this function will fail but the found place is returned. If the city is
// being added, the city with the id from the database is returned.
func Add(ctx context.Context, gql *graphql.GraphQL, place Place) (Place, error) {
	if place.ID != "" {
		return Place{}, errors.New("place contains id")
	}
	if place.City.ID == "" {
		return Place{}, errors.New("cityid not provided")
	}

	if place, err := OneByName(ctx, gql, place.Name); err == nil {
		return place, ErrExists
	}

	place, err := add(ctx, gql, place)
	if err != nil {
		return Place{}, errors.Wrap(err, "adding place to database")
	}

	return place, nil
}

// One returns the specified place from the database by the place id.
func One(ctx context.Context, gql *graphql.GraphQL, placeID string) (Place, error) {
	query := fmt.Sprintf(`
query {
	getPlace(id: %q) {
		id
		address
		avg_user_rating
		category
		city {
			id
		}
		city_name
		gmaps_url
		lat
		lng
		location_type
		name
		no_user_rating
		place_id
		photo_id
	}
}`, placeID)

	var result struct {
		GetPlace struct {
			Place
		} `json:"getPlace"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return Place{}, errors.Wrap(err, "query failed")
	}

	if result.GetPlace.Place.ID == "" {
		return Place{}, ErrNotFound
	}

	return result.GetPlace.Place, nil
}

// OneByName returns the specified place from the database by name.
func OneByName(ctx context.Context, gql *graphql.GraphQL, name string) (Place, error) {
	query := fmt.Sprintf(`
query {
	queryPlace(filter: { name: { eq: %q } }) {
		id
		address
		avg_user_rating
		category
		city {
			id
		}
		city_name
		gmaps_url
		lat
		lng
		location_type
		name
		no_user_rating
		place_id
		photo_id
	}
}`, name)

	var result struct {
		QueryPlace []Place `json:"queryPlace"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return Place{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return Place{}, ErrNotFound
	}

	return result.QueryPlace[0], nil
}

// OneByCategory returns the collection of places from the database
// by the cagtegory name.
func OneByCategory(ctx context.Context, gql *graphql.GraphQL, category string) ([]Place, error) {
	query := fmt.Sprintf(`
query {
	queryPlace(filter: { category: { eq: %q } }) {
		id
		address
		avg_user_rating
		category
		city {
			id
		}
		city_name
		gmaps_url
		lat
		lng
		location_type
		name
		no_user_rating
		place_id
		photo_id
	}
}`, category)

	var result struct {
		QueryPlace []Place `json:"queryPlace"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return nil, ErrNotFound
	}

	return result.QueryPlace, nil
}

// List returns the collection of places from the database by the city id.
func List(ctx context.Context, gql *graphql.GraphQL, cityID string) ([]Place, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		places {
			id
			address
			avg_user_rating
			category
			city {
				id
			}
			city_name
			gmaps_url
			lat
			lng
			location_type
			name
			no_user_rating
			place_id
			photo_id
		}
	}
}`, cityID)

	var result struct {
		GetCity struct {
			Places []Place `json:"places"`
		} `json:"getCity"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	return result.GetCity.Places, nil
}

// =============================================================================

func add(ctx context.Context, gql *graphql.GraphQL, place Place) (Place, error) {
	for i := range place.LocationType {
		if !strings.HasPrefix(place.LocationType[i], `"`) {
			place.LocationType[i] = fmt.Sprintf(`"%s"`, place.LocationType[i])
		}
	}

	mutation, result := prepareAdd(place)
	if err := gql.Query(ctx, mutation, &result); err != nil {
		return Place{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.AddPlace.Place) != 1 {
		return Place{}, errors.New("place id not returned")
	}

	place.ID = result.AddPlace.Place[0].ID
	return place, nil
}

func updateCity(ctx context.Context, gql *graphql.GraphQL, cityID string, placeID string) error {
	mutation, result := prepareUpdateCity(cityID, placeID)
	err := gql.Query(ctx, mutation, &result)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

// =============================================================================

func prepareAdd(place Place) (string, addResult) {
	var result addResult
	mutation := fmt.Sprintf(`
mutation {
	addPlace(input: [{
		address: %q
		avg_user_rating: %f
		category: %q
		city: {
			id: %q
		}
		city_name: %q
		gmaps_url: %q
		lat: %f
		lng: %f
		location_type: [%q]
		name: %q
		no_user_rating: %d
		place_id: %q
		photo_id: %q
	}])
	%s
}`, place.Address, place.AvgUserRating, place.Category, place.City.ID, place.CityName, place.GmapsURL,
		place.Lat, place.Lng, strings.Join(place.LocationType, ","), place.Name,
		place.NumberOfRatings, place.PlaceID, place.PhotoReferenceID,
		result.document())

	return mutation, result
}

func prepareUpdateCity(cityID string, placeID string) (string, updateCityResult) {
	var result updateCityResult
	mutation := fmt.Sprintf(`
mutation {
	updateCity(input: {
		filter: {
		  id: [%q]
		},
		set: {
			places: [{
				id: %q
			}]
		}
	})
	%s
}`, cityID, placeID, result.document())

	return mutation, result
}
