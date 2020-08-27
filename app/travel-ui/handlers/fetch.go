package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/AvraamMavridis/randomcolor"
	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/city"
	"github.com/dgraph-io/travel/business/data/place"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/pkg/errors"
)

type fetch struct {
	gqlConfig data.GraphQLConfig
}

func (f fetch) data(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	gql := data.NewGraphQL(f.gqlConfig)

	params := httptreemux.ContextParams(r.Context())
	city, err := city.QueryByName(context.Background(), gql, params["city"])
	if err != nil {
		return errors.Wrap(err, "query city")
	}

	places, err := place.QueryByCity(context.Background(), gql, city.ID)
	if err != nil {
		return errors.Wrap(err, "query places")
	}

	out, err := marshalCity(params["city"], places)
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

func marshalCity(cityName string, places []place.Place) (string, error) {

	// Need the unique set of categories.
	categories := make(map[string]string)
	for _, place := range places {
		categories[place.Category] = ""
	}

	d := doc{
		Nodes: []node{
			{cityName, "city", 0, 20, "blue"},
			{"Advisory", "advisory", 1, 15, "red"},
			{"Weather", "weather", 2, 15, "orange"},
		},
		Links: []link{
			{cityName, "Advisory", 5},
			{cityName, "Weather", 5},
		},
	}

	for category := range categories {
		colorString := randomcolor.GetRandomColorInHex()
		categories[category] = colorString
		d.Nodes = append(d.Nodes, node{category, "place", 3, 15, colorString})
		d.Links = append(d.Links, link{cityName, category, 2})
	}

	for _, place := range places {
		d.Nodes = append(d.Nodes, node{place.Name, place.Category, 3, 8, categories[place.Category]})
		d.Links = append(d.Links, link{place.Category, place.Name, 2})
	}

	data, err := json.Marshal(d)
	if err != nil {
		return "", errors.Wrap(err, "marshal data")
	}

	return string(data), nil
}
