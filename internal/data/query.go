package data

import (
	"context"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type query struct {
	graphql *graphql.GraphQL
}

// Schema return the defined schema from the database.
func (q *query) Schema(ctx context.Context) ([]Schema, error) {

	// Define the schema query.
	query := "schema {}"

	// Execute the graphql query against the database and decode the results.
	var result struct {
		Schema []Schema
	}
	if err := q.graphql.Query(ctx, graphql.CmdQuery, query, &result); err != nil {
		return nil, errors.Wrap(err, query)
	}

	return result.Schema, nil
}

// Schema represents information per predicate set in the schema.
type Schema struct {
	Predicate string   `json:"predicate"`
	Type      string   `json:"type"`
	Index     bool     `json:"index"`
	Tokenizer []string `json:"tokenizer"`
	Upsert    bool     `json:"upsert"`
}
