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
enum Roles {
	ADMIN
	EMAIL
	MUTATE
	QUERY
}

type User {
	id: ID!
	email: String! @search(by: [exact])
	name: String!
	roles: [Roles]!
	password_hash: String!
	date_created: DateTime!
	date_updated: DateTime!
}

type City {
	id: ID!
	name: String! @search(by: [exact])
	lat: Float!
	lng: Float!
	places: [Place] @hasInverse(field: city)
	advisory: Advisory
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
}

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

type schema struct {
	graphql *graphql.GraphQL
}

// DropAll perform an alter operatation against the configured server
// to remove all the data and schema.
func (s *schema) DropAll(ctx context.Context) error {
	query := strings.NewReader(`{"drop_all": true}`)
	return s.graphql.Do(ctx, "alter", query, nil)
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

	var dataErrors Errors
	for {
		if err := s.graphql.QueryWithVars(ctx, graphql.CmdAdmin, query, vars, nil); err != nil {
			dataErrors = append(dataErrors, err)

			// If the context deadline exceeded then we are done trying.
			if ctx.Err() != nil {
				return errors.Wrap(dataErrors, "updating schema")
			}

			// Dgraph can fail for too many reasons. Keep trying until the
			// context deadline exceeds.
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	return nil
}

// Query returns the current schema in graphql format.
func (s *schema) Query(ctx context.Context) (string, error) {
	query := `query { getGQLSchema { schema }}`
	result := make(map[string]interface{})
	if err := s.graphql.QueryWithVars(ctx, graphql.CmdAdmin, query, nil, &result); err != nil {
		return "", errors.Wrap(err, "query schema")
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", errors.Wrap(err, "marshal schema")
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
