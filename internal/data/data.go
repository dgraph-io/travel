package data

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ardanlabs/graphql"
)

// Errors represents a group of errors.
type Errors []error

// Error implements the error interface.
func (errors Errors) Error() string {
	var sb strings.Builder
	for _, err := range errors {
		sb.WriteString(err.Error() + "\n")
	}
	return sb.String()
}

// Dgraph represents the IP and Ports we need to talk to the
// server for the different functions we need to perform.
type Dgraph struct {
	URL            string
	AuthHeaderName string
	AuthToken      string
}

// DB provides support for storing data inside of Dgraph.
type DB struct {
	Schema schema
	Mutate mutate
	Query  query
}

// NewDB constructs a data value for use to store data inside
// of the Dgraph database.
func NewDB(dgraph Dgraph) (*DB, error) {
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

	auth := graphql.WithAuth(dgraph.AuthHeaderName, dgraph.AuthToken)
	graphql := graphql.New(dgraph.URL, &client, auth)

	db := DB{
		Schema: schema{graphql: graphql},
		Mutate: mutate{graphql: graphql, query: query{graphql: graphql}},
		Query:  query{graphql: graphql},
	}

	return &db, nil
}
