package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// Error represents an error that can be returned from a graphql server.
type Error struct {
	Message string
}

// Error implements the error interface.
func (err *Error) Error() string {
	return "graphql: " + err.Message
}

// GraphQL represents a system that can accept a graphql query.
type GraphQL struct {
	host   string
	client *http.Client
}

// New constructs a GraphQL for use to making queries agains a
// specified host.
func New(host string, client *http.Client) *GraphQL {
	return &GraphQL{
		host:   host,
		client: client,
	}
}

// Query performs a basic query against the configured server.
func (g *GraphQL) Query(ctx context.Context, queryString string, response interface{}) error {
	return g.query(ctx, queryString, nil, response)
}

// QueryWithVars performs a query against the configured server with variable substituion.
func (g *GraphQL) QueryWithVars(ctx context.Context, queryString string, queryVars map[string]interface{}, response interface{}) error {
	return g.query(ctx, queryString, queryVars, response)
}

// Query performs a query against the configured server.
func (g *GraphQL) query(ctx context.Context, queryString string, queryVars map[string]interface{}, response interface{}) error {

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
		return err
	}

	// Cosntruct a request for the call.
	req, err := http.NewRequest(http.MethodPost, g.host, &body)
	if err != nil {
		return err
	}

	// Prepare the header variables.
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make the call to the DB server.
	req = req.WithContext(ctx)
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Pull the result from the server.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return err
	}

	// Define the structure of a result.
	type result struct {
		Data   interface{}
		Errors []Error
	}

	// Decode the result into our own data type.
	gr := result{
		Data: response,
	}
	if err := json.NewDecoder(&buf).Decode(&gr); err != nil {
		return err
	}

	// If there is an error, just return the first one.
	if len(gr.Errors) > 0 {
		return &gr.Errors[0]
	}

	return nil
}
