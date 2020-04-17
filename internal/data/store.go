package data

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/dgraph-io/travel/internal/advisory"
	"github.com/dgraph-io/travel/internal/places"
	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/dgraph-io/travel/internal/weather"
	"github.com/pkg/errors"
)

type store struct {
	graphql *graphql.GraphQL
	*dgo.Dgraph
}

// City is used to identify if the specified city exists in
// the database. If it doesn't, then the city is added to the database.
// It will return a new City with the city ID from the database.
func (s *store) City(ctx context.Context, city places.City) (string, error) {

	// Convert the city value into json for the mutation call.
	data, err := json.Marshal(city)
	if err != nil {
		return "", errors.Wrapf(err, "city[%+v]", city)
	}

	// Define a graphql function to find the specified city by name.
	q1 := fmt.Sprintf(`{ findCity(func: eq(city_name, %s)) { id as uid } }`, city.Name)

	// Define and execute a request to add the city if it doesn't exist yet.
	req := api.Request{
		CommitNow: true,
		Query:     q1,
		Mutations: []*api.Mutation{
			{
				// Only perform the mutation if the node for city.Name
				// doesn't exist yet.
				Cond:    `@if(eq(len(id), 0))`,
				SetJson: []byte(data),
			},
		},
	}
	result, err := s.NewTxn().Do(ctx, &req)
	if err != nil {
		return "", errors.Wrapf(err, "req[%+v]", &req)
	}

	// If there is a key/value pair inside of this map of
	// Uids, then we just added the city to the database.
	// This is the only way to get this new ID.
	if len(result.Uids) == 1 {
		for _, id := range result.Uids {
			return id, nil
		}
	}

	// City id was not found in the result map, so look for it
	// in the json response from the database.
	var uid struct {
		FindCity []struct {
			ID string `json:"uid"`
		} `json:"findCity"`
	}
	if err := json.Unmarshal(result.Json, &uid); err != nil {
		return "", errors.Wrapf(err, "json[%+v]", result.Json)
	}
	if len(uid.FindCity) == 0 {
		err := errors.New("unable to find city")
		return "", errors.Wrapf(err, "city[%s]", uid.FindCity)
	}
	return uid.FindCity[0].ID, nil
}

// Advisory will add the specified Advisory into the database.
func (s *store) Advisory(ctx context.Context, cityID string, a advisory.Advisory) error {

	// Add the city id to the weather node.
	db := struct {
		CityID string `json:"city_id"`
		advisory.Advisory
	}{
		CityID:   cityID,
		Advisory: a,
	}

	// Convert the data to store into json for the mutation call.
	data, err := json.Marshal(db)
	if err != nil {
		return err
	}

	// Check whether Advisory for the country code exists.
	query := fmt.Sprintf(`{ findAdvisory(func: eq(country_code, %s)) { id as uid } }`, a.CountryCode)

	//  Define a mutation by connecting the advisory to the city with the
	// `advisory` relationship.
	mutation := fmt.Sprintf(`{ "uid": "%s", "advisory" : %s }`, cityID, string(data))

	// Define and execute a request to add the city if it doesn't exist yet.
	req := api.Request{
		CommitNow: true,
		Query:     query,
		Mutations: []*api.Mutation{
			{
				Cond:    `@if(eq(len(id), 0))`,
				SetJson: []byte(mutation),
			},
		},
	}
	if _, err := s.NewTxn().Do(ctx, &req); err != nil {
		return errors.Wrapf(err, "req[%+v] query[%s] mut[%s]", &req, query, mutation)
	}

	return nil
}

// Weather will add the specified Place into the database.
func (s *store) Weather(ctx context.Context, cityID string, w weather.Weather) error {

	// Add the city id to the weather node.
	db := struct {
		CityID string `json:"city_id"`
		weather.Weather
	}{
		CityID:  cityID,
		Weather: w,
	}

	// Convert the data to store into json for the mutation call.
	data, err := json.Marshal(db)
	if err != nil {
		return errors.Wrapf(err, "db[%+v]", db)
	}

	// Define a graphql function to find the weather by its unique id. The
	// cityID will be the unique id for the weather.
	query := fmt.Sprintf(`{ findWeather(func: eq(weather_id, %d)) { id as uid } }`, w.ID)

	//  Define a mutation by connecting the weather to the city with the
	// `weather` relationship.
	mutation := fmt.Sprintf(`{ "uid": "%s", "weather" : %s }`, cityID, string(data))

	// Define and execute a request to add the city if it doesn't exist yet.
	req := api.Request{
		CommitNow: true,
		Query:     query,
		Mutations: []*api.Mutation{
			{
				Cond:    `@if(eq(len(id), 0))`,
				SetJson: []byte(mutation),
			},
		},
	}
	if _, err := s.NewTxn().Do(ctx, &req); err != nil {
		return errors.Wrapf(err, "req[%+v] query[%s] mut[%s]", &req, query, mutation)
	}

	return nil
}

// Place will add the specified Place into the database.
func (s *store) Place(ctx context.Context, cityID string, place places.Place) error {

	// Add the city id to the place node.
	db := struct {
		CityID string `json:"city_id"`
		places.Place
	}{
		CityID: cityID,
		Place:  place,
	}

	// Convert the data to store into json for the mutation call.
	data, err := json.Marshal(db)
	if err != nil {
		return errors.Wrapf(err, "db[%+v]", db)
	}

	// Define a graphql function to find a place by its unique id. The
	// GooglePlaceID will be the unique id for the place.
	query := fmt.Sprintf(`{ findPlace(func: eq(place_id, %s)) { id as uid } }`, place.PlaceID)

	//  Define a mutation by connecting the place to the city with the
	// `places` relationship.
	mutation := fmt.Sprintf(`{ "uid": "%s", "places" : %s }`, cityID, string(data))

	// Define and execute a request to add the city if it doesn't exist yet.
	req := api.Request{
		CommitNow: true,
		Query:     query,
		Mutations: []*api.Mutation{
			{
				Cond:    `@if(eq(len(id), 0))`,
				SetJson: []byte(mutation),
			},
		},
	}
	if _, err := s.NewTxn().Do(ctx, &req); err != nil {
		return errors.Wrapf(err, "req[%+v] query[%s] mut[%s]", &req, query, mutation)
	}

	return nil
}
