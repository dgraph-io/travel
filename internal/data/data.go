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
	"github.com/dgraph-io/travel/internal/weather"
	"google.golang.org/grpc"
)

// Data provides support for storing data inside of DGraph.
type Data struct {
	Validate validate
	Store    store
}

// New constructs a data value for use to store data inside
// of the Dgraph database.
func New(dbHost string) (*Data, error) {

	// Dial up an grpc connection to dgraph and construct
	// a dgraph client.
	conn, err := grpc.Dial(dbHost, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	dgraph := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	data := Data{
		Store:    store{dgraph: dgraph},
		Validate: validate{dgraph: dgraph},
	}
	return &data, nil
}

type validate struct {
	dgraph *dgo.Dgraph
}

// Schema is used to identify if a schema exists. If the schema
// does not exist, the one is created.
func (v *validate) Schema(ctx context.Context) error {

	// Define a dgraph schema operation for validating and
	// creating a schema.
	op := &api.Operation{
		Schema: `
			lat: float .
			lng: float .
			city_name: string @index(trigram, hash) @upsert .
			place_id: string @index(hash) @upsert .
			weather_id: int @index(int) @upsert .
			
			weather: [uid] .
			places: [uid] .
		`,
	}

	// Perform that operation.
	if err := v.dgraph.Alter(ctx, op); err != nil {
		return err
	}

	return nil
}

// City is used to identify if the specified city exists in
// the database. If it doesn't, then the city is added to the database.
// It will return a new City with the city ID from the database.
func (v *validate) City(ctx context.Context, city places.City) (string, error) {

	// Convert the city value into json for the mutation call.
	data, err := json.Marshal(city)
	if err != nil {
		return "", err
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
	result, err := v.dgraph.NewTxn().Do(ctx, &req)
	if err != nil {
		return "", err
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
		return "", err
	}
	if len(uid.FindCity) == 0 {
		return "", fmt.Errorf("unable to capture id for city: %s", result.Json)
	}
	return uid.FindCity[0].ID, nil
}

type store struct {
	dgraph *dgo.Dgraph
}

// Weather will add the specified Place into the database.
func (s *store) Weather(ctx context.Context, log *log.Logger, cityID string, w weather.Weather) error {

	// Add the city id to the weather node.
	db := struct {
		// TODO: Just connect the weather node with city node via an edge.
		// No need to use a foriegn key kind of relationship.
		CityID string `json:"city_id"`
		weather.Weather
	}{
		CityID:  cityID,
		Weather: w,
	}

	// Convert the data to store into json for the mutation call.
	data, err := json.Marshal(db)
	if err != nil {
		return err
	}

	// Define a graphql function to find the weather by its unique id. The
	// cityID will be the unique id for the weather.
	// TODO: Instead, check whether the City node has it's weather information available
	// via the `weather` edge. Also update the weather info if its last udpated time
	// is more than 24 hours.
	// TODO: We need to flip the feed fetching.
	// TODO: First find whether the weather info exists and its not outdated, and
	// TODO: then go fetch from the Feed only if required.
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
	if _, err := s.dgraph.NewTxn().Do(ctx, &req); err != nil {
		log.Printf("places : StoreWeather : query : %s", query)
		log.Printf("places : StoreWeather : mutation : %s", mutation)
		return err
	}

	return nil
}

// Place will add the specified Place into the database.
func (s *store) Place(ctx context.Context, log *log.Logger, cityID string, place places.Place) error {

	// Add the city id to the place node.
	db := struct {
		// TODO: Establish the relationship by creating an edge with the city node.
		CityID string `json:"city_id"`
		places.Place
	}{
		CityID: cityID,
		Place:  place,
	}

	// Convert the data to store into json for the mutation call.
	data, err := json.Marshal(db)
	if err != nil {
		return err
	}

	// Define a graphql function to find a place by its unique id. The
	// GooglePlaceID will be the unique id for the place.
	query := fmt.Sprintf(`{ findPlace(func: eq(place_id, %s)) { id as uid } }`, place.PlaceID)

	//  Define a mutation by connecting the place to the city with the
	// `has_place` relationship.
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
	if _, err := s.dgraph.NewTxn().Do(ctx, &req); err != nil {
		log.Printf("places : StorePlace : query : %s", query)
		log.Printf("places : StorePlace : mutation : %s", mutation)
		return err
	}

	return nil
}

func (s *store) Advisory(ctx context.Context, cityID string, a advisory.Advisory) error {

	// Add the city id to the weather node.
	db := struct {
		// TODO: Establish the relationship by creating an edge with the city node.
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

	// Check Check whether Advisory for the country code exists.
	// TODO: Check the last update time of the advisory, update if it's more than 7 days.
	// TODO: We need to flip the feed fetching.
	// TODO: First find whether the Advisory exists and its not outdated, and
	// TODO: then go fetch from the Feed only if required.
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
	if _, err := s.dgraph.NewTxn().Do(ctx, &req); err != nil {
		log.Printf("places : StoreAdvisory : query : %s", query)
		log.Printf("places : StoreAdvisory : mutation : %s", mutation)
		return err
	}

	return nil
}
