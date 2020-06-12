package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/loader"
	"github.com/dgraph-io/travel/internal/platform/web"
	"github.com/pkg/errors"
)

type feed struct {
	keys loader.Keys
	url  loader.URL
}

func (l *feed) upload(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var request data.UploadFeedRequest
	if err := web.Decode(r, &request); err != nil {
		return errors.Wrap(err, "decoding request")
	}

	// Load the data here!!

	resp := data.UploadFeedResponse{
		CountryCode: request.CountryCode,
		CityName:    request.CityName,
		Lat:         request.Lat,
		Lng:         request.Lng,
		Message:     fmt.Sprintf("Uploading data for city %q [%f,%f] in country %q", request.CityName, request.Lat, request.Lng, request.CountryCode),
	}
	return web.Respond(ctx, w, resp, http.StatusOK)
}
