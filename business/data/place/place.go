// Package place provides support for managing place data in the database.
package place

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ardanlabs/graphql"
	"github.com/dgraph-io/travel/business/data"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("place not found")
)

// Store manages the set of API's for place access.
type Store struct {
	log *log.Logger
	gql *graphql.GraphQL
}

// NewStore constructs a place store for api access.
func NewStore(log *log.Logger, gql *graphql.GraphQL) Store {
	return Store{
		log: log,
		gql: gql,
	}
}

// Upsert adds a new place to the database if it doesn't already exist by name.
// If the place already exists in the database, the function will return an Place
// value with the existing id.
func (s Store) Upsert(ctx context.Context, traceID string, plc Place) (Place, error) {
	if plc.ID != "" {
		return Place{}, errors.New("place contains id")
	}
	if plc.City.ID == "" {
		return Place{}, errors.New("cityid not provided")
	}

	for i := range plc.LocationType {
		if !strings.HasPrefix(plc.LocationType[i], `"`) {
			plc.LocationType[i] = fmt.Sprintf(`"%s"`, plc.LocationType[i])
		}
	}

	return s.upsert(ctx, traceID, plc)
}

// QueryByID returns the specified place from the database by the place id.
func (s Store) QueryByID(ctx context.Context, traceID string, placeID string) (Place, error) {
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

	s.log.Printf("%s: %s: %s", traceID, "place.QueryByID", data.Log(query))

	var result struct {
		GetPlace struct {
			Place
		} `json:"getPlace"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return Place{}, errors.Wrap(err, "query failed")
	}

	if result.GetPlace.Place.ID == "" {
		return Place{}, ErrNotFound
	}

	return result.GetPlace.Place, nil
}

// QueryByName returns the specified place from the database by name.
func (s Store) QueryByName(ctx context.Context, traceID string, name string) (Place, error) {
	query := fmt.Sprintf(`
query {
	queryPlace(filter: { name: { alloftext: %q } }) {
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

	s.log.Printf("%s: %s: %s", traceID, "place.QueryByName", data.Log(query))

	var result struct {
		QueryPlace []Place `json:"queryPlace"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return Place{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return Place{}, ErrNotFound
	}

	return result.QueryPlace[0], nil
}

// QueryByCategory returns the collection of places from the database
// by the cagtegory name.
func (s Store) QueryByCategory(ctx context.Context, traceID string, category string) ([]Place, error) {
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

	s.log.Printf("%s: %s: %s", traceID, "place.QueryByCategory", data.Log(query))

	var result struct {
		QueryPlace []Place `json:"queryPlace"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return nil, ErrNotFound
	}

	return result.QueryPlace, nil
}

// QueryByCity returns the collection of places from the database by the city id.
func (s Store) QueryByCity(ctx context.Context, traceID string, cityID string) ([]Place, error) {
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

	s.log.Printf("%s: %s: %s", traceID, "place.QueryByCity", data.Log(query))

	var result struct {
		GetCity struct {
			Places []Place `json:"places"`
		} `json:"getCity"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	return result.GetCity.Places, nil
}

// =============================================================================

func (s Store) upsert(ctx context.Context, traceID string, plc Place) (Place, error) {
	var result id
	mutation := fmt.Sprintf(`
	mutation {
		resp: addPlace(input: [{
			name: %q
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
			no_user_rating: %d
			place_id: %q
			photo_id: %q
		}], upsert: true)
		%s
	}`, plc.Name, plc.Address, plc.AvgUserRating, plc.Category, plc.City.ID,
		plc.CityName, plc.GmapsURL, plc.Lat, plc.Lng, strings.Join(plc.LocationType, ","),
		plc.NumberOfRatings, plc.PlaceID, plc.PhotoReferenceID,
		result.document())

	s.log.Printf("%s: %s: %s", traceID, "place.Upsert", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, &result); err != nil {
		return Place{}, errors.Wrap(err, "failed to upsert place")
	}

	if len(result.Resp.Entities) != 1 {
		return Place{}, errors.New("place id not returned")
	}

	plc.ID = result.Resp.Entities[0].ID
	return plc, nil
}
