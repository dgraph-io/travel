// Package graphql provides support for executing mutations and queries against a
// database using GraphQL. It was designed specifically for working with [Dgraph](https://dgraph.io/).
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// These commands represents the set of know graphql commands.
const (
	CmdAdmin   = "admin"
	CmdQuery   = "graphql"
	CmdQueryPM = "query"
)

// GraphQL represents a system that can accept a graphql query.
type GraphQL struct {
	url            string
	authHeaderName string
	authToken      string
	client         *http.Client
}

// New constructs a GraphQL for use to making queries agains a specified host.
// The url is the fully qualifying URL without the /graphql path.
func New(url string, client *http.Client, options ...func(gql *GraphQL)) *GraphQL {
	gql := GraphQL{
		url:    strings.TrimRight(url, "/") + "/",
		client: client,
	}
	for _, option := range options {
		option(&gql)
	}
	return &gql
}

// WithAuth adds authentication parameters to the graphql client.
func WithAuth(authHeaderName string, authToken string) func(gql *GraphQL) {
	return func(gql *GraphQL) {
		gql.authHeaderName = authHeaderName
		gql.authToken = authToken
	}
}

// Query performs a GraphQL query against the configured server.
func (g *GraphQL) Query(ctx context.Context, queryString string, response interface{}) error {
	return g.QueryWithVars(ctx, CmdQuery, queryString, nil, response)
}

// QueryPM performs a GraphQL+- query against the configured Dgraph server.
func (g *GraphQL) QueryPM(ctx context.Context, queryString string, response interface{}) error {
	return g.QueryWithVars(ctx, CmdQueryPM, queryString, nil, response)
}

// QueryWithVars performs a query against the configured server with variable substituion.
func (g *GraphQL) QueryWithVars(ctx context.Context, command string, queryString string, queryVars map[string]interface{}, response interface{}) error {
	request := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     queryString,
		Variables: queryVars,
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return fmt.Errorf("graphql encoding error: %w", err)
	}

	return g.Do(ctx, command, &b, response)
}

// Do provides the mechanics of handling a GraphQL request and response.
func (g *GraphQL) Do(ctx context.Context, command string, r io.Reader, response interface{}) error {

	// Want to capture the query being executed for logging.
	// The TeeReader will write the query to this buffer when
	// the request reads the query for the http call.
	var query bytes.Buffer
	r = io.TeeReader(r, &query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.url+command, r)
	if err != nil {
		return fmt.Errorf("graphql create request error: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if g.authToken != "" {
		req.Header.Set(g.authHeaderName, g.authToken)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("graphql request error: %w", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("graphql copy error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("graphql op error: status code: %s", resp.Status)
	}

	// fmt.Println("*****graphql*******>\n", query.String(), "\n", string(data))

	result := struct {
		Data   interface{}
		Errors []struct {
			Message string
		}
	}{
		Data: response,
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("graphql decoding error: %w response: %s", err, string(data))
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("graphql op error:\nquery:\n%sgraphql error:\n%s", query.String(), result.Errors[0].Message)
	}

	return nil
}
