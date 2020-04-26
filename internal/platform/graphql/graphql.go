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
	cmdAlter   = "alter"
	cmdSchema  = "admin/schema"
	cmdQuery   = "graphql"
	cmdQueryPM = "query"
)

// GraphQL represents a system that can accept a graphql query.
type GraphQL struct {
	url    string
	client *http.Client
}

// New constructs a GraphQL for use to making queries agains a
// specified host. The apiHost is just the IP:Port of the
// Dgraph API endpoint.
func New(apiHost string, client *http.Client) *GraphQL {
	return &GraphQL{
		url:    "http://" + apiHost + "/",
		client: client,
	}
}

// CreateSchema performs a schema operation against the configured server.
func (g *GraphQL) CreateSchema(ctx context.Context, schemaString string, response interface{}) error {

	// Place the schema string into a buffer for processing.
	r := strings.NewReader(schemaString)

	// Make the http call to the server.
	return g.do(ctx, cmdSchema, r, response)
}

// Mutate performs a mutation operation against the configured server.
func (g *GraphQL) Mutate(ctx context.Context, mutationString string, response interface{}) error {
	return g.QueryWithVars(ctx, cmdQuery, mutationString, nil, response)
}

// Query performs a GraphQL query against the configured server.
func (g *GraphQL) Query(ctx context.Context, queryString string, response interface{}) error {
	return g.QueryWithVars(ctx, cmdQuery, queryString, nil, response)
}

// QueryPM performs a GraphQL+- query against the configured server.
func (g *GraphQL) QueryPM(ctx context.Context, queryString string, response interface{}) error {
	return g.QueryWithVars(ctx, cmdQueryPM, queryString, nil, response)
}

// QueryWithVars performs a query against the configured server with variable substituion.
func (g *GraphQL) QueryWithVars(ctx context.Context, command string, queryString string, queryVars map[string]interface{}, response interface{}) error {

	// Prepare the request for the HTTP call.
	var b bytes.Buffer
	request := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     queryString,
		Variables: queryVars,
	}
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return fmt.Errorf("graphql encoding error: %w", err)
	}

	// Make the http call to the server.
	return g.do(ctx, command, &b, response)
}

// do provides the mechanics of handling a GraphQL request and response.
func (g *GraphQL) do(ctx context.Context, command string, r io.Reader, response interface{}) error {

	// Want to capture the query being executed for logging.
	// The TeeReader will write the query to this buffer when
	// the request reads the query for the http call.
	var query bytes.Buffer
	r = io.TeeReader(r, &query)

	// Construct a request for the call.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.url+command, r)
	if err != nil {
		return fmt.Errorf("graphql create request error: %w", err)
	}

	// Prepare the header variables.
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make the call to the DB server.
	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("graphql request error: %w", err)
	}
	defer resp.Body.Close()

	// Pull the entire result from the server.
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("graphql copy error: %w", err)
	}

	// Define the structure of a result.
	type result struct {
		Data   interface{}
		Errors []struct {
			Message string
		}
	}

	// Check we got an error on the call.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("graphql op error: status code: %s", resp.Status)
	}

	// Decode the result into our own data type.
	gr := result{
		Data: response,
	}
	if err := json.Unmarshal(data, &gr); err != nil {
		return fmt.Errorf("graphql decoding error: %w response: %s", err, string(data))
	}

	// If there is an error, just return the first one.
	if len(gr.Errors) > 0 {
		return fmt.Errorf("graphql op error:\nquery:\n%sgraphql error:\n%s", query.String(), gr.Errors[0].Message)
	}

	return nil
}
