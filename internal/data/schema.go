package data

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// This is the schema for the application. This could be kept in a file
// and maintained for wider use. In these cases I would use gogenerate
// to hardcode the contents into the binary.
const gQLSchema = `
type City {
	id: ID!
	advisory: Advisory
	lat: Float!
	lng: Float!
	name: String! @search(by: [exact])
	places: [Place] @hasInverse(field: city)
	weather: Weather
}

type Advisory {
	id: ID!
	continent: String!
	country: String!
	country_code: String!
	last_updated: String
	message: String
	score: Float!
	source: String
}

type Place {
	id: ID!
	address: String
	avg_user_rating: Float
	category: String @search(by: [exact])
	city: City!
	city_name: String!
	gmaps_url: String
	lat: Float!
	lng: Float!
	location_type: [String]
	name: String! @search(by: [exact])
	no_user_rating: Int
	place_id: String!
	photo_id: String
}

type Weather {
	id: ID!
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

// Waiting on Dgraph to support this in a stable version.
/*
	const gQLCustomFunctions = `
	type EmailResponse @remote {
		id: ID!
		email: String
		subject: String
		err: String
	}

	type Query{
		sendEmail(email: String!, subject: String!): EmailResponse @custom(http:{
			url: "http://0.0.0.0:3000/v1/email",
			method: "POST",
			body: "{ email: $email, subject: $subject }"
		})
	}`
*/

type schema struct {
	graphql *graphql.GraphQL
}

// QuerySchema performs a schema query operation against the configured server.
func (s *schema) QuerySchema(ctx context.Context, response interface{}) error {
	query := `query { getGQLSchema { schema }}`
	return s.graphql.QueryWithVars(ctx, graphql.CmdAdmin, query, nil, response)
}

// Create is used create the schema in the database.
func (s *schema) Create(ctx context.Context) error {
	if err := s.Validate(ctx); err == nil {
		return nil
	}

	query := `mutation updateGQLSchema($schema: String!) {
		updateGQLSchema(input: {
			set: { schema: $schema }
		}) {
			gqlSchema {
				schema
			}
		}
	}`
	vars := map[string]interface{}{"schema": gQLSchema}

	// Give the database 10 seconds to accept the schema.
	numRetries := 10
	for i := 1; i <= numRetries; i++ {
		if err := s.graphql.QueryWithVars(ctx, graphql.CmdAdmin, query, vars, nil); err != nil {

			// Dgraph can fail because it's not ready to accept a schema yet or if indexing is going
			// on in background. Just retry a few times if this is the case.
			if i < numRetries {
				time.Sleep(time.Second)
				continue
			}
			return errors.Wrap(err, "updating schema")
		}
		break
	}

	return nil
}

// Query returns the current schema in graphql format.
func (s *schema) Query(ctx context.Context) (string, error) {
	result := make(map[string]interface{})
	if err := s.QuerySchema(ctx, &result); err != nil {
		return "", errors.Wrap(err, "query schema")
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", errors.Wrap(err, "validate schema")
	}

	return string(data), nil
}

// Validate compares the schema in the database with what is
// defined for the application.
func (s *schema) Validate(ctx context.Context) error {
	schema, err := s.Query(ctx)
	if err != nil {
		return errors.Wrap(err, "query schema")
	}

	if schema == `{"getGQLSchema":null}` {
		return errors.New("no schema exists")
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return errors.Wrap(err, "regex compile")
	}

	exp := reg.ReplaceAllString(gQLSchema, "")
	schema = strings.ReplaceAll(schema[27:], "\\n", "")
	schema = strings.ReplaceAll(schema, "\\t", "")
	schema = reg.ReplaceAllString(schema, "")

	if exp != schema {
		return errors.New("invalid schema")
	}

	return nil
}

// DropAll perform an alter operatation against the configured server
// to remove all the data and schema.
func (s *schema) DropAll(ctx context.Context) error {
	query := strings.NewReader(`{"drop_all": true}`)
	return s.graphql.Do(ctx, "alter", query, nil)
}
