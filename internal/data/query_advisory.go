package data

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// Advisory returns the specified advisory from the database by the city id.
func (q query) Advisory(ctx context.Context, cityID string) (Advisory, error) {
	query := fmt.Sprintf(`
query {
	getCity(id: %q) {
		advisory {
			id
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
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return Advisory{}, errors.Wrap(err, "query failed")
	}

	if result.GetCity.Advisory.ID == "" {
		return Advisory{}, ErrAdvisoryNotFound
	}

	return result.GetCity.Advisory, nil
}
