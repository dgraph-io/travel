package places

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"googlemaps.github.io/maps"
)

// Places provides support for retrieving and storing results from
// a Google Places search.
type Places struct {
	mc     *maps.Client
	dgraph *dgo.Dgraph
}

// New constructs a Places value that is initialized to both
// search Google map Places and store the results in Dgraph.
func New(apiKey string, dbHost string, CityInfo City) (*Places, error) {
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

	// initialize the schema once
	op := &api.Operation{
		Schema: `
			id: string @index(hash)  .
			icon: string .
			lat: float .
			lng: float .
			name: string @index(trigram, hash) .
			photo_reference: string @index(hash) .
			place_id: string @index(hash) .
			scope: string @index(hash) .
		`,
	}

	// Construct the places value for use.
	p := Places{
		mc: mc,
		dgraph: dgo.NewDgraphClient(
			api.NewDgraphClient(conn),
		),
	}

	log.Printf("Initializing the database schema")

	if err := p.dgraph.Alter(ctx, op); err != nil {
		return nil, err
	}

	// TODO: Find if the node for sydney already exists, if yes, return the UID
	log.Printf("Storing the city information for Sydney")
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

	return &p, nil
}

func (p *Places) Retrieve(ctx context.Context, loc *PlacesSearchRequest) ([]Place, error) {
	// Retrieve finds places for the specified location.

	// If this call is not looking for page 1, we need to pace
	// the searches out. We are using three seconds.
	if loc.pageToken != "" {
		time.Sleep(3000 * time.Millisecond)
	}

	// We will make three attempts to perform a search. You need to
	// space your paged searches by an undefined amount of time :(.
	// The call may result in an INVALID_REQUEST error if the call
	// is happening at a pace too fast for the API.
	var resp maps.PlacesSearchResponse

	var placesResult []Places

	for i := 0; i < 3; i++ {
		// Construct the search request value for the call.
		nsr := maps.NearbySearchRequest{
			Location: &maps.LatLng{
				Lat: loc.CityInfo.Lat,
				Lng: loc.CityInfo.Lng,
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

	// For quick refeence https://godoc.org/googlemaps.github.io/maps#NearbySearchRequest
	searchResults := resp.Results
	placeResult := Places{}

	for i := 0; i < len(searchResults); i++ {
		placeResult.Name = searchResults[i].Name
		placeResult.Address = searchResults[i].Geometry.FormattedAddress
		placeResult.Lat = searchResults[i].Geometry.Location.Lat
		placeResult.Lng = searchResults[i].Geometry.Location.Lng
		placeResult.GooglePlaceID = searchResults[i].PlaceID
		placeResult.LocationType = searchResults[i].Types
		placeResult.AvgUserRating = searchResults[i].Rating
		placeResult.NumberOfRatings = searchResults[i].UserRatingsTotal
		placeResult.PhotoReferenceID = searchResults[i].FormattedAddress
		if len(SearchResults[i].Photos) {
			placeResult.PhotoReferenceID = SearchResults[i].Photos[0].PhotoReference
		}
		placesResult.append(placeResult)
	}

	fmt.Printf("\nplace: \n %v", places[0])
	// If the NextPageToken on the result is empty, we have all
	// the results. Send an EOF to confirm that back to the caller.
	loc.pageToken = resp.NextPageToken
	if resp.NextPageToken == "" {
		return placesResult, io.EOF
	}

	return placeResult, nil
}

// Store takes the result from a retrieve and stores that into DGraph.
func (p *Places) Store(ctx context.Context, placesResult []Place) error {
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
