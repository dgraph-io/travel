/*
Package data is providing data support for storing and querying
data to and from the database.

For quick reference links:
https://godoc.org/googlemaps.github.io/maps#NearbySearchRequest
https://github.com/dgraph-io/dgo/blob/master/upsert_test.go
https://dgraph.io/docs/mutations/#upsert-block
https://godoc.org/github.com/dgraph-io/dgo#Txn.Do

Query:
{
	sydneyUid(func: uid(0x1)) {
		uid
		city_name
		lat
		lng
		weather {
			weather_id
			city_name
			wind_direction
		}
		places {
			place_id
      		city_name
      		name
    	}
	}
}
*/
package data
