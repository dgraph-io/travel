package handlers

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/dgraph-io/travel/business/data"
	"github.com/pkg/errors"
)

type indexGroup struct {
	tmpl            *template.Template
	graphQLEndpoint string
	authHeaderName  string
	authToken       string
	mapsKey         string
}

func newIndex(gqlConfig data.GraphQLConfig, browserEndpoint string, mapsKey string) (indexGroup, error) {
	index, err := os.Open("assets/views/index.tmpl")
	if err != nil {
		return indexGroup{}, errors.Wrap(err, "open index page")
	}
	defer index.Close()
	rawTmpl, err := io.ReadAll(index)
	if err != nil {
		return indexGroup{}, errors.Wrap(err, "reading index page")
	}

	tmpl := template.New("index")
	if _, err := tmpl.Parse(string(rawTmpl)); err != nil {
		return indexGroup{}, errors.Wrap(err, "creating template")
	}

	ig := indexGroup{
		tmpl:            tmpl,
		graphQLEndpoint: browserEndpoint,
		authHeaderName:  gqlConfig.AuthHeaderName,
		authToken:       gqlConfig.AuthToken,
		mapsKey:         mapsKey,
	}

	return ig, nil
}

func (ig *indexGroup) handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var markup bytes.Buffer
	vars := map[string]interface{}{
		"GraphQLEndpoint": ig.graphQLEndpoint + "/graphql",
		"MapsKey":         ig.mapsKey,
		"AuthHeaderName":  ig.authHeaderName,
		"AuthToken":       ig.authToken,
	}

	if err := ig.tmpl.Execute(&markup, vars); err != nil {
		return errors.Wrap(err, "executing template")
	}

	io.Copy(w, &markup)
	return nil
}
