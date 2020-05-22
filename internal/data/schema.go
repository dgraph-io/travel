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
var gQLSchema = `
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

type schema struct {
	graphql *graphql.GraphQL
}

// Create is used create the schema in the database.
func (s *schema) Create(ctx context.Context) error {
	got, err := s.Query(ctx)
	if err != nil {
		return errors.Wrap(err, "creating schema")
	}

	switch {
	case got == `{"getGQLSchema":null}`:
		for i := 0; i < 2; i++ {
			if err := s.graphql.CreateSchema(ctx, gQLSchema, nil); err != nil {

				// Dgraph can fail because it's not ready to accept a schema
				// yet. Just retry once if this is the case.
				if i == 0 {
					time.Sleep(time.Second)
					continue
				}
				return errors.Wrap(err, "creating schema")
			}
			break
		}
	default:
		if err := s.Validate(ctx); err != nil {
			return errors.Wrap(err, "creating schema")
		}
	}

	return nil
}

// Query returns the current schema in graphql format.
func (s *schema) Query(ctx context.Context) (string, error) {
	result := make(map[string]interface{})
	if err := s.graphql.QuerySchema(ctx, &result); err != nil {
		return "", errors.Wrap(err, "validate schema")
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
	got, err := s.Query(ctx)
	if err != nil {
		return errors.Wrap(err, "validate schema")
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return errors.Wrap(err, "validate schema")
	}

	exp := reg.ReplaceAllString(gQLSchema, "")
	got = strings.ReplaceAll(got[27:], "\\n", "")
	got = strings.ReplaceAll(got, "\\t", "")
	got = reg.ReplaceAllString(got, "")

	if exp != got {
		return errors.New("invalid schema")
	}

	return nil
}
