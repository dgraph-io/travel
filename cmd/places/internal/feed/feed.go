package feed

import (
	"context"
	"encoding/json"
	"fmt"

	"googlemaps.github.io/maps"
)

// Pull extracts the feed from the source.
func Pull(ctx context.Context, host string) error {
	mc, err := maps.NewClient(maps.WithAPIKey("AIzaSyBR0-ToiYlrhPlhidE7DA-Zx7EfE7FnUek"))
	if err != nil {
		return err
	}

	latLng := maps.LatLng{
		Lat: -33.865143,
		Lng: 151.209900,
	}
	nsr := maps.NearbySearchRequest{
		Location:  &latLng,
		Keyword:   "Sydney",
		PageToken: "",
		Radius:    5000,
	}
	resp, err := mc.NearbySearch(context.TODO(), &nsr)
	if err != nil {
		return err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	//	 {
	//     "geometry": {
	//         "location": {
	//             "lat": -33.8688197,
	//             "lng": 151.2092955
	//         },
	//         "location_type": "",
	//         "bounds": {
	//             "northeast": {
	//                 "lat": 0,
	//                 "lng": 0
	//             },
	//             "southwest": {
	//                 "lat": 0,
	//                 "lng": 0
	//             }
	//         },
	//         "viewport": {
	//             "northeast": {
	//                 "lat": -33.5781409,
	//                 "lng": 151.3430209
	//             },
	//             "southwest": {
	//                 "lat": -34.118347,
	//                 "lng": 150.5209286
	//             }
	//         },
	//         "types": null
	//     },
	//     "name": "Sydney",
	//     "icon": "https://maps.gstatic.com/mapfiles/place_api/icons/geocode-71.png",
	//     "place_id": "ChIJP3Sa8ziYEmsRUKgyFmh9AQM",
	//     "scope": "GOOGLE",
	//     "types": [
	//         "colloquial_area",
	//         "locality",
	//         "political"
	//     ],
	//     "photos": [
	//         {
	//             "photo_reference": "CmRaAAAAe0Qh5YfppyNenci7n5qbL8cBAY9tym9xkrFDhMBC6XzSO5dgcJQigzBv2V0WQfqP1-xNgx62oD8lLqJLxW4OonCnLIW5d4_LiteSBqc5-WnziOmiyw5BbspEMqciu5axEhAGZevIYmTdRvvQ1uSUi6LdGhQtaMoKtH7GHIemi25JhCTTUzQ8Jw",
	//             "height": 452,
	//             "width": 720,
	//             "html_attributions": [
	//                 "\u003ca href=\"https://maps.google.com/maps/contrib/115479276413292861472\"\u003eAshutosh Kumar\u003c/a\u003e"
	//             ]
	//         }
	//     ],
	//     "vicinity": "Sydney NSW, Australia",
	//     "id": "044785c67d3ee62545861361f8173af6c02f4fae"
	// }

	fmt.Println(string(data))

	// conn, err := grpc.Dial(host, grpc.WithInsecure())
	// if err != nil {
	// 	return err
	// }

	// client := dgo.NewDgraphClient(
	// 	api.NewDgraphClient(conn),
	// )

	// txn := client.NewTxn()

	// mut := api.Mutation{
	// 	SetJson: data,
	// }
	// if _, err := txn.Mutate(ctx, &mut); err != nil {
	// 	txn.Discard(ctx)
	// 	return err
	// }

	// if err := txn.Commit(ctx); err != nil {
	// 	return nil
	// }

	return nil
}
