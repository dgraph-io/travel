package handlers

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dgraph-io/travel/internal/mid"
	"github.com/dgraph-io/travel/internal/platform/web"
)

// UI constructs an http.Handler with all application routes defined.
func UI(build string, shutdown chan os.Signal, log *log.Logger) (*web.App, error) {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Load homepage from html in index.go
	index := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		io.WriteString(w, indexHTML)
		return nil
	}
	app.Handle("GET", "/", index)

	// Load the data to be graphed.
	data := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		io.WriteString(w, retrieveData())
		return nil
	}
	app.Handle("GET", "/data", data)

	return app, nil
}

// retrieveData
func retrieveData() string {
	return doc
}

var doc = `{
	"nodes": [
	  {"id": "Sydney", "group": 0, "radius": 15, "color": "blue"},
	  {"id": "Advisory", "group": 1, "radius": 10, "color": "red"},
	  {"id": "Weather", "group": 2, "radius": 10, "color": "orange"},
	  {"id": "Places", "group": 3, "radius": 10, "color": "purple"},
	  {"id": "Bill_Bar_And_Grill", "group": 3, "radius": 8, "color": "purple"},
	  {"id": "Ale_Raw_Bar", "group": 3, "radius": 8, "color": "purple"}
	],
	"links": [
	  {"source": "Sydney", "target": "Advisory", "width": 4},
	  {"source": "Sydney", "target": "Weather", "width": 4},
	  {"source": "Sydney", "target": "Places", "width": 4},
	  {"source": "Places", "target": "Bill_Bar_And_Grill", "width": 2},
	  {"source": "Places", "target": "Ale_Raw_Bar", "width": 2}
	]
}`
