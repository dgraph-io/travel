package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/dgraph-io/travel/internal/places"
	"google.golang.org/grpc"
)

// City represents a city and its coordinates.
type City struct {
	ID   string  `json:"uid"`
	Name string  `json:"city_name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// Data provides support for storing data inside of DGraph.
type Data struct {
	dgraph *dgo.Dgraph
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
		dgraph: dgraph,
	}
	return &data, nil
}

// ValidateSchema is used to identify if a schema exists. If the schema
// does not exist, the one is created.
func (d *Data) ValidateSchema(ctx context.Context) error {

	// Define a dgraph schema operation for validating and
	// creating a schema.
	op := &api.Operation{
		Schema: `
			id: string @index(hash)  .
			lat: float .
			lng: float .
			city_name: string @index(trigram, hash) @upsert .
			google_place_id: string @index(hash) @upsert .
			place_id: string @index(hash) .
			has_place: [uid] .
		`,
	}

	// Perform that operation.
	if err := d.dgraph.Alter(ctx, op); err != nil {
		return err
	}

	return nil
}

// ValidateCity is used to identify if the specified city exists in
// the database. If it doesn't, then the city is added to the database.
// It will return a new City with the city ID from the database.
func (d *Data) ValidateCity(ctx context.Context, city places.City) (string, error) {

	// Convert the city value into json for the mutation call.
	data, err := json.Marshal(city)
	if err != nil {
		return "", err
	}

	// Define a graphql function to find the specified city by name.
	q1 := fmt.Sprintf(`{ findCity(func: eq(city_name, %s)) { v as uid city_name } }`, city.Name)

	// Define and execute a request to add the city if it doesn't exist yet.
	req := api.Request{
		CommitNow: true,
		Query:     q1,
		Mutations: []*api.Mutation{
			{
				Cond:    `@if(eq(len(v), 0))`,
				SetJson: []byte(data),
			},
		},
	}
	result, err := d.dgraph.NewTxn().Do(ctx, &req)
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
		FindCity []City `json:"findCity"`
	}
	if err := json.Unmarshal(result.Json, &uid); err != nil {
		return "", err
	}
	if len(uid.FindCity) == 0 {
		return "", fmt.Errorf("unable to capture id for city: %s", result.Json)
	}
	return uid.FindCity[0].ID, nil
}

// StorePlace will add the specified Place into the database.
func (d *Data) StorePlace(ctx context.Context, log *log.Logger, cityID string, place places.Place) error {

	// Convert the place into json for the mutation call.
	data, err := json.Marshal(place)
	if err != nil {
		return err
	}

	// Define a graphql function to find the specified city by name and
	// a mutation connecting the place to the City node with the
	// `has_place` relationship.
	query := fmt.Sprintf(`{ findPlace(func: eq(google_place_id, %s)) { v as uid  } }`, place.GooglePlaceID)
	mutation := fmt.Sprintf(`{ "uid": "%s", "has_place" : %s }`, cityID, string(data))

	// Define and execute a request to add the city if it doesn't exist yet.
	req := api.Request{
		CommitNow: true,
		Query:     query,
		Mutations: []*api.Mutation{
			{
				Cond:    `@if(eq(len(v), 0))`,
				SetJson: []byte(mutation),
			},
		},
	}
	if _, err := d.dgraph.NewTxn().Do(ctx, &req); err != nil {
		log.Printf("places : Store : query : %s", query)
		log.Printf("places : Store : mutation : %s", mutation)
		return err
	}

	return nil
}
