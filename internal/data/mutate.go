package data

import (
	"context"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrUserExists    = errors.New("user exists")
	ErrUserNotExists = errors.New("user does not exist")
	ErrCityExists    = errors.New("city exists")
	ErrPlaceExists   = errors.New("place exists")
)

type mutate struct {
	query   query
	graphql *graphql.GraphQL
}

// AddUser adds a new user to the database. If the user already exists
// this function will fail but the found user is returned. If the user is
// being added, the user with the id from the database is returned.
func (m *mutate) AddUser(ctx context.Context, newUser NewUser, now time.Time) (User, error) {
	if user, err := m.query.UserByEmail(ctx, newUser.Email); err == nil {
		return user, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "generating password hash")
	}

	user := User{
		Name:         newUser.Name,
		Email:        newUser.Email,
		Roles:        newUser.Roles,
		PasswordHash: string(hash),
		DateCreated:  now,
		DateUpdated:  now,
	}

	user, err = mutUser.add(ctx, m.graphql, user)
	if err != nil {
		return User{}, errors.Wrap(err, "adding user to database")
	}

	return user, nil
}

// UpdateUser updates a user in the database by its ID. If the user doesn't
// already exist, this function will fail.
func (m *mutate) UpdateUser(ctx context.Context, user User) error {
	if _, err := m.query.User(ctx, user.ID); err != nil {
		return ErrUserNotExists
	}

	if err := mutUser.update(ctx, m.graphql, user); err != nil {
		return errors.Wrap(err, "updating user in database")
	}

	return nil
}

// DeleteUser removes a user from the database by its ID. If the user doesn't
// already exist, this function will fail.
func (m *mutate) DeleteUser(ctx context.Context, userID string) error {
	if _, err := m.query.User(ctx, userID); err != nil {
		return ErrUserNotExists
	}

	if err := mutUser.delete(ctx, m.graphql, userID); err != nil {
		return errors.Wrap(err, "deleting user in database")
	}

	return nil
}

// AddCity adds a new city to the database. If the city already exists
// this function will fail but the found city is returned. If the city is
// being added, the city with the id from the database is returned.
func (m *mutate) AddCity(ctx context.Context, city City) (City, error) {
	if city, err := m.query.CityByName(ctx, city.Name); err == nil {
		return city, ErrCityExists
	}

	city, err := mutCity.add(ctx, m.graphql, city)
	if err != nil {
		return City{}, errors.Wrap(err, "adding city to database")
	}

	return city, nil
}

// AddPlace adds a new place to the database. If the place already exists
// this function will fail but the found place is returned. If the city is
// being added, the city with the id from the database is returned.
func (m *mutate) AddPlace(ctx context.Context, place Place) (Place, error) {
	if place, err := m.query.PlaceByName(ctx, place.Name); err == nil {
		return place, ErrPlaceExists
	}

	place, err := mutPlace.add(ctx, m.graphql, place)
	if err != nil {
		return Place{}, errors.Wrap(err, "adding place to database")
	}

	if err := mutPlace.updateCity(ctx, m.graphql, place.CityID.ID, place.ID); err != nil {
		return Place{}, errors.Wrap(err, "adding place to city in database")
	}

	return place, nil
}

// ReplaceAdvisory replaces an advisory in the database and connects it
// to the specified city.
func (m *mutate) ReplaceAdvisory(ctx context.Context, cityID string, advisory Advisory) (Advisory, error) {
	if err := mutAdvisory.delete(ctx, m.query, m.graphql, cityID); err != nil {
		if err != ErrAdvisoryNotFound {
			return Advisory{}, errors.Wrap(err, "deleting advisory from database")
		}
	}

	advisory, err := mutAdvisory.add(ctx, m.graphql, advisory)
	if err != nil {
		return Advisory{}, errors.Wrap(err, "adding advisory to database")
	}

	if err := mutAdvisory.updateCity(ctx, m.graphql, cityID, advisory); err != nil {
		return Advisory{}, errors.Wrap(err, "replace advisory in city")
	}

	return advisory, nil
}

// ReplaceWeather replaces a weather in the database and connects it
// to the specified city.
func (m *mutate) ReplaceWeather(ctx context.Context, cityID string, weather Weather) (Weather, error) {
	if err := mutWeather.delete(ctx, m.query, m.graphql, cityID); err != nil {
		if err != ErrWeatherNotFound {
			return Weather{}, errors.Wrap(err, "deleting weather from database")
		}
	}

	weather, err := mutWeather.add(ctx, m.graphql, weather)
	if err != nil {
		return Weather{}, errors.Wrap(err, "adding weather to database")
	}

	if err := mutWeather.updateCity(ctx, m.graphql, cityID, weather); err != nil {
		return Weather{}, errors.Wrap(err, "replace weather in city")
	}

	return weather, nil
}
