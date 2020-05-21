package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/pkg/errors"
)

type index struct {
	tmpl            *template.Template
	graphQLEndpoint string
}

func newIndex(dgraph data.Dgraph) (*index, error) {
	data, err := ioutil.ReadFile("assets/views/index.tmpl")
	if err != nil {
		return nil, errors.Wrap(err, "reading index page")
	}

	tmpl := template.New("index")
	if _, err := tmpl.Parse(string(data)); err != nil {
		return nil, errors.Wrap(err, "creating template")
	}

	index := index{
		graphQLEndpoint: fmt.Sprintf("%s://%s/graphql", dgraph.Protocol, dgraph.APIHostOutside),
		tmpl:            tmpl,
	}

	return &index, nil
}

func (i *index) handler(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var markup bytes.Buffer
	vars := map[string]interface{}{"GraphQLEndpoint": i.graphQLEndpoint}

	if err := i.tmpl.Execute(&markup, vars); err != nil {
		return errors.Wrap(err, "executing template")
	}

	io.Copy(w, &markup)
	return nil
}

type fetch struct {
	dgraph data.Dgraph
}

func (f fetch) data(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	cityName := "sydney"

	db, err := data.NewDB(f.dgraph)
	if err != nil {
		return errors.Wrap(err, "new db")
	}

	city, err := db.Query.CityByName(context.Background(), cityName)
	if err != nil {
		return errors.Wrap(err, "query city")
	}

	places, err := db.Query.Places(context.Background(), city.ID)
	if err != nil {
		return errors.Wrap(err, "query places")
	}

	out, err := marshalCity(cityName, places)
	if err != nil {
		return errors.Wrap(err, "marshal city")
	}

	io.WriteString(w, out)
	return nil
}

type node struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Group  int    `json:"group"`
	Radius int    `json:"radius"`
	Color  string `json:"color"`
}

type link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Width  int    `json:"width"`
}

type doc struct {
	Nodes []node `json:"nodes"`
	Links []link `json:"links"`
}

func marshalCity(cityName string, places []data.Place) (string, error) {
	d := doc{
		Nodes: []node{
			{cityName, "city", 0, 20, "blue"},
			{"Advisory", "advisory", 1, 15, "red"},
			{"Weather", "weather", 2, 15, "orange"},
			{"Places", "places", 3, 15, "purple"},
		},
		Links: []link{
			{cityName, "Advisory", 5},
			{cityName, "Weather", 5},
			{cityName, "Places", 5},
		},
	}

	for _, place := range places {
		d.Nodes = append(d.Nodes, node{place.Name, "place", 3, 8, "purple"})
		d.Links = append(d.Links, link{"Places", place.Name, 2})
	}

	data, err := json.Marshal(d)
	if err != nil {
		return "", errors.Wrap(err, "marshal data")
	}

	return string(data), nil
}
