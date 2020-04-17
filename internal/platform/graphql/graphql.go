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
	CmdAlter  = "alter"
	CmdMutate = "mutate"
	CmdSchema = "admin/schema"
	CmdQuery  = "query"
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

// Schema performs a schema operation against the configured server.
func (g *GraphQL) Schema(ctx context.Context, schemaString string, response interface{}) error {

	// Place the schema string into a reader for processing.
	reader := strings.NewReader(schemaString)

	// Make the http call to the server.
	return g.do(ctx, CmdSchema, reader, response)
}

// Query performs a basic query against the configured server.
func (g *GraphQL) Query(ctx context.Context, queryString string, response interface{}) error {
	return g.QueryWithVars(ctx, queryString, nil, response)
}

// QueryWithVars performs a query against the configured server with variable substituion.
func (g *GraphQL) QueryWithVars(ctx context.Context, queryString string, queryVars map[string]interface{}, response interface{}) error {

	// Prepare the request for the HTTP call.
	var body bytes.Buffer
	request := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     queryString,
		Variables: queryVars,
	}
	if err := json.NewEncoder(&body).Encode(request); err != nil {
		return fmt.Errorf("encoding error : %w", err)
	}

	// Make the http call to the server.
	return g.do(ctx, CmdQuery, &body, response)
}

// Error represents an error that can be returned from a graphql server.
type Error struct {
	Message string
}

// Error implements the error interface.
func (err *Error) Error() string {
	return "graphql: " + err.Message
}

func (g *GraphQL) do(ctx context.Context, command string, reader io.Reader, response interface{}) error {

	// Construct a request for the call.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.url+command, reader)
	if err != nil {
		return fmt.Errorf("create request error : %w", err)
	}

	// Prepare the header variables.
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make the call to the DB server.
	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("perform request error : %w", err)
	}
	defer resp.Body.Close()

	// Pull the entire result from the server.
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("copy error : %w", err)
	}

	// Define the structure of a result.
	type result struct {
		Data   interface{}
		Errors []Error
	}

	// Check we got an error on the call.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("operation error : status code : %s", resp.Status)
	}

	// Decode the result into our own data type.
	gr := result{
		Data: response,
	}
	if err := json.Unmarshal(data, &gr); err != nil {
		return fmt.Errorf("decoding error : %w : %s", err, string(data))
	}

	// If there is an error, just return the first one.
	if len(gr.Errors) > 0 {
		return fmt.Errorf("operation error : %w", &gr.Errors[0])
	}

	return nil
}
