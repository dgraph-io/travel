// Package graphql provides client support for executing graphql requests
// against a host that supports the graphql protocol.
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

// This provides a default client configuration, but it's recommended
// this is replaced by the user with application specific settings using
// the WithClient function at the time a GraphQL is constructed.
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

// GraphQL represents a client that can execute graphql and raw requests
// against a host.
type GraphQL struct {
	url     string
	headers map[string]string
	client  *http.Client
	logFunc func(s string)
}

// New constructs a GraphQL that can be used to execute graphql and raw requests
// against the specified url. The url represents a fully qualified URL without
// the `graphql` endpoint attached. If `/graphql` is provided, it's trimmed off.
func New(url string, options ...func(gql *GraphQL)) *GraphQL {
	url = strings.TrimSuffix(url, "/graphql")
	url = strings.TrimSuffix(url, "/") + "/"

	gql := GraphQL{
		url:     url,
		headers: make(map[string]string),
		client:  &defaultClient,
	}

	for _, option := range options {
		option(&gql)
	}

	return &gql
}

// WithClient adds a custom client for processing requests. It's recommend
// to not use the default client and provide your own.
func WithClient(client *http.Client) func(gql *GraphQL) {
	return func(gql *GraphQL) {
		gql.client = client
	}
}

// WithLogging acceps a function for capturing raw execution messages for the
// purpose of application logging.
func WithLogging(logFunc func(s string)) func(gql *GraphQL) {
	return func(gql *GraphQL) {
		gql.logFunc = logFunc
	}
}

// WithHeader adds a key/value pair to the request header for all calls made to
// the host. This is for things like authentication or application specific needs.
// These headers are already included:
// "Cache-Control": "no-cache", "Content-Type": "application/json", "Accept": "application/json"
func WithHeader(key string, value string) func(gql *GraphQL) {
	return func(gql *GraphQL) {
		if key != "" {
			gql.headers[key] = value
		}
	}
}

// WithVariable allows for the submission of variables when executing graphql
// against the host for queries that supports variable substitution.
func WithVariable(key string, value interface{}) func(m map[string]interface{}) {
	return func(m map[string]interface{}) {
		m[key] = value
	}
}

// Execute performs a graphql request against the configured host on the
// url/graphql endpoint.
func (g *GraphQL) Execute(ctx context.Context, graphql string, response interface{}, variables ...func(m map[string]interface{})) error {
	var queryVars map[string]interface{}
	if len(variables) > 0 {
		queryVars = make(map[string]interface{})
		for _, variable := range variables {
			variable(queryVars)
		}
	}
	return g.query(ctx, "graphql", graphql, queryVars, response)
}

// ExecuteOnEndpoint performs a graphql request against the configured host on
// the specified url/endpoint
func (g *GraphQL) ExecuteOnEndpoint(ctx context.Context, endpoint string, graphql string, response interface{}, variables ...func(m map[string]interface{})) error {
	var queryVars map[string]interface{}
	if len(variables) > 0 {
		queryVars = make(map[string]interface{})
		for _, variable := range variables {
			variable(queryVars)
		}
	}
	return g.query(ctx, endpoint, graphql, queryVars, response)
}

// query prepares the graphql request by applying the graphql request document
// around the query and variables. Then executes the request against the
// configured url/endpoint.
func (g *GraphQL) query(ctx context.Context, endpoint string, graphql string, queryVars map[string]interface{}, response interface{}) error {
	request := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     graphql,
		Variables: queryVars,
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return fmt.Errorf("graphql encoding error: %w", err)
	}

	return g.RawRequest(ctx, endpoint, &b, response)
}

// RawRequest performs the actual execution of a request against the specified
// url/endpoint. Use this function only when the request doesn't require a
// graphql document wrapper.
func (g *GraphQL) RawRequest(ctx context.Context, endpoint string, r io.Reader, response interface{}) error {

	// Use the TeeReader to capture the request being sent. This is needed if the
	// requrest fails for the error being returned or for logging if a log
	// function is provided. The TeeReader will write the request to this buffer
	// during the http operation.
	var request bytes.Buffer
	r = io.TeeReader(r, &request)

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

	if g.logFunc != nil {
		g.logFunc(fmt.Sprintf("request:[%s] data:[%s]", request.String(), string(data)))
	}

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
		return fmt.Errorf("graphql op error: request:[%s] error:[%s]", request.String(), result.Errors[0].Message)
	}

	return nil
}
