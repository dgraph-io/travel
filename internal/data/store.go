package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/dgraph-io/travel/internal/feeds/advisory"
	"github.com/dgraph-io/travel/internal/feeds/places"
	"github.com/dgraph-io/travel/internal/feeds/weather"
	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type store struct {
	graphql *graphql.GraphQL
}

// City is used to identify if the specified city exists in
// the database. If it doesn't, then the city is added to the database.
// It will return a new City with the city ID from the database.
func (s *store) City(ctx context.Context, city places.City) (string, error) {

	// Define a graphql mutation to add the city to the database and return
	// the database generated id for the city.
	mutation := fmt.Sprintf(`
mutation {
	addCity(input: [
		{name: %q, lat: %f, lng: %f}
	])
	{
		city {
			id
		}
	}
}`, city.Name, city.Lat, city.Lng)

	// addCity will return the new city id if the function
	// does not fail.
	return addCity(ctx, s.graphql, mutation)
}

// Advisory will add the specified Advisory into the database.
func (s *store) Advisory(ctx context.Context, cityID string, advisory advisory.Advisory) error {

	// Define a graphql mutation to update the city in the database with
	// the advisory and return the database generated id for the city.
	mutation := fmt.Sprintf(`
mutation {
	updateCity(input: {
		filter: {
		  id: [%q]
		},
		set: {
			advisory: {
				continent: %q,
				country: %q,
				country_code: %q,
				last_updated: %q,
				message: %q,
				score: %f,
				source: %q
			}
		}
	})
	{
		city {
			id
		}
	}
}`, cityID, advisory.Continent, advisory.Country, advisory.CountryCode,
		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source)

	return updCity(ctx, s.graphql, mutation)
}

// Weather will add the specified Place into the database.
func (s *store) Weather(ctx context.Context, cityID string, weather weather.Weather) error {

	// Define a graphql mutation to update the city in the database with
	// the weather and return the database generated id for the city.
	mutation := fmt.Sprintf(`
mutation {
	updateCity(input: {
		filter: {
		  id: [%q]
		},
		set: {
			weather: {
				city_name: %q,
				description: %q,
				feels_like: %f,
				humidity: %d,
				pressure: %d,
				sunrise: %d,
				sunset: %d,
				temp: %f,
				temp_min: %f,
				temp_max: %f,
				visibility: %q,
				wind_direction: %d,
				wind_speed: %f
			}
		}
	})
	{
		city {
			id
		}
	}
}`, cityID, weather.CityName, weather.Desc, weather.FeelsLike, weather.Humidity,
		weather.Pressure, weather.Sunrise, weather.Sunset, weather.Temp,
		weather.MinTemp, weather.MaxTemp, weather.Visibility, weather.WindDirection,
		weather.WindSpeed)

	return updCity(ctx, s.graphql, mutation)
}

// Places will add the specified Places into the database.
func (s *store) Places(ctx context.Context, cityID string, places []places.Place) error {

	// Define a graphql mutation to update the city in the database with
	// the places and return the database generated id for the city.
	mutation := fmt.Sprintf(`
mutation {
	updateCity(input: {
		filter: {
		  id: [%q]
		},
		set: {
			places: %s
		}
	})
	{
		city {
			id
		}
	}
}`, cityID, marshalPlaces(ctx, places))

	return updCity(ctx, s.graphql, mutation)
}

// marshalPlaces takes a base graphql document and a collection of places
// to generate a graphql collection of palces.
func marshalPlaces(ctx context.Context, places []places.Place) string {

	// Define a graphql document for a place.
	doc := `
{
	address: %q,
	avg_user_rating: %f,
	city_name: %q,
	gmaps_url: %q,
	lat: %f,
	lng: %f,
	location_type: [%q],
	name: %q,
	no_user_rating: %d,
	place_id: %q,
	photo_id: %q
}`

	var b strings.Builder
	b.WriteString("[")
	for _, place := range places {
		for i := range place.LocationType {
			place.LocationType[i] = fmt.Sprintf(`"%s"`, place.LocationType[i])
		}
		b.WriteString(fmt.Sprintf(doc,
			place.Address, place.AvgUserRating, place.CityName, place.GmapsURL,
			place.Lat, place.Lng, strings.Join(place.LocationType, ","), place.Name,
			place.NumberOfRatings, place.PlaceID, place.PhotoReferenceID))
	}
	b.WriteString("]")
	return b.String()
}

// addCity perform the actual graphql call against the database.
func addCity(ctx context.Context, graphql *graphql.GraphQL, mutation string) (string, error) {

	// Construct a result value for the call.
	var result struct {
		AddCity struct {
			City []struct {
				ID string `json:"id"`
			} `json:"city"`
		} `json:"addCity"`
	}

	// Perform the mutation.
	if err := graphql.Mutate(ctx, mutation, &result); err != nil {
		return "", errors.Wrap(err, "failed to add city")
	}

	// Validate we got back the city id.
	if len(result.AddCity.City) != 1 {
		return "", errors.New("city id not returned")
	}

	return result.AddCity.City[0].ID, nil
}

// updCity perform the actual graphql call against the database.
func updCity(ctx context.Context, graphql *graphql.GraphQL, mutation string) error {

	// Construct a result value for the call.
	var result struct {
		UpdCity struct {
			City []struct {
				ID string `json:"id"`
			} `json:"city"`
		} `json:"updateCity"`
	}

	// Perform the mutation.
	err := graphql.Mutate(ctx, mutation, &result)
	if err != nil {
		return errors.Wrap(err, "failed to update city")
	}

	// Validate we got back the city id.
	if len(result.UpdCity.City) != 1 {
		return errors.New("city id not returned")
	}

	return nil
}
