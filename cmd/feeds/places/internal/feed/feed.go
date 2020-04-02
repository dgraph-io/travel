package feed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"googlemaps.github.io/maps"
)

// Pull extracts the feed from the source.
func Pull(ctx context.Context, host string) error {
	mc, err := maps.NewClient(maps.WithAPIKey("AIzaSyBR0-ToiYlrhPlhidE7DA-Zx7EfE7FnUek"))
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
		PageToken: "",
		Radius:    5000,
	}
	resp, err := mc.NearbySearch(context.TODO(), &nsr)
	if err != nil {
		return err
	}

	data, err := json.Marshal(resp.Results)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := dgo.NewDgraphClient(
		api.NewDgraphClient(conn),
	)

	op := &api.Operation{
		Schema: `
			height: int .
			width: int .
			id: string @index(hash)  .
			icon: string .
			lat: float .
			lng: float .
			location_type: string  .
			name: string @index(trigram, hash) .
			photo_reference: string @index(hash) .
			place_id: string @index(hash) .
			scope: string @index(hash) .
			vicinity: string @index(trigram) .
			types: [string] .
			html_attributions: [string] .

			location: uid .
			bounds: uid .
			geometry: uid @reverse .
			photos: [uid] @reverse @count .
			northeast: uid .
			southwest: uid .
			viewport: uid .
		`,
	}
	if err := client.Alter(ctx, op); err != nil {
		return err
	}

	txn := client.NewTxn()

	mut := api.Mutation{
		SetJson: data,
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
