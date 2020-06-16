// Package data contains the schema and models for data access.
package data

import (
	"net"
	"net/http"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// Set of error variables for CRUD operations.
var (
	ErrUserExists    = errors.New("user exists")
	ErrUserNotExists = errors.New("user does not exist")
	ErrCityExists    = errors.New("city exists")
	ErrPlaceExists   = errors.New("place exists")
)

// Not found errors.
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrCityNotFound     = errors.New("city not found")
	ErrPlaceNotFound    = errors.New("place not found")
	ErrAdvisoryNotFound = errors.New("advisory not found")
	ErrWeatherNotFound  = errors.New("weather not found")
)

// Schema error variables.
var (
	ErrNoSchemaExists = errors.New("no schema exists")
	ErrInvalidSchema  = errors.New("schema doesn't match")
)

// DBConfig represents comfiguration needed to support managing, mutating,
// and querying the database.
type DBConfig struct {
	URL            string
	AuthHeaderName string
	AuthToken      string
}

// DB provides support for query and mutation operations against the database.
type DB struct {
	Mutate mutate
	Query  query
}

// NewDB constructs a DB value for use to query and mutate the database.
func NewDB(dbConfig DBConfig) (*DB, error) {
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

	auth := graphql.WithAuth(dbConfig.AuthHeaderName, dbConfig.AuthToken)
	graphql := graphql.New(dbConfig.URL, &client, auth)
	query := newQuery(graphql)

	db := DB{
		Mutate: newMutate(graphql, query),
		Query:  newQuery(graphql),
	}

	return &db, nil
}

// query represents the set of queries that can be performed.
type query struct {
	graphql *graphql.GraphQL
}

// New constructs a Query value for use against the database.
func newQuery(graphql *graphql.GraphQL) query {
	return query{
		graphql: graphql,
	}
}

// Method set can be found in the set of files named query.

type mutate struct {
	graphql *graphql.GraphQL
	query   query
}

func newMutate(graphql *graphql.GraphQL, query query) mutate {
	return mutate{
		graphql: graphql,
		query:   newQuery(graphql),
	}
}

// Method set can be found in the set of files named mutate.
