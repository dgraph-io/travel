package data

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// ReplaceAdvisory replaces an advisory in the database and connects it
// to the specified city.
func (m mutate) ReplaceAdvisory(ctx context.Context, cityID string, adv Advisory) (Advisory, error) {
	if err := advisory.delete(ctx, m.query, m.graphql, cityID); err != nil {
		if err != ErrAdvisoryNotFound {
			return Advisory{}, errors.Wrap(err, "deleting advisory from database")
		}
	}

	adv, err := advisory.add(ctx, m.graphql, adv)
	if err != nil {
		return Advisory{}, errors.Wrap(err, "adding advisory to database")
	}

	if err := advisory.updateCity(ctx, m.graphql, cityID, adv); err != nil {
		return Advisory{}, errors.Wrap(err, "replace advisory in city")
	}

	return adv, nil
}

// =============================================================================

type adv struct {
	prepare advisoryPrepare
}

var advisory adv

func (a adv) add(ctx context.Context, graphql *graphql.GraphQL, advisory Advisory) (Advisory, error) {
	if advisory.ID != "" {
		return Advisory{}, errors.New("advisory contains id")
	}

	mutation, result := a.prepare.add(advisory)
	if err := graphql.Query(ctx, mutation, &result); err != nil {
		return Advisory{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.AddAdvisory.Advisory) != 1 {
		return Advisory{}, errors.New("advisory id not returned")
	}

	advisory.ID = result.AddAdvisory.Advisory[0].ID
	return advisory, nil
}

func (a adv) updateCity(ctx context.Context, graphql *graphql.GraphQL, cityID string, advisory Advisory) error {
	if advisory.ID == "" {
		return errors.New("advisory missing id")
	}

	mutation, result := a.prepare.updateCity(cityID, advisory)
	err := graphql.Query(ctx, mutation, &result)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	return nil
}

func (a adv) delete(ctx context.Context, query query, graphql *graphql.GraphQL, cityID string) error {
	advisory, err := query.Advisory(ctx, cityID)
	if err != nil {
		return err
	}

	mutation, result := a.prepare.delete(advisory.ID)
	if err := graphql.Query(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete advisory")
	}

	if result.DeleteAdvisory.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteAdvisory.NumUids, result.DeleteAdvisory.Msg)
		return errors.New(msg)
	}

	return nil
}

// =============================================================================

type advisoryPrepare struct{}

func (advisoryPrepare) add(advisory Advisory) (string, advisoryAddResult) {
	var result advisoryAddResult
	mutation := fmt.Sprintf(`
mutation {
	addAdvisory(input: [{
		continent: %q
		country: %q
		country_code: %q
		last_updated: %q
		message: %q
		score: %f
		source: %q
	}])
	%s
}`, advisory.Continent, advisory.Country, advisory.CountryCode,
		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source,
		result.document())

	return mutation, result
}

func (advisoryPrepare) updateCity(cityID string, advisory Advisory) (string, cityUpdateResult) {
	var result cityUpdateResult
	mutation := fmt.Sprintf(`
mutation {
	updateCity(input: {
		filter: {
		  id: [%q]
		},
		set: {
			advisory: {
				id: %q
				continent: %q
				country: %q
				country_code: %q
				last_updated: %q
				message: %q
				score: %f
				source: %q
			}
		}
	})
	%s
}`, cityID, advisory.ID, advisory.Continent, advisory.Country, advisory.CountryCode,
		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source,
		result.document())

	return mutation, result
}

func (advisoryPrepare) delete(advisoryID string) (string, advisoryDeleteResult) {
	var result advisoryDeleteResult
	mutation := fmt.Sprintf(`
mutation {
	deleteAdvisory(filter: { id: [%q] })
	%s
}`, advisoryID, result.document())

	return mutation, result
}

type advisoryAddResult struct {
	AddAdvisory struct {
		Advisory []struct {
			ID string `json:"id"`
		} `json:"advisory"`
	} `json:"addAdvisory"`
}

func (advisoryAddResult) document() string {
	return `{
		advisory {
			id
		}
	}`
}

type advisoryDeleteResult struct {
	DeleteAdvisory struct {
		Msg     string
		NumUids int
	} `json:"deleteAdvisory"`
}

func (advisoryDeleteResult) document() string {
	return `{
		msg,
		numUids,
	}`
}
