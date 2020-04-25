package data

import (
	"context"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

// Maintaining alphabetical ordering since the database does this anyway.
var gQLSchema = `
type City {
	id: ID!
	advisory: Advisory
	lat: Float!
	lng: Float!
	name: String! @search(by: [term])
	places: [Place]
	weather: Weather
}

type Advisory {
	continent: String!
	country: String!
	country_code: String!
	last_updated: String
	message: String
	score: Float!
	source: String
}

type Place {
	address: String
	avg_user_rating: Float
	city_name: String!
	gmaps_url: String
	lat: Float!
	lng: Float!
	location_type: [String]
	name: String!
	no_user_rating: Int
	place_id: Int!
	photo_id: String
}

type Weather {
	city_name: String!
	description: String
	feels_like: Float
	humidity: Int
	pressure: Int
	sunrise: Int
	sunset: Int
	temp: Float
	temp_min: Float
	temp_max: Float
	visibility: String
	wind_direction: Int
	wind_speed: Float
}`

// The schema is returned by the database in alphabetical order.
var goSchema = []Schema{
	{"Advisory.continent", "string", false, nil, false},
	{"Advisory.country", "string", false, nil, false},
	{"Advisory.country_code", "string", false, nil, false},
	{"Advisory.last_updated", "string", false, nil, false},
	{"Advisory.message", "string", false, nil, false},
	{"Advisory.score", "float", false, nil, false},
	{"Advisory.source", "string", false, nil, false},

	{"City.advisory", "uid", false, nil, false},
	{"City.lat", "float", false, nil, false},
	{"City.lng", "float", false, nil, false},
	{"City.name", "string", true, []string{"term", "hash"}, true},
	{"City.places", "uid", false, nil, false},
	{"City.weather", "uid", false, nil, false},

	{"Place.address", "string", false, nil, false},
	{"Place.avg_user_rating", "float", false, nil, false},
	{"Place.city_name", "string", false, nil, false},
	{"Place.gmaps_url", "string", false, nil, false},
	{"Place.lat", "float", false, nil, false},
	{"Place.lng", "float", false, nil, false},
	{"Place.location_type", "string", false, nil, false},
	{"Place.name", "string", false, nil, false},
	{"Place.no_user_rating", "int", false, nil, false},
	{"Place.photo_id", "string", false, nil, false},
	{"Place.place_id", "int", false, nil, false},

	{"Weather.city_name", "string", false, nil, false},
	{"Weather.description", "string", false, nil, false},
	{"Weather.feels_like", "float", false, nil, false},
	{"Weather.humidity", "int", false, nil, false},
	{"Weather.pressure", "int", false, nil, false},
	{"Weather.sunrise", "int", false, nil, false},
	{"Weather.sunset", "int", false, nil, false},
	{"Weather.temp", "float", false, nil, false},
	{"Weather.temp_max", "float", false, nil, false},
	{"Weather.temp_min", "float", false, nil, false},
	{"Weather.visibility", "string", false, nil, false},
	{"Weather.wind_direction", "int", false, nil, false},
	{"Weather.wind_speed", "float", false, nil, false},

	{"dgraph.graphql.schema", "string", false, nil, false},
	{"dgraph.type", "string", true, []string{"exact"}, false},
}

// GrapQLSchema is used to return the GraphQL schema and the
// Go representation for testing purpose.
func GrapQLSchema() (string, []Schema) {
	return gQLSchema, goSchema
}

// Schema represents information per predicate set in the schema.
type Schema struct {
	Predicate string   `json:"predicate"`
	Type      string   `json:"type"`
	Index     bool     `json:"index"`
	Tokenizer []string `json:"tokenizer"`
	Upsert    bool     `json:"upsert"`
}

type schema struct {
	graphql *graphql.GraphQL
}

// Create is used to identify if a schema exists. If the schema
// does not exist, then one is created.
func (s *schema) Create(ctx context.Context) error {

	// Perform a query to validate if the schema exists.
	schema, err := s.Retrieve(ctx)
	if err != nil {
		return errors.Wrap(err, "validating schema")
	}

	// If a schema was returned, then report it exits.
	if len(schema) > 0 {
		return nil
	}

	// Add the schema since it doesn't exist yet.
	if err := s.graphql.CreateSchema(ctx, gQLSchema, nil); err != nil {
		return errors.Wrap(err, gQLSchema)
	}

	return nil
}

// Retrieve returns the defined schema from the database.
func (s *schema) Retrieve(ctx context.Context) ([]Schema, error) {
	query := "schema {}"

	var result struct {
		Schema []Schema
	}
	if err := s.graphql.QueryPM(ctx, query, &result); err != nil {
		return nil, errors.Wrap(err, query)
	}

	return result.Schema, nil
}
