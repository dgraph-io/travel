package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

type store struct {
	query   query
	graphql *graphql.GraphQL
}

// City first checks to validate the specified city doesn't exists in
// the database. If it doesn't, then the city is added to the database.
// It will return a new City with the city ID from the database.
func (s *store) City(ctx context.Context, city City) (City, error) {

	// Validate the city doesn't already exist.
	exists, err := s.query.CityByName(ctx, city.Name)
	if err != nil && err != ErrCityNotFound {
		return City{}, errors.Wrap(err, "failed to create city")
	}

	if err == ErrCityNotFound {

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

		// addCity will return the new city id if the function does not fail.
		return addCity(ctx, s.graphql, mutation, city)
	}

	return exists, nil
}

// Advisory will add the specified Advisory into the database.
func (s *store) Advisory(ctx context.Context, cityID string, advisory Advisory) (Advisory, error) {

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
			advisory {
				id
			}
		}
	}
}`, cityID, advisory.Continent, advisory.Country, advisory.CountryCode,
		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source)

	id, err := updateCity(ctx, s.graphql, mutation)
	if err != nil {
		return Advisory{}, errors.Wrap(err, "failed to add advisory")
	}

	advisory.ID = id
	return advisory, nil
}

// Weather will add the specified Place into the database.
func (s *store) Weather(ctx context.Context, cityID string, weather Weather) (Weather, error) {

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
			weather {
				id
			}
		}
	}
}`, cityID, weather.CityName, weather.Desc, weather.FeelsLike, weather.Humidity,
		weather.Pressure, weather.Sunrise, weather.Sunset, weather.Temp,
		weather.MinTemp, weather.MaxTemp, weather.Visibility, weather.WindDirection,
		weather.WindSpeed)

	id, err := updateCity(ctx, s.graphql, mutation)
	if err != nil {
		return Weather{}, errors.Wrap(err, "failed to add weather")
	}

	weather.ID = id
	return weather, nil
}

// Places will add the specified Places into the database.
func (s *store) Places(ctx context.Context, cityID string, places []Place) error {

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

	if _, err := updateCity(ctx, s.graphql, mutation); err != nil {
		return errors.Wrap(err, "failed to add places")
	}

	return nil
}

// marshalPlaces takes a base graphql document and a collection of places
// to generate a graphql collection of palces.
func marshalPlaces(ctx context.Context, places []Place) string {

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
func addCity(ctx context.Context, graphql *graphql.GraphQL, mutation string, city City) (City, error) {
	var result struct {
		AddCity struct {
			City []struct {
				ID string `json:"id"`
			} `json:"city"`
		} `json:"addCity"`
	}

	if err := graphql.Mutate(ctx, mutation, &result); err != nil {
		return City{}, errors.Wrap(err, "failed to add city")
	}

	if len(result.AddCity.City) != 1 {
		return City{}, errors.New("city id not returned")
	}

	city.ID = result.AddCity.City[0].ID

	return city, nil
}

// updateCity perform the actual graphql call against the database.
func updateCity(ctx context.Context, graphql *graphql.GraphQL, mutation string) (string, error) {
	var result struct {
		UpdCity struct {
			City []struct {
				Advisory struct {
					ID string `json:"id"`
				} `json:"advisory"`
				Weather struct {
					ID string `json:"id"`
				} `json:"weather"`
			} `json:"city"`
		} `json:"updateCity"`
	}

	err := graphql.Mutate(ctx, mutation, &result)
	if err != nil {
		return "", errors.Wrap(err, "failed to update city")
	}

	if len(result.UpdCity.City) != 1 {
		return "", errors.New("no data returned")
	}

	if result.UpdCity.City[0].Advisory.ID != "" {
		return result.UpdCity.City[0].Advisory.ID, nil
	}

	if result.UpdCity.City[0].Weather.ID != "" {
		return result.UpdCity.City[0].Weather.ID, nil
	}

	return "", nil
}
