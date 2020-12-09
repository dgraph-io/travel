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
	ErrExists   = errors.New("place exists")
	ErrNotFound = errors.New("place not found")
)

// Place manages the set of API's for place access.
type Place struct {
	log *log.Logger
	gql *graphql.GraphQL
}

// New constructs a Place for api access.
func New(log *log.Logger, gql *graphql.GraphQL) Place {
	return Place{
		log: log,
		gql: gql,
	}
}

// Add adds a new place to the database. If the place already exists
// this function will fail but the found place is returned. If the city is
// being added, the city with the id from the database is returned.
func (p Place) Add(ctx context.Context, traceID string, plc Info) (Info, error) {
	if plc.ID != "" {
		return Info{}, errors.New("place contains id")
	}
	if plc.City.ID == "" {
		return Info{}, errors.New("cityid not provided")
	}

	if plc, err := p.QueryByName(ctx, traceID, plc.Name); err == nil {
		return plc, ErrExists
	}

	plc, err := p.add(ctx, traceID, plc)
	if err != nil {
		return Info{}, errors.Wrap(err, "adding place to database")
	}

	return plc, nil
}

// QueryByID returns the specified place from the database by the place id.
func (p Place) QueryByID(ctx context.Context, traceID string, placeID string) (Info, error) {
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

	p.log.Printf("%s: %s: %s", traceID, "place.QueryByID", data.Log(query))

	var result struct {
		GetPlace struct {
			Info
		} `json:"getPlace"`
	}
	if err := p.gql.Query(ctx, query, &result); err != nil {
		return Info{}, errors.Wrap(err, "query failed")
	}

	if result.GetPlace.Info.ID == "" {
		return Info{}, ErrNotFound
	}

	return result.GetPlace.Info, nil
}

// QueryByName returns the specified place from the database by name.
func (p Place) QueryByName(ctx context.Context, traceID string, name string) (Info, error) {
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

	p.log.Printf("%s: %s: %s", traceID, "place.QueryByName", data.Log(query))

	var result struct {
		QueryPlace []Info `json:"queryPlace"`
	}
	if err := p.gql.Query(ctx, query, &result); err != nil {
		return Info{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return Info{}, ErrNotFound
	}

	return result.QueryPlace[0], nil
}

// QueryByCategory returns the collection of places from the database
// by the cagtegory name.
func (p Place) QueryByCategory(ctx context.Context, traceID string, category string) ([]Info, error) {
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

	p.log.Printf("%s: %s: %s", traceID, "place.QueryByCategory", data.Log(query))

	var result struct {
		QueryPlace []Info `json:"queryPlace"`
	}
	if err := p.gql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	if len(result.QueryPlace) != 1 {
		return nil, ErrNotFound
	}

	return result.QueryPlace, nil
}

// QueryByCity returns the collection of places from the database by the city id.
func (p Place) QueryByCity(ctx context.Context, traceID string, cityID string) ([]Info, error) {
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

	p.log.Printf("%s: %s: %s", traceID, "place.QueryByCity", data.Log(query))

	var result struct {
		GetCity struct {
			Places []Info `json:"places"`
		} `json:"getCity"`
	}
	if err := p.gql.Query(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, "query failed")
	}

	return result.GetCity.Places, nil
}

// =============================================================================

func (p Place) add(ctx context.Context, traceID string, plc Info) (Info, error) {
	for i := range plc.LocationType {
		if !strings.HasPrefix(plc.LocationType[i], `"`) {
			plc.LocationType[i] = fmt.Sprintf(`"%s"`, plc.LocationType[i])
		}
	}

	mutation, result := prepareAdd(plc)
	p.log.Printf("%s: %s: %s", traceID, "place.Add", data.Log(mutation))

	if err := p.gql.Query(ctx, mutation, &result); err != nil {
		return Info{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.Resp.Entities) != 1 {
		return Info{}, errors.New("place id not returned")
	}

	plc.ID = result.Resp.Entities[0].ID
	return plc, nil
}

// =============================================================================

func prepareAdd(plc Info) (string, id) {
	var result id
	mutation := fmt.Sprintf(`
mutation {
	resp: addPlace(input: [{
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
}`, plc.Address, plc.AvgUserRating, plc.Category, plc.City.ID, plc.CityName, plc.GmapsURL,
		plc.Lat, plc.Lng, strings.Join(plc.LocationType, ","), plc.Name,
		plc.NumberOfRatings, plc.PlaceID, plc.PhotoReferenceID,
		result.document())

	return mutation, result
}
