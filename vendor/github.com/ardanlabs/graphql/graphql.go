// Package graphql provides support for executing mutations and queries against a
// database using GraphQL. It was designed specifically for working with
// [Dgraph](https://dgraph.io/) and has some Dgraph specific support.
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

// This provides a default client configuration but it is recommended
// this is replaced by the user using the WithClient function.
var defaultClient = http.Client{
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

// GraphQL represents a system that can accept a graphql query.
type GraphQL struct {
	url     string
	headers map[string]string
	client  *http.Client
}

// New constructs a GraphQL for use to making queries agains a specified host.
// The url is the fully qualifying URL without the /graphql path.
func New(url string, options ...func(gql *GraphQL)) *GraphQL {
	gql := GraphQL{
		url:     strings.TrimRight(url, "/") + "/",
		headers: make(map[string]string),
		client:  &defaultClient,
	}
	for _, option := range options {
		option(&gql)
	}
	return &gql
}

// WithClient adds a custom client for processing requests.
func WithClient(client *http.Client) func(gql *GraphQL) {
	return func(gql *GraphQL) {
		gql.client = client
	}
}

// WithHeader adds a key value pair to the header for requests.
func WithHeader(key string, value string) func(gql *GraphQL) {
	return func(gql *GraphQL) {
		gql.headers[key] = value
	}
}

// WithVariable allows for the submission of variables to the query.
func WithVariable(key string, value interface{}) func(m map[string]interface{}) {
	return func(m map[string]interface{}) {
		m[key] = value
	}
}

// Query performs a GraphQL query against the configured server.
func (g *GraphQL) Query(ctx context.Context, queryString string, response interface{}, variables ...func(m map[string]interface{})) error {
	var queryVars map[string]interface{}
	if len(variables) > 0 {
		queryVars = make(map[string]interface{})
		for _, variable := range variables {
			variable(queryVars)
		}
	}
	return g.query(ctx, "graphql", queryString, queryVars, response)
}

// QueryEndpoint performs a GraphQL query against the configured server at the
// specified endpoint from the base URL.
func (g *GraphQL) QueryEndpoint(ctx context.Context, endpoint string, queryString string, response interface{}, variables ...func(m map[string]interface{})) error {
	var queryVars map[string]interface{}
	if len(variables) > 0 {
		queryVars = make(map[string]interface{})
		for _, variable := range variables {
			variable(queryVars)
		}
	}
	return g.query(ctx, endpoint, queryString, queryVars, response)
}

// Do performs a request against the endpoint and doesn't prepare the request
// as a GraphQL request.
func (g *GraphQL) Do(ctx context.Context, endpoint string, queryString string, response interface{}) error {
	return g.do(ctx, endpoint, strings.NewReader(queryString), response)
}

// query performs a query against the configured server with variable substituion.
func (g *GraphQL) query(ctx context.Context, endpoint string, queryString string, queryVars map[string]interface{}, response interface{}) error {
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

	return g.do(ctx, endpoint, &b, response)
}

// Do provides the mechanics of handling a GraphQL request and response.
func (g *GraphQL) do(ctx context.Context, endpoint string, r io.Reader, response interface{}) error {

	// Want to capture the query being executed for development level logging
	// below. The TeeReader will write the query to this buffer when the request
	// reads the query for the http call.
	var query bytes.Buffer
	r = io.TeeReader(r, &query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.url+endpoint, r)
	if err != nil {
		return fmt.Errorf("graphql create request error: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, value := range g.headers {
		req.Header.Set(key, value)
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

	// This is for development level logging if running into a problem.
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
