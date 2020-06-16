package data

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// Place returns the specified place from the database by the place id.
func (q query) Place(ctx context.Context, placeID string) (Place, error) {
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
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return Place{}, errors.Wrap(err, "query failed")
	}

	if result.GetPlace.Place.ID == "" {
		return Place{}, ErrPlaceNotFound
	}

	return result.GetPlace.Place, nil
}

// PlaceByName returns the specified place from the database by name.
func (q query) PlaceByName(ctx context.Context, name string) (Place, error) {
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
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return Place{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return Place{}, ErrPlaceNotFound
	}

	return result.QueryPlace[0], nil
}

// PlaceByCategory returns the collection of places from the database
// by the cagtegory name.
func (q query) PlaceByCategory(ctx context.Context, category string) ([]Place, error) {
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
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return nil, ErrPlaceNotFound
	}

	return result.QueryPlace, nil
}

// Places returns the collection of places from the database by the city id.
func (q query) Places(ctx context.Context, cityID string) ([]Place, error) {
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
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	return result.GetCity.Places, nil
}
