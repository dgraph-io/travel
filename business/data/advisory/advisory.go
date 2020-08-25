// Package advisory provides support for managing advisory data in the database.
package advisory

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("advisory not found")
)

// Replace replaces an advisory in the database and connects it
// to the specified city.
func Replace(ctx context.Context, gql *graphql.GraphQL, advisory Advisory) (Advisory, error) {
	if advisory.ID != "" {
		return Advisory{}, errors.New("advisory contains id")
	}
	if advisory.City.ID == "" {
		return Advisory{}, errors.New("cityid not provided")
	}

	if err := delete(ctx, gql, advisory.City.ID); err != nil {
		if err != ErrNotFound {
			return Advisory{}, errors.Wrap(err, "deleting advisory from database")
		}
	}

	advisory, err := add(ctx, gql, advisory)
	if err != nil {
		return Advisory{}, errors.Wrap(err, "adding advisory to database")
	}

	return advisory, nil
}

// One returns the specified advisory from the database by the city id.
func One(ctx context.Context, gql *graphql.GraphQL, cityID string) (Advisory, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		advisory {
			id
			city {
				id
			}
			continent
			country
			country_code
			last_updated
			message
			score
			source
		}
	}
}`, cityID)

	var result struct {
		GetCity struct {
			Advisory Advisory `json:"advisory"`
		} `json:"getCity"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return Advisory{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.Advisory.ID == "" {
		return Advisory{}, ErrNotFound
	}

	return result.GetCity.Advisory, nil
}

// =============================================================================

func add(ctx context.Context, gql *graphql.GraphQL, advisory Advisory) (Advisory, error) {
	mutation, result := prepareAdd(advisory)
	if err := gql.Query(ctx, mutation, &result); err != nil {
		return Advisory{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.AddAdvisory.Advisory) != 1 {
		return Advisory{}, errors.New("advisory id not returned")
	}

	advisory.ID = result.AddAdvisory.Advisory[0].ID
	return advisory, nil
}

func delete(ctx context.Context, gql *graphql.GraphQL, cityID string) error {
	advisory, err := One(ctx, gql, cityID)
	if err != nil {
		return err
	}

	mutation, result := prepareDelete(advisory.ID)
	if err := gql.Query(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete advisory")
	}

	if result.DeleteAdvisory.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteAdvisory.NumUids, result.DeleteAdvisory.Msg)
		return errors.New(msg)
	}

	return nil
}

// =============================================================================

func prepareAdd(advisory Advisory) (string, addResult) {
	var result addResult
	mutation := fmt.Sprintf(`
mutation {
	addAdvisory(input: [{
		city: {
			id: %q
		}
		continent: %q
		country: %q
		country_code: %q
		last_updated: %q
		message: %q
		score: %f
		source: %q
	}])
	%s
}`, advisory.City.ID, advisory.Continent, advisory.Country, advisory.CountryCode,
		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source,
		result.document())

	return mutation, result
}

func prepareDelete(advisoryID string) (string, deleteResult) {
	var result deleteResult
	mutation := fmt.Sprintf(`
mutation {
	deleteAdvisory(filter: { id: [%q] })
	%s
}`, advisoryID, result.document())

	return mutation, result
}
