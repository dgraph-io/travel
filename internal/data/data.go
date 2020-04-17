package data

import (
	"net"
	"net/http"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// DB provides support for storing data inside of DGraph.
type DB struct {
	Schema schema
	Store  store
	Query  query
}

// NewDB constructs a data value for use to store data inside
// of the Dgraph database.
func NewDB(dbHost string, apiHost string) (*DB, error) {

	// Dial up an grpc connection to dgraph and construct
	// a dgraph client.
	conn, err := grpc.Dial(dbHost, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrapf(err, "dbHost[%s]", dbHost)
	}
	dgraph := dgo.NewDgraphClient(api.NewDgraphClient(conn))

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
		Store:  store{graphql: graphql, Dgraph: dgraph},
		Query:  query{graphql: graphql},
	}

	return &db, nil
}
