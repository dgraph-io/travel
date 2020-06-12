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
		UserID:   request.UserID,
		CityName: request.CityName,
		Message:  fmt.Sprintf("Uploading data for city %q by user %q", request.CityName, request.UserID),
	}
	return web.Respond(ctx, w, resp, http.StatusOK)
}
