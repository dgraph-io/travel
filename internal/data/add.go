package data

import (
	"context"

	"github.com/dgraph-io/travel/internal/platform/graphql"
	"github.com/pkg/errors"
)

// Error variables to indicate entities exsit.
var (
	ErrCityExists  = errors.New("city exists")
	ErrPlaceExists = errors.New("place exists")
)

type add struct {
	query   query
	graphql *graphql.GraphQL
}

// City first checks to validate the specified city doesn't exists in
// the database. If it doesn't, then the city is added to the database.
// It will return a new City with the city ID from the database.
func (a *add) City(ctx context.Context, city City) (City, error) {
	if addCity.exists(ctx, a.query, city) {
		return City{}, ErrCityExists
	}

	city, err := addCity.add(ctx, a.graphql, city)
	if err != nil {
		return City{}, errors.New("adding city to database")
	}

	return city, nil
}

// Place adds a new place to the database and connects it to the specified
// city. If the place already exists (by name), the function will return
// an error ErrPlaceExists.
func (a *add) Place(ctx context.Context, cityID string, place Place) (Place, error) {
	if addPlace.exists(ctx, a.query, place) {
		return Place{}, ErrPlaceExists
	}

	place, err := addPlace.add(ctx, a.graphql, place)
	if err != nil {
		return Place{}, errors.New("adding place to database")
	}

	if err := addPlace.updateCity(ctx, a.graphql, cityID, place); err != nil {
		return Place{}, errors.Wrap(err, "adding place to city in database")
	}

	return place, nil
}

// // Advisory will add the specified Advisory into the database.
// func (s *store) Advisory(ctx context.Context, cityID string, advisory Advisory) (Advisory, error) {

// 	// Define a graphql mutation to update the city in the database with
// 	// the advisory and return the database generated id for the city.
// 	mutation := fmt.Sprintf(`
// mutation {
// 	updateCity(input: {
// 		filter: {
// 		  id: [%q]
// 		},
// 		set: {
// 			advisory: {
// 				continent: %q,
// 				country: %q,
// 				country_code: %q,
// 				last_updated: %q,
// 				message: %q,
// 				score: %f,
// 				source: %q
// 			}
// 		}
// 	})
// 	{
// 		city {
// 			advisory {
// 				id
// 			}
// 		}
// 	}
// }`, cityID, advisory.Continent, advisory.Country, advisory.CountryCode,
// 		advisory.LastUpdated, advisory.Message, advisory.Score, advisory.Source)

// 	// id, err := updateCity(ctx, s.graphql, mutation)
// 	// if err != nil {
// 	// 	return Advisory{}, errors.Wrap(err, "failed to add advisory")
// 	// }

// 	advisory.ID = id
// 	return advisory, nil
// }

// // Weather will add the specified Place into the database.
// func (s *store) Weather(ctx context.Context, cityID string, weather Weather) (Weather, error) {

// 	// Define a graphql mutation to update the city in the database with
// 	// the weather and return the database generated id for the city.
// 	mutation := fmt.Sprintf(`
// mutation {
// 	updateCity(input: {
// 		filter: {
// 		  id: [%q]
// 		},
// 		set: {
// 			weather: {
// 				city_name: %q,
// 				description: %q,
// 				feels_like: %f,
// 				humidity: %d,
// 				pressure: %d,
// 				sunrise: %d,
// 				sunset: %d,
// 				temp: %f,
// 				temp_min: %f,
// 				temp_max: %f,
// 				visibility: %q,
// 				wind_direction: %d,
// 				wind_speed: %f
// 			}
// 		}
// 	})
// 	{
// 		city {
// 			weather {
// 				id
// 			}
// 		}
// 	}
// }`, cityID, weather.CityName, weather.Desc, weather.FeelsLike, weather.Humidity,
// 		weather.Pressure, weather.Sunrise, weather.Sunset, weather.Temp,
// 		weather.MinTemp, weather.MaxTemp, weather.Visibility, weather.WindDirection,
// 		weather.WindSpeed)

// 	// id, err := updateCity(ctx, s.graphql, mutation)
// 	// if err != nil {
// 	// 	return Weather{}, errors.Wrap(err, "failed to add weather")
// 	// }

// 	// weather.ID = id
// 	return weather, nil
// }
