package data

import (
	"context"
	"fmt"

	"github.com/dgraph-io/travel/internal/places"
	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type query struct {
	*graphql.GraphQL
}

// Schema returns the defined schema from the database.
func (q *query) Schema(ctx context.Context) ([]Schema, error) {
	query := "schema {}"

	var result struct {
		Schema []Schema
	}
	if err := q.Query(ctx, graphql.CmdQuery, query, &result); err != nil {
		return nil, errors.Wrap(err, query)
	}

	return result.Schema, nil
}

// CityByID returns the specified city from the database by ID.
func (q *query) CityByID(ctx context.Context, cityID string) (places.City, error) {
	query := fmt.Sprintf("{city(func: uid(%s)) {city_name lat lng}}", cityID)

	var result struct {
		City []places.City
	}
	if err := q.Query(ctx, graphql.CmdQuery, query, &result); err != nil {
		return places.City{}, errors.Wrap(err, query)
	}

	if len(result.City) == 0 {
		return places.City{}, errors.Wrap(errors.New("no city found"), query)
	}

	return result.City[0], nil
}

// Schema represents information per predicate set in the schema.
type Schema struct {
	Predicate string   `json:"predicate"`
	Type      string   `json:"type"`
	Index     bool     `json:"index"`
	Tokenizer []string `json:"tokenizer"`
	Upsert    bool     `json:"upsert"`
}
