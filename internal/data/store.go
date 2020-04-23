package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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

	// Define a graphql function to find the specified city by name.
	mutation := fmt.Sprintf(`mutation{
		addCity(input:[
		{
			name: "%s",
			lat: %f,
			lng: %f
		}
		]){
		  city {
			name
			lat
			lng
		  }
		}
	  }`, city.Name, city.Lat, city.Lng)

	// Data structure to parse the result of the mutation
	var result struct {
		AddCity struct {
			City []places.City `json:"city"`
		} `json:"addCity"`
	}

	log.Printf("*********> City: Store: Mutation: \n%s", mutation)
	err := s.graphql.Mutate(ctx, mutation, result)
	if err != nil {
		return "", err
	}

	log.Printf("*********> City: Store: Result: \n%v", result)

	if len(result.AddCity.City) == 0 {
		err := errors.New("unable to add city, empty response from GraphQL mutation")
		return "", err
	}

	return "", nil
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
