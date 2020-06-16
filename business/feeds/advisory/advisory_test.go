package advisory_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgraph-io/travel/business/feeds/advisory"
	"github.com/dgraph-io/travel/foundation/tests"
	"github.com/google/go-cmp/cmp"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

// TestAdvisory validates searches can be conducted against www.travel-advisory.info.
func TestAdvisory(t *testing.T) {
	t.Log("Given the need to retreve an advisory.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single city.", testID)
		{
			server := mockServer()
			t.Cleanup(server.Close)

			countryCode := "AU"
			ctx := context.Background()

			found, err := advisory.Search(ctx, server.URL, countryCode)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to search for an advisory : %v", failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to search for an advisory.", success, testID)

			advisory := advisory.Advisory{
				Country:     "Australia",
				CountryCode: "AU",
				Continent:   "OC",
				Score:       2.8,
				LastUpdated: "2020-05-03 07:22:19",
				Message:     "none at this time",
				Source:      "https://www.travel-advisory.info/australia",
			}

			if diff := cmp.Diff(found, advisory); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the expected advisory. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the expected advisory.", tests.Success, testID)
		}
	}
}

func mockServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		io.WriteString(w, result)
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

var result = `
{
	"api_status":{
	   "request":{
		  "item":"au"
	   },
	   "reply":{
		  "cache":"cached",
		  "code":200,
		  "status":"ok",
		  "note":"The api works, we could match requested country code.",
		  "count":1
	   }
	},
	"data":{
	   "AU":{
		  "iso_alpha2":"AU",
		  "name":"Australia",
		  "continent":"OC",
		  "advisory":{
			 "score":2.79999999999999982236431605997495353221893310546875,
			 "sources_active":6,
			 "message":"none at this time",
			 "updated":"2020-05-03 07:22:19",
			 "source":"https:\/\/www.travel-advisory.info\/australia"
		  }
	   }
	}
 }`
