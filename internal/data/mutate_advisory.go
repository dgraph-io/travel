package data

import (
	"context"
	"fmt"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type _mutAdvisory struct{}

var mutAdvisory _mutAdvisory

func (_mutAdvisory) add(ctx context.Context, graphql *graphql.GraphQL, advisory Advisory) (Advisory, error) {
	if advisory.ID != "" {
		return Advisory{}, errors.New("advisory contains id")
	}

	var result struct {
		AddAdvisory struct {
			Advisory []struct {
				ID string `json:"id"`
			} `json:"advisory"`
		} `json:"addAdvisory"`
	}

	if err := graphql.Mutate(ctx, mutAdvisory.marshalAdd(advisory), &result); err != nil {
		return Advisory{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.AddAdvisory.Advisory) != 1 {
		return Advisory{}, errors.New("advisory id not returned")
	}

	advisory.ID = result.AddAdvisory.Advisory[0].ID
	return advisory, nil
}

func (_mutAdvisory) updateCity(ctx context.Context, graphql *graphql.GraphQL, cityID string, advisory Advisory) error {
	if advisory.ID == "" {
		return errors.New("advisory missing id")
	}

	err := graphql.Mutate(ctx, mutAdvisory.marshalUpdCity(cityID, advisory), nil)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

func (_mutAdvisory) delete(ctx context.Context, query query, graphql *graphql.GraphQL, cityID string) error {
	advisory, err := query.Advisory(ctx, cityID)
	if err != nil {
		return err
	}

	var result struct {
		DeleteAdvisory struct {
			Msg     string
			NumUids int
		} `json:"deleteAdvisory"`
	}
	if err := graphql.Mutate(ctx, mutAdvisory.marshalDelete(advisory.ID), &result); err != nil {
		return errors.Wrap(err, "failed to delete advisory")
	}

	if result.DeleteAdvisory.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteAdvisory.NumUids, result.DeleteAdvisory.Msg)
		return errors.New(msg)
	}

	return nil
}

func (_mutAdvisory) marshalAdd(advisory Advisory) string {
	return fmt.Sprintf(`
mutation {
	addAdvisory(input: [{
		continent: %q,
		country: %q,
		country_code: %q,
		last_updated: %q,
		message: %q,
		score: %f,
		source: %q
	}])
	{
		advisory {
			id
		}
	}
}`, advisory.Continent, advisory.Country, advisory.CountryCode,
		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source)
}

func (_mutAdvisory) marshalUpdCity(cityID string, advisory Advisory) string {
	mutation := fmt.Sprintf(`
mutation {
	updateCity(input: {
		filter: {
		  id: [%q]
		},
		set: {
			advisory: {
				id: %q,
				continent: %q,
				country: %q,
				country_code: %q,
				last_updated: %q,
				message: %q,
				score: %f,
				source: %q
			}
		}
	})
	{
		city {
			id
		}
	}
}`, cityID, advisory.ID, advisory.Continent, advisory.Country, advisory.CountryCode,
		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source)

	return mutation
}

func (_mutAdvisory) marshalDelete(advisoryID string) string {
	return fmt.Sprintf(`
mutation {
	deleteAdvisory(filter: { id: [%q] })
	{
		msg,
		numUids,
	}
}`, advisoryID)
}
