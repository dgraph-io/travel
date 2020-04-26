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
	place_id: String!
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

/*
curl -H "Content-Type: application/json" http://localhost:8080/admin -XPOST
-d $'{"query": "query { getGQLSchema { generatedSchema } }"}
*/

type schema struct {
	graphql *graphql.GraphQL
}

// Create is used to identify if a schema exists. If the schema
// does not exist, then one is created.
func (s *schema) Create(ctx context.Context) error {

	// Add the schema since it doesn't exist yet.
	if err := s.graphql.CreateSchema(ctx, gQLSchema[1:], nil); err != nil {
		return errors.Wrap(err, "creating schema")
	}

	return nil
}
