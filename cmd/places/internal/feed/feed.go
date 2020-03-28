package feed

import (
	"context"
	"encoding/json"
	"fmt"

	"log"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/machinebox/graphql"
	"google.golang.org/grpc"
	"googlemaps.github.io/maps"
)

// Person ...
type Person struct {
	Name    string
	Balance int
}

// Response ...
type Response struct {
	All []Person
}

// Query ...
func Query(ctx context.Context) error {
	gql := graphql.NewClient("http://localhost:8080/query")

	q := `{
all(func: anyofterms(Name, "Bill")) {
	Name
	Balance
}}`
	req := graphql.NewRequest(q)
	req.Header.Set("Cache-Control", "no-cache")

	var respData Response
	if err := gql.Run(ctx, req, &respData); err != nil {
		return err
	}

	fmt.Printf("%+v\n", respData)
	return nil
}

// DB ...
func DB(ctx context.Context, host string) error {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := dgo.NewDgraphClient(
		api.NewDgraphClient(conn),
	)

	op := api.Operation{
		DropOp: api.Operation_ALL,
	}
	if err := client.Alter(ctx, &op); err != nil {
		return err
	}

	op = api.Operation{
		Schema: `
			Name: string @index(term) .
			Balance: int .
		`,
	}
	if err := client.Alter(context.Background(), &op); err != nil {
		return err
	}

	people := []Person{
		{Name: "Bill", Balance: 10},
		{Name: "Ale", Balance: 100},
	}
	out, err := json.Marshal(people)
	if err != nil {
		return err
	}

	txn := client.NewTxn()

	mut := api.Mutation{
		SetJson: out,
	}
	if _, err := txn.Mutate(ctx, &mut); err != nil {
		txn.Discard(ctx)
		return err
	}

	if err := txn.Commit(ctx); err != nil {
		return nil
	}

	return nil
}

// Pull extracts the feed from the source.
func Pull(log *log.Logger) error {

	// Construct a new client for API access.
	mc, err := maps.NewClient(maps.WithAPIKey("AIzaSyAA6GLbxGfMf_8E7VeiwCqB_ukJtCXN5p4"))
	if err != nil {
		return err
	}

	latLng := maps.LatLng{
		Lat: -33.865143,
		Lng: 151.209900,
	}
	nsr := maps.NearbySearchRequest{
		Location:  &latLng,
		Keyword:   "Sydney",
		PageToken: "pg1",
	}
	resp, err := mc.NearbySearch(context.TODO(), &nsr)
	if err != nil {
		return err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	log.Println(string(data))

	return nil
}
