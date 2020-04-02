package feed

import (
	"context"
	"encoding/json"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"googlemaps.github.io/maps"
)

type location struct {
	lat     float64
	lng     float64
	keyword string
	radius  uint
}

func retrieveLocation(ctx context.Context, apiKey string, loc location) ([]byte, error) {
	mc, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	latLng := maps.LatLng{
		Lat: loc.lat,
		Lng: loc.lng,
	}
	nsr := maps.NearbySearchRequest{
		Location:  &latLng,
		Keyword:   loc.keyword,
		PageToken: "",
		Radius:    loc.radius,
	}
	resp, err := mc.NearbySearch(ctx, &nsr)
	if err != nil {
		return nil, err
	}

	return json.Marshal(resp.Results)
}

func storeLocation(ctx context.Context, dbHost string, jsonData []byte) error {
	conn, err := grpc.Dial(dbHost, grpc.WithInsecure())
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
	{
		mut := api.Mutation{
			SetJson: jsonData,
		}
		if _, err := txn.Mutate(ctx, &mut); err != nil {
			txn.Discard(ctx)
			return err
		}

		if err := txn.Commit(ctx); err != nil {
			return nil
		}
	}

	return nil
}
