package places

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"googlemaps.github.io/maps"
)

// Location represents a geo-location on a map for search.
type Location struct {
	Lat       float64
	Lng       float64
	Keyword   string
	Radius    uint
	pageToken string
}

// Places provides support for retrieving and storing results from
// a Google Places search.
type Places struct {
	mc     *maps.Client
	dgraph *dgo.Dgraph
}

// New constructs a Places value that is initialized to both
// search Google map Places and store the results in Dgraph.
func New(apiKey string, dbHost string) (*Places, error) {

	// Initialize the Google maps client with our key.
	mc, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	// Dial up an grpc connection to Dgraph.
	conn, err := grpc.Dial(dbHost, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// Construct the places value for use.
	p := Places{
		mc: mc,
		dgraph: dgo.NewDgraphClient(
			api.NewDgraphClient(conn),
		),
	}

	return &p, nil
}

// Retrieve finds places for the specified location.
func (p *Places) Retrieve(ctx context.Context, loc *Location) ([]byte, error) {

	// If this call is not looking for page 1, we need to pace
	// the searches out. We are using 1/2 second for now.
	if loc.pageToken != "" {
		time.Sleep(500 * time.Millisecond)
	}

	// We will make three attempts to perform a search. You need to
	// space your paged searches by an undefined amount of time :(.
	// The call may result in an INVALID_REQUEST error if the call
	// is happening at a pace too fast for the API.
	var resp maps.PlacesSearchResponse
	for i := 0; i < 3; i++ {
		nsr := maps.NearbySearchRequest{
			Location: &maps.LatLng{
				Lat: loc.Lat,
				Lng: loc.Lng,
			},
			Keyword:   loc.Keyword,
			PageToken: loc.pageToken,
			Radius:    loc.Radius,
		}

		// Perform the Google Places search.
		var err error
		resp, err = p.mc.NearbySearch(ctx, &nsr)

		// This is the problem. We need to check for the INVALID_REQUEST
		// error. The only way to do that is to compare this string :(
		// If this is the error, then wait 1/2 a second before trying again.
		if err != nil {
			if err.Error() == "maps: INVALID_REQUEST - " {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			return nil, err
		}
		break
	}

	// Marshal the result to JSON for processing with Dgraph.
	data, err := json.Marshal(resp.Results)
	if err != nil {
		return nil, err
	}

	// If the NextPageToken on the result is empty, we have all
	// the results. Send an EOF to confirm that back to the caller.
	loc.pageToken = resp.NextPageToken
	if resp.NextPageToken == "" {
		return data, io.EOF
	}

	return data, nil
}

// Store takes the result from a retrieve and stores that into DGraph.
func (p *Places) Store(ctx context.Context, data []byte) error {
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
	if err := p.dgraph.Alter(ctx, op); err != nil {
		return err
	}

	txn := p.dgraph.NewTxn()
	{
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
	}

	return nil
}
