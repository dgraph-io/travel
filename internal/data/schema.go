package data

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// Schema error variables.
var (
	ErrNoSchemaExists = errors.New("no schema exists")
	ErrInvalidSchema  = errors.New("schema doesn't match")
)

type schema struct {
	graphql *graphql.GraphQL
}

// DropAll perform an alter operatation against the configured server
// to remove all the data and schema.
func (s *schema) DropAll(ctx context.Context) error {
	// if _, err := s.retrieve(ctx); err != nil {
	// 	return errors.Wrap(err, "can't drop schema, db not ready")
	// }

	query := strings.NewReader(`{"drop_all": true}`)
	if err := s.graphql.Do(ctx, "alter", query, nil); err != nil {
		return errors.Wrap(err, "dropping schema and data")
	}

	schema, err := s.retrieve(ctx)
	if err != nil {
		return errors.Wrap(err, "can't validate schema, db not ready")
	}

	if err := s.validate(ctx, schema); err != ErrNoSchemaExists {
		return errors.Wrap(err, "unable to drop schema and data")
	}

	return nil
}

// Create is used create the schema in the database.
func (s *schema) Create(ctx context.Context) error {
	schema, err := s.retrieve(ctx)
	if err != nil {
		return errors.Wrap(err, "can't create schema, db not ready")
	}

	// If the schema matches against what we know the
	// schema to be, don't try to update it.
	if err := s.validate(ctx, schema); err == nil {
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

	if err := s.graphql.QueryWithVars(ctx, graphql.CmdAdmin, query, vars, nil); err != nil {
		return errors.Wrap(err, "create schema")
	}

	schema, err = s.retrieve(ctx)
	if err != nil {
		return errors.Wrap(err, "can't create schema, db not ready")
	}

	if err := s.validate(ctx, schema); err != nil {
		return errors.Wrap(err, "invalid schema")
	}

	return nil
}

// retrieve queries the database for the schema and handles situations
// when the database is not ready for schema operations.
func (s *schema) retrieve(ctx context.Context) (string, error) {
	for {
		schema, err := s.query(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "Server not ready") {

				// If the context deadline exceeded then we are done trying.
				if ctx.Err() != nil {
					return "", errors.Wrap(err, "server not ready")
				}

				// We need to wait for the server to be ready for this :(.
				time.Sleep(2 * time.Second)
				continue
			}

			return "", errors.Wrap(err, "server not ready")
		}

		return schema, nil
	}
}

func (s *schema) query(ctx context.Context) (string, error) {
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

func (s *schema) validate(ctx context.Context, schema string) error {
	if schema == `{"getGQLSchema":null}` || schema == `{"getGQLSchema":{"schema":""}}` {
		return ErrNoSchemaExists
	}

	if len(schema) < 27 {
		return ErrInvalidSchema
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return errors.Wrap(err, "regex compile")
	}

	exp := strings.ReplaceAll(gQLSchema, "\\n", "")
	exp = reg.ReplaceAllString(exp, "")
	schema = strings.ReplaceAll(schema[27:], "\\n", "")
	schema = strings.ReplaceAll(schema, "\\t", "")
	schema = reg.ReplaceAllString(schema, "")

	if exp != schema {
		log.Println("exp:", exp)
		log.Println("got:", schema)
		return ErrInvalidSchema
	}

	return nil
}
