// Package schema provides schema support for the database.
package schema

import (
	_ "embed" // Embed all documents

	"bytes"
	"context"
	"strings"
	"text/template"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// These variables will contain the contents of the schema and public key
// used in Dgraph for JWT support.
var (
	//go:embed graphql/public_key.pem
	publicKey string

	//go:embed graphql/schema.graphql
	schemaDoc string
)

// Schema error variables.
var (
	ErrNoSchemaExists = errors.New("no schema exists")
	ErrInvalidSchema  = errors.New("schema doesn't match")
)

// CustomFunctions is the set of custom functions defined in the schema. The
// URL to the function is required as part of the function declaration.
type CustomFunctions struct {
	UploadFeedURL string
}

// Config contains information required for the schema document.
type Config struct {
	CustomFunctions
}

// Schema provides support for schema operations against the database.
type Schema struct {
	graphql  *graphql.GraphQL
	document string
}

// New constructs a Schema value for use to manage the schema.
func New(graphql *graphql.GraphQL, config Config) (*Schema, error) {

	// The actual CRLF (\n) must be converted to the characters '\n' so the
	// entire key sits on one line.
	publicKey := strings.ReplaceAll(publicKey, "\n", "\\n")

	// Create the final schema document with the variable replacments by
	// processing the template.
	tmpl := template.New("schema")
	if _, err := tmpl.Parse(schemaDoc); err != nil {
		return nil, errors.Wrap(err, "parsing template")
	}
	var document bytes.Buffer
	vars := map[string]interface{}{
		"UploadFeedURL": config.CustomFunctions.UploadFeedURL,
		"PublicKey":     publicKey,
	}
	if err := tmpl.Execute(&document, vars); err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	schema := Schema{
		graphql:  graphql,
		document: document.String(),
	}

	return &schema, nil
}

// DropAll perform an alter operatation against the configured server
// to remove all the data and schema.
func (s *Schema) DropAll(ctx context.Context) error {
	r := strings.NewReader(`{"drop_all": true}`)
	if err := s.graphql.RawRequest(ctx, "alter", r, nil); err != nil {
		return errors.Wrap(err, "dropping schema and data")
	}

	return nil
}

// DropData perform an alter operatation against the configured server
// to remove all the data and schema.
func (s *Schema) DropData(ctx context.Context) error {
	r := strings.NewReader(`{"drop_op": "DATA"}`)
	if err := s.graphql.RawRequest(ctx, "alter", r, nil); err != nil {
		return errors.Wrap(err, "dropping data")
	}

	return nil
}

// Create is used create the schema in the database.
func (s *Schema) Create(ctx context.Context) error {
	query := `mutation updateGQLSchema($schema: String!) {
		updateGQLSchema(input: {
			set: { schema: $schema }
		}) {
			gqlSchema {
				schema
			}
		}
	}`
	err := s.graphql.ExecuteOnEndpoint(ctx, "admin", query, nil,
		graphql.WithVariable("schema", s.document),
	)
	if err != nil {
		return errors.Wrap(err, "create schema")
	}

	return nil
}
