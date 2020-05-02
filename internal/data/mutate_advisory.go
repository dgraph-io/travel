package data

import (
	"context"
	"fmt"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type mutateAdvisory struct {
	marshal advisoryMarshal
}

var mutAdvisory mutateAdvisory

func (mutateAdvisory) add(ctx context.Context, graphql *graphql.GraphQL, advisory Advisory) (Advisory, error) {
	if advisory.ID != "" {
		return Advisory{}, errors.New("advisory contains id")
	}

	mutation, result := mutAdvisory.marshal.add(advisory)
	if err := graphql.Mutate(ctx, mutation, &result); err != nil {
		return Advisory{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.AddAdvisory.Advisory) != 1 {
		return Advisory{}, errors.New("advisory id not returned")
	}

	advisory.ID = result.AddAdvisory.Advisory[0].ID
	return advisory, nil
}

func (mutateAdvisory) updateCity(ctx context.Context, graphql *graphql.GraphQL, cityID string, advisory Advisory) error {
	if advisory.ID == "" {
		return errors.New("advisory missing id")
	}

	mutation, result := mutAdvisory.marshal.updCity(cityID, advisory)
	err := graphql.Mutate(ctx, mutation, &result)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

func (mutateAdvisory) delete(ctx context.Context, query query, graphql *graphql.GraphQL, cityID string) error {
	advisory, err := query.Advisory(ctx, cityID)
	if err != nil {
		return err
	}

	mutation, result := mutAdvisory.marshal.delete(advisory.ID)
	if err := graphql.Mutate(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete advisory")
	}

	if result.DeleteAdvisory.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteAdvisory.NumUids, result.DeleteAdvisory.Msg)
		return errors.New(msg)
	}

	return nil
}

type advisoryMarshal struct{}

func (advisoryMarshal) add(advisory Advisory) (string, advisoryIDResult) {
	var result advisoryIDResult
	mutation := fmt.Sprintf(`
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
	%s
}`, advisory.Continent, advisory.Country, advisory.CountryCode,
		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source,
		result.graphql())

	return mutation, result
}

func (advisoryMarshal) updCity(cityID string, advisory Advisory) (string, cityIDResult) {
	var result cityIDResult
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
	%s
}`, cityID, advisory.ID, advisory.Continent, advisory.Country, advisory.CountryCode,
		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source,
		result.graphql())

	return mutation, result
}

func (advisoryMarshal) delete(advisoryID string) (string, deleteAdvisoryResult) {
	var result deleteAdvisoryResult
	mutation := fmt.Sprintf(`
mutation {
	deleteAdvisory(filter: { id: [%q] })
	%s
}`, advisoryID, result.graphql())

	return mutation, result
}

type advisoryIDResult struct {
	AddAdvisory struct {
		Advisory []struct {
			ID string `json:"id"`
		} `json:"advisory"`
	} `json:"addAdvisory"`
}

func (advisoryIDResult) graphql() string {
	return `{
		advisory {
			id
		}
	}`
}

type deleteAdvisoryResult struct {
	DeleteAdvisory struct {
		Msg     string
		NumUids int
	} `json:"deleteAdvisory"`
}

func (deleteAdvisoryResult) graphql() string {
	return `{
		msg,
		numUids,
	}`
}
