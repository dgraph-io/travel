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

// Client provides support for retrieving and storing results from a
// Google Places search.
type Client struct {
	mapClient *maps.Client
	dgraph    *dgo.Dgraph
}

// NewClient constructs a Client value that is initialized for use with
// Google places search and Dgraph.
func NewClient(ctx context.Context, apiKey string, dbHost string) (*Client, error) {

	// Initialize the Google maps client with our key.
	mapClient, err := maps.NewClient(maps.WithAPIKey(apiKey))
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

	// Construct the places value for use.
	client := Client{
		mapClient: mapClient,
		dgraph:    dgraph,
	}

	return &client, nil
}

// Search defines parameters that can be used in a search.
type Search struct {
	Keyword   string
	Radius    uint
	pageToken string
}

// Place represents a location that can be retrieved from a search.
type Place struct {
	Name             string   `json: "place_name"`
	Address          string   `json: "address`
	Lat              float64  `json:"lat"`
	Lng              float64  `json:"lng"`
	GooglePlaceID    string   `json: "google_place_id"`
	LocationType     []string `json: "location_type"`
	AvgUserRating    float32  `json: "avg_user_rating"`
	NumberOfRatings  int      `json: "no_user_rating"`
	GmapsURL         string   `json: "gmaps_url"`
	PhotoReferenceID string   `json: "photo_id"`
}

// City represents a city and its coordinates.
type City struct {
	*Client
	ID   string  `json:"uid"`
	Name string  `json:"city_name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// NewCity constructs a city that can be used to preform searches and
// database operations.
func NewCity(ctx context.Context, client *Client, name string, lat float64, lng float64) (*City, error) {

	// Validate the schema we need in dgraph.
	if err := validateSchema(ctx, client.dgraph); err != nil {
		return nil, err
	}

	// Construct a city value.
	city := City{
		ID : "_:sydney",
		Name:   name,
		Lat:    lat,
		Lng:    lng,
		Client: client,
	}

	// Convert the city value into json for the mutation call.
	data, err := json.Marshal(city)
	if err != nil {
		return nil, err
	}

	// Define a graphql function to find the specified city by name.
	q1 := fmt.Sprintf(`{ findCity(func: eq(city_name, %s)) { v as uid city_name } }`, city.Name)

	// examples for upserts https://github.com/dgraph-io/dgo/blob/master/upsert_test.go
	// Docs https://dgraph.io/docs/mutations/#upsert-block
	// Query variable example https://godoc.org/github.com/dgraph-io/dgo#Txn.Do

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
	result, err := client.dgraph.NewTxn().Do(ctx, &req)
	if err != nil {
		return nil, err
	}

	// When the city node is inserted for the first time
	
		if val, ok := result.Uids["sydney"]; ok {
			city.ID = val
			log.Printf("id :%s\n", val)
		}
	
	resultByte, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	log.Printf("uid1: \n%v", string(resultByte))
	log.Printf("uid1=2: \n%v", result)
	// TODO: Find if the node for sydney already exists, if yes, return the UID

	return &city, nil
}

// validateSchema is used to identify if a schema exists. If the schema
// does not exist, the one is created.
func validateSchema(ctx context.Context, dgraph *dgo.Dgraph) error {

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
		`,
	}

	if err := dgraph.Alter(ctx, op); err != nil {
		return err
	}

	return nil
}

// Search finds places for the specified search criteria.
func (city *City) Search(ctx context.Context, search Search) ([]Place, error) {

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
				Lat: city.Lat,
				Lng: city.Lng,
			},
			Keyword:   search.Keyword,
			PageToken: search.pageToken,
			Radius:    search.Radius,
		}

		// Perform the Google Places search.
		var err error
		resp, err = city.mapClient.NearbySearch(ctx, &nsr)

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
func (city *City) Store(ctx context.Context, log *log.Logger, place Place) error {

	// Convert the collection of palces into json for the mutation call.
	data, err := json.Marshal(place)
	if err != nil {
		return err
	}

	// Define a graphql function to find the specified city by name.
	query := fmt.Sprintf(`{ findPlace(func: eq(google_place_id, %s)) { v as uid  } }`, place.GooglePlaceID)

	// Mutation connecting the hotel to the City node with the `has_hotel` relationship.

	mutation := fmt.Sprintf(`{ "uid": "%s", "has_hotel" : %s }`, city.ID, string(data))

	// examples for upserts https://github.com/dgraph-io/dgo/blob/master/upsert_test.go
	// Docs https://dgraph.io/docs/mutations/#upsert-block
	// Query variable example https://godoc.org/github.com/dgraph-io/dgo#Txn.Do

	log.Printf("query: \n%s", query)
	log.Printf("mutation: \n%s", mutation)
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
	_, err = city.dgraph.NewTxn().Do(ctx, &req)
	if err != nil {
		return err
	}

	return nil
}
