// Package data contains the schema and models for data access.
package data

import (
	"net"
	"net/http"
	"time"

	"github.com/ardanlabs/graphql"
)

// GraphQLConfig represents comfiguration needed to support managing, mutating,
// and querying the database.
type GraphQLConfig struct {
	URL            string
	AuthHeaderName string
	AuthToken      string
}

// NewGraphQL constructs a graphql value for use to access the databse.
func NewGraphQL(gqlConfig GraphQLConfig) *graphql.GraphQL {
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	auth := graphql.WithAuth(gqlConfig.AuthHeaderName, gqlConfig.AuthToken)
	graphql := graphql.New(gqlConfig.URL, &client, auth)

	return graphql
}
