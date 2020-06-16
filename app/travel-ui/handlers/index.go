package handlers

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/dgraph-io/travel/business/data"
	"github.com/pkg/errors"
)

type index struct {
	tmpl            *template.Template
	graphQLEndpoint string
	authHeaderName  string
	authToken       string
	mapsKey         string
}

func newIndex(dbConfig data.DBConfig, browserEndpoint string, mapsKey string) (*index, error) {
	rawTmpl, err := ioutil.ReadFile("assets/views/index.tmpl")
	if err != nil {
		return nil, errors.Wrap(err, "reading index page")
	}

	tmpl := template.New("index")
	if _, err := tmpl.Parse(string(rawTmpl)); err != nil {
		return nil, errors.Wrap(err, "creating template")
	}

	index := index{
		tmpl:            tmpl,
		graphQLEndpoint: browserEndpoint,
		authHeaderName:  dbConfig.AuthHeaderName,
		authToken:       dbConfig.AuthToken,
		mapsKey:         mapsKey,
	}

	return &index, nil
}

func (i *index) handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var markup bytes.Buffer
	vars := map[string]interface{}{
		"GraphQLEndpoint": i.graphQLEndpoint + "/graphql",
		"MapsKey":         i.mapsKey,
		"AuthHeaderName":  i.authHeaderName,
		"AuthToken":       i.authToken,
	}

	if err := i.tmpl.Execute(&markup, vars); err != nil {
		return errors.Wrap(err, "executing template")
	}

	io.Copy(w, &markup)
	return nil
}
