package places

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"googlemaps.github.io/maps"
)

// Places provides support for retrieving and storing results from a
// Google Places search.
type Places struct {
	mc      *maps.Client
	dgraph  *dgo.Dgraph
	cityUID int
}

// New constructs a Places value that is initialized to both search Google
// map Places and store the results in Dgraph.
func New(ctx context.Context, city City, apiKey string, dbHost string) (*Places, error) {
	// Initialize the Google maps client with our key.
	mc, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	// Dial up an grpc connection to dgraph and construct
	// a dgraph client.
	conn, err := grpc.Dial(dbHost, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	dgraph := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	// Validate the schema we need in dgraph.
	if err := validateSchema(ctx, dgraph); err != nil {
		return nil, err
	}

	if _, err := SetCity(ctx, dgraph, city); err != nil {
		return nil, err
	}
	// Construct the places value for use.
	p := Places{
		mc:     mc,
		dgraph: dgraph,
	}

	return &p, nil
}

// validateSchema is used to identify if a schema exists. If the schema
// does not exist, the one is created.
func validateSchema(ctx context.Context, dgraph *dgo.Dgraph) error {

	// Define a dgraph schema operation for validating and
	// creating a schema.
	op := &api.Operation{
		Schema: `
			id: string @index(hash)  .
			icon: string .
			lat: float .
			lng: float .
			city_name: string @index(trigram, hash) @upsert .
			photo_reference: string @index(hash) .
			place_id: string @index(hash) .
		`,
	}

	if err := dgraph.Alter(ctx, op); err != nil {
		return err
	}

	return nil
}

// SetCity checks to see if the specified city exits in the database or
// creates the city. In either case, it returns the city id associated
// with the city.
func SetCity(ctx context.Context, dgraph *dgo.Dgraph, city City) (int, error) {

	// Convert the city value into json for the mutation call.
	data, err := json.Marshal(city)
	if err != nil {
		return 0, err
	}
	log.Printf("city: %s", string(data))
	// examples for upserts https://github.com/dgraph-io/dgo/blob/master/upsert_test.go
	// Docs https://dgraph.io/docs/mutations/#upsert-block
	// Query variable example https://godoc.org/github.com/dgraph-io/dgo#Txn.Do

	q1 := `
{
  findCity(func: eq(city_name, %s)) {
	v as uid
	city_name
  }
}`

	// add the city name to the query string
	q1 = fmt.Sprintf(q1, city.Name)

	log.Printf("query: \n %s", q1)

	// Create a node for the city, only if it doesn't exist.
	req := &api.Request{
		CommitNow: true,
		Query:     q1,
		Mutations: []*api.Mutation{
			&api.Mutation{
				Cond:    ` @if(eq(len(v), 0)) `,
				SetJson: []byte(data),
			},
		},
	}

	// Update email only if matching uid found.
	upsertResp, err := dgraph.NewTxn().Do(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(" upsert %v", upsertResp)
	// TODO: Find if the node for sydney already exists, if yes, return the UID

	return 0, nil
}

// Search finds places for the specified search criteria.
func (p *Places) Search(ctx context.Context, search Search) ([]Place, error) {

	// If this call is not looking for page 1, we need to pace
	// the searches out. We are using three seconds.
	if search.pageToken != "" {
		time.Sleep(3000 * time.Millisecond)
	}

	// We will make three attempts to perform a search. You need to
	// space your paged searches by an undefined amount of time :(.
	// The call may result in an INVALID_REQUEST error if the call
	// is happening at a pace too fast for the API.
	var resp maps.PlacesSearchResponse
	for i := 0; i < 3; i++ {

		// Construct the search request value for the call.
		nsr := maps.NearbySearchRequest{
			Location: &maps.LatLng{
				Lat: search.Lat,
				Lng: search.Lng,
			},
			Keyword:   search.Keyword,
			PageToken: search.pageToken,
			Radius:    search.Radius,
		}

		// Perform the Google Places search.
		var err error
		resp, err = p.mc.NearbySearch(ctx, &nsr)

		// This is the problem. We need to check for the INVALID_REQUEST
		// error. The only way to do that is to compare this string :(
		// If this is the error, then wait for a second before trying again.
		if err != nil {
			if err.Error() == "maps: INVALID_REQUEST - " {
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			return nil, err
		}
		break
	}

	// For quick reference
	// https://godoc.org/googlemaps.github.io/maps#NearbySearchRequest

	var places []Place
	for _, result := range resp.Results {

		// Validate if a photo even exists for this place.
		var photoReferenceID string
		if len(result.Photos) == 0 {
			photoReferenceID = result.Photos[0].PhotoReference
		}

		// Construct a place value based on search results.
		place := Place{
			Name:             result.Name,
			Address:          result.FormattedAddress,
			Lat:              result.Geometry.Location.Lat,
			Lng:              result.Geometry.Location.Lng,
			GooglePlaceID:    result.PlaceID,
			LocationType:     result.Types,
			AvgUserRating:    result.Rating,
			NumberOfRatings:  result.UserRatingsTotal,
			PhotoReferenceID: photoReferenceID,
		}

		// Save the place in the collection of places.
		places = append(places, place)
	}

	// If the NextPageToken on the result is empty, we have all
	// the results. Send an EOF to confirm that back to the caller.
	search.pageToken = resp.NextPageToken
	if resp.NextPageToken == "" {
		return places, io.EOF
	}

	return places, nil
}

// Store takes the result from a retrieve and stores that into DGraph.
func (p *Places) Store(ctx context.Context, log *log.Logger, place Place) error {

	// Convert the collection of palces into json for the mutation call.
	data, err := json.Marshal(place)
	if err != nil {
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
