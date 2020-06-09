package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type index struct {
	tmpl            *template.Template
	graphQLEndpoint string
	cities          []string
	mapsKey         string
}

func newIndex(browserEndpoint string, cities []string, mapsKey string) (*index, error) {
	data, err := ioutil.ReadFile("assets/views/index.tmpl")
	if err != nil {
		return nil, errors.Wrap(err, "reading index page")
	}

	tmpl := template.New("index")
	if _, err := tmpl.Parse(string(data)); err != nil {
		return nil, errors.Wrap(err, "creating template")
	}

	index := index{
		tmpl:            tmpl,
		graphQLEndpoint: fmt.Sprintf("%s/graphql", browserEndpoint),
		cities:          cities,
		mapsKey:         mapsKey,
	}

	return &index, nil
}

func (i *index) handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var markup bytes.Buffer
	vars := map[string]interface{}{
		"GraphQLEndpoint": i.graphQLEndpoint,
		"Cities":          i.cities,
		"MapsKey":         i.mapsKey,
	}

	if err := i.tmpl.Execute(&markup, vars); err != nil {
		return errors.Wrap(err, "executing template")
	}

	io.Copy(w, &markup)
	return nil
}
