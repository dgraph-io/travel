package places

import (
	"context"
	"encoding/json"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"googlemaps.github.io/maps"
)

// Location represents a geo-location on a map for search.
type Location struct {
	Lat     float64
	Lng     float64
	Keyword string
	Radius  uint
}

// Retrieve finds places for the specified location.
func Retrieve(ctx context.Context, apiKey string, loc Location) ([]byte, error) {
	mc, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	latLng := maps.LatLng{
		Lat: loc.Lat,
		Lng: loc.Lng,
	}
	nsr := maps.NearbySearchRequest{
		Location:  &latLng,
		Keyword:   loc.Keyword,
		PageToken: "",
		Radius:    loc.Radius,
	}
	resp, err := mc.NearbySearch(ctx, &nsr)
	if err != nil {
		return nil, err
	}

	return json.Marshal(resp.Results)
}

// Store takes the result from a retrieve and stores that into DGraph.
func Store(ctx context.Context, dbHost string, result []byte) error {
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
			SetJson: result,
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
