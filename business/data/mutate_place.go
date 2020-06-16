package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// AddPlace adds a new place to the database. If the place already exists
// this function will fail but the found place is returned. If the city is
// being added, the city with the id from the database is returned.
func (m mutate) AddPlace(ctx context.Context, plc Place) (Place, error) {
	if place, err := m.query.PlaceByName(ctx, plc.Name); err == nil {
		return place, ErrPlaceExists
	}

	plc, err := place.add(ctx, m.graphql, plc)
	if err != nil {
		return Place{}, errors.Wrap(err, "adding place to database")
	}

	if err := place.updateCity(ctx, m.graphql, plc.CityID.ID, plc.ID); err != nil {
		return Place{}, errors.Wrap(err, "adding place to city in database")
	}

	return plc, nil
}

// =============================================================================

type plc struct {
	prepare placePrepare
}

var place plc

func (p plc) add(ctx context.Context, graphql *graphql.GraphQL, place Place) (Place, error) {
	if place.ID != "" {
		return Place{}, errors.New("place contains id")
	}
	if place.CityID.ID == "" {
		return Place{}, errors.New("cityid not provided")
	}

	for i := range place.LocationType {
		if !strings.HasPrefix(place.LocationType[i], `"`) {
			place.LocationType[i] = fmt.Sprintf(`"%s"`, place.LocationType[i])
		}
	}

	mutation, result := p.prepare.add(place)
	if err := graphql.Query(ctx, mutation, &result); err != nil {
		return Place{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.AddPlace.Place) != 1 {
		return Place{}, errors.New("place id not returned")
	}

	place.ID = result.AddPlace.Place[0].ID
	return place, nil
}

func (p plc) updateCity(ctx context.Context, graphql *graphql.GraphQL, cityID string, placeID string) error {
	mutation, result := p.prepare.updateCity(cityID, placeID)
	err := graphql.Query(ctx, mutation, &result)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

// =============================================================================

type placePrepare struct{}

func (placePrepare) add(place Place) (string, placeAddResult) {
	var result placeAddResult
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
}`, place.Address, place.AvgUserRating, place.Category, place.CityID.ID, place.CityName, place.GmapsURL,
		place.Lat, place.Lng, strings.Join(place.LocationType, ","), place.Name,
		place.NumberOfRatings, place.PlaceID, place.PhotoReferenceID,
		result.document())

	return mutation, result
}

func (placePrepare) updateCity(cityID string, placeID string) (string, cityUpdateResult) {
	var result cityUpdateResult
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

type placeAddResult struct {
	AddPlace struct {
		Place []struct {
			ID string `json:"id"`
		} `json:"place"`
	} `json:"addPlace"`
}

func (placeAddResult) document() string {
	return `{
		place {
			id
		}
	}`
}
