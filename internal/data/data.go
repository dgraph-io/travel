package data

import (
	"net"
	"net/http"
	"time"

	"github.com/dgraph-io/travel/internal/platform/graphql"
)

// DB provides support for storing data inside of DGraph.
type DB struct {
	Schema schema
	Mutate mutate
	Query  query
}

// NewDB constructs a data value for use to store data inside
// of the Dgraph database.
func NewDB(apiHost string) (*DB, error) {

	// Construct a client for graphql calls.
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

	// Construct a graphql value for making queries.
	graphql := graphql.New(apiHost, &client)

	// Construct a data value for use.
	db := DB{
		Schema: schema{graphql: graphql},
		Mutate: mutate{graphql: graphql, query: query{graphql: graphql}},
		Query:  query{graphql: graphql},
	}

	return &db, nil
}
