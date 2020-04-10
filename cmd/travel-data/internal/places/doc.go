// Package places is providing support to query the Google maps places API
// and retrieve places for a specified city. Places also stores the results
// of these searches into Dgraph.
//
// For quick reference links:
// https://godoc.org/googlemaps.github.io/maps#NearbySearchRequest
// https://github.com/dgraph-io/dgo/blob/master/upsert_test.go
// https://dgraph.io/docs/mutations/#upsert-block
// https://godoc.org/github.com/dgraph-io/dgo#Txn.Do
//
// Query:
// {
// 	hasPlace(func: has(city_name)) {
// 	  city_name
// 	  lat
// 	  lng
// 	  has_place {
// 		uid
// 		place_name
// 		address
// 		lat
// 		lng
// 		location_type
// 		avg_user_rating
// 	  }
// 	}
// }
package places
