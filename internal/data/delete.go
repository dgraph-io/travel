package data

import (
	"context"
	"fmt"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type delete struct {
	query   query
	graphql *graphql.GraphQL
}

// Advisory will delete the current Advisory from the database.
func (d *delete) Advisory(ctx context.Context, cityID string) error {

	// Get the current advisory for the city.
	advisory, err := d.query.Advisory(ctx, cityID)
	if err != nil {
		return errors.Wrap(err, "delete advisory")
	}

	// Define a graphql mutation to delete the advisory in the database
	// for the specified city.
	mutation := fmt.Sprintf(`
mutation {
	deleteAdvisory(filter: { id: [%q] }) {
		msg,
		numUids,
	}
}`, advisory.ID)

	var result struct {
		Msg     string
		NumUids int
	}
	if err := d.graphql.Mutate(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete advisory")
	}

	if result.NumUids != 1 {
		return errors.Wrap(err, "failed to delete advisory")
	}

	return nil
}
