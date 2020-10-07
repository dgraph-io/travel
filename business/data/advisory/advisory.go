// Package advisory provides support for managing advisory data in the database.
package advisory

import (
	"context"
	"fmt"
	"log"

	"github.com/ardanlabs/graphql"
	"github.com/dgraph-io/travel/business/data"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("advisory not found")
)

// Advisory manages the set of API's for advisory access.
type Advisory struct {
	log *log.Logger
	gql *graphql.GraphQL
}

// New constructs a Advisory for api access.
func New(log *log.Logger, gql *graphql.GraphQL) Advisory {
	return Advisory{
		log: log,
		gql: gql,
	}
}

// Replace replaces an advisory in the database and connects it
// to the specified city.
func (a Advisory) Replace(ctx context.Context, traceID string, adv Info) (Info, error) {
	if adv.ID != "" {
		return Info{}, errors.New("advisory contains id")
	}
	if adv.City.ID == "" {
		return Info{}, errors.New("cityid not provided")
	}

	if err := a.delete(ctx, traceID, adv.City.ID); err != nil {
		if err != ErrNotFound {
			return Info{}, errors.Wrap(err, "deleting advisory from database")
		}
	}

	adv, err := a.add(ctx, traceID, adv)
	if err != nil {
		return Info{}, errors.Wrap(err, "adding advisory to database")
	}

	return adv, nil
}

// QueryByCity returns the specified advisory from the database by the city id.
func (a Advisory) QueryByCity(ctx context.Context, traceID string, cityID string) (Info, error) {
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

	a.log.Printf("%s: %s: %s", traceID, "advisory.QueryByID", data.Log(query))

	var result struct {
		GetCity struct {
			Advisory Info `json:"advisory"`
		} `json:"getCity"`
	}
	if err := a.gql.Query(ctx, query, &result); err != nil {
		return Info{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.Advisory.ID == "" {
		return Info{}, ErrNotFound
	}

	return result.GetCity.Advisory, nil
}

// =============================================================================

func (a Advisory) add(ctx context.Context, traceID string, adv Info) (Info, error) {
	mutation, result := prepareAdd(adv)
	a.log.Printf("%s: %s: %s", traceID, "advisory.Add", data.Log(mutation))

	if err := a.gql.Query(ctx, mutation, &result); err != nil {
		return Info{}, errors.Wrap(err, "failed to add place")
	}

	if len(result.AddAdvisory.Advisory) != 1 {
		return Info{}, errors.New("advisory id not returned")
	}

	adv.ID = result.AddAdvisory.Advisory[0].ID
	return adv, nil
}

func (a Advisory) delete(ctx context.Context, traceID string, cityID string) error {
	adv, err := a.QueryByCity(ctx, traceID, cityID)
	if err != nil {
		return err
	}

	mutation, result := prepareDelete(adv.ID)
	a.log.Printf("%s: %s: %s", traceID, "advisory.Delete", data.Log(mutation))

	if err := a.gql.Query(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete advisory")
	}

	if result.DeleteAdvisory.NumUids != 1 {
		msg := fmt.Sprintf("failed to delete advisory: NumUids: %d  Msg: %s", result.DeleteAdvisory.NumUids, result.DeleteAdvisory.Msg)
		return errors.New(msg)
	}

	return nil
}

// =============================================================================

func prepareAdd(adv Info) (string, addResult) {
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
}`, adv.City.ID, adv.Continent, adv.Country, adv.CountryCode,
		adv.LastUpdated, adv.Message, adv.Score, adv.Source,
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
