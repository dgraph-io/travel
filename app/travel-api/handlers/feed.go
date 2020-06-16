package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/loader"
	"github.com/dgraph-io/travel/foundation/web"
	"github.com/pkg/errors"
)

type feed struct {
	log          *log.Logger
	dbConfig     data.DBConfig
	loaderConfig loader.Config
}

func (f *feed) upload(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var request data.UploadFeedRequest
	if err := web.Decode(r, &request); err != nil {
		return errors.Wrap(err, "decoding request")
	}

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	go func() {
		search := loader.Search{
			CityName:    request.CityName,
			CountryCode: request.CountryCode,
			Lat:         request.Lat,
			Lng:         request.Lng,
		}
		if err := loader.UpdateData(f.log, f.dbConfig, f.loaderConfig, search); err != nil {
			log.Printf("%s : (%d) : %s %s -> %s (%s) : ERROR : %v",
				v.TraceID, v.StatusCode,
				r.Method, r.URL.Path,
				r.RemoteAddr, time.Since(v.Now), err,
			)
			return
		}
		log.Printf("%s : (%d) : %s %s -> %s (%s)",
			v.TraceID, v.StatusCode,
			r.Method, r.URL.Path,
			r.RemoteAddr, time.Since(v.Now),
		)
	}()

	resp := data.UploadFeedResponse{
		CountryCode: request.CountryCode,
		CityName:    request.CityName,
		Lat:         request.Lat,
		Lng:         request.Lng,
		Message:     fmt.Sprintf("Uploading data for city %q [%f,%f] in country %q", request.CityName, request.Lat, request.Lng, request.CountryCode),
	}
	return web.Respond(ctx, w, resp, http.StatusOK)
}
