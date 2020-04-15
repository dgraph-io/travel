package data

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/dgraph-io/travel/internal/places"
	"github.com/pkg/errors"
)

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
		return errors.Wrapf(err, "op[%+v]", op)
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
	result, err := v.dgraph.NewTxn().Do(ctx, &req)
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
