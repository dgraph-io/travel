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
