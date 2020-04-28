package data

import (
	"github.com/dgraph-io/travel/internal/platform/graphql"
)

type delete struct {
	query   query
	graphql *graphql.GraphQL
}

// // Advisory will delete the current Advisory from the database.
// func (d *delete) Advisory(ctx context.Context) error {

// 	// Define a graphql mutation to update the city in the database with
// 	// the advisory and return the database generated id for the city.
// 	mutation := fmt.Sprintf(`
// mutation {
// 	updateCity(input: {
// 		filter: {
// 		  id: [%q]
// 		},
// 		set: {
// 			advisory: {
// 				continent: %q,
// 				country: %q,
// 				country_code: %q,
// 				last_updated: %q,
// 				message: %q,
// 				score: %f,
// 				source: %q
// 			}
// 		}
// 	})
// 	{
// 		city {
// 			id
// 		}
// 	}
// }`, cityID, advisory.Continent, advisory.Country, advisory.CountryCode,
// 		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source)

// 	return updCity(ctx, s.graphql, mutation)
// }
