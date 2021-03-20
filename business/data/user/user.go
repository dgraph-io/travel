// Package user provides support for managing users in the database.
package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/dgraph-io/travel/business/data"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotExists = errors.New("user does not exist")
	ErrExists    = errors.New("user exists")
	ErrNotFound  = errors.New("user not found")
)

// Store manages the set of API's for user access.
type Store struct {
	log *log.Logger
	gql *graphql.GraphQL
}

// NewStore constructs a user store for api access.
func NewStore(log *log.Logger, gql *graphql.GraphQL) Store {
	return Store{
		log: log,
		gql: gql,
	}
}

// Add adds a new user to the database. If the user already exists
// this function will fail but the found user is returned. If the user is
// being added, the user with the id from the database is returned.
func (s Store) Add(ctx context.Context, traceID string, nu NewUser, now time.Time) (User, error) {
	if usr, err := s.QueryByEmail(ctx, traceID, nu.Email); err == nil {
		return usr, ErrExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "generating password hash")
	}

	usr := User{
		Name:         nu.Name,
		Email:        nu.Email,
		Role:         nu.Role,
		PasswordHash: string(hash),
		DateCreated:  now,
		DateUpdated:  now,
	}

	return s.add(ctx, traceID, usr)
}

// Update updates a user in the database by its ID. If the user doesn't
// already exist, this function will fail.
func (s Store) Update(ctx context.Context, traceID string, usr User) error {
	if usr.ID == "" {
		return errors.New("user missing id")
	}

	if _, err := s.QueryByID(ctx, traceID, usr.ID); err != nil {
		return ErrNotExists
	}

	return s.update(ctx, traceID, usr)
}

// Delete removes a user from the database by its ID. If the user doesn't
// already exist, this function will fail.
func (s Store) Delete(ctx context.Context, traceID string, userID string) error {
	if userID == "" {
		return errors.New("missing user id")
	}

	if _, err := s.QueryByID(ctx, traceID, userID); err != nil {
		return ErrNotExists
	}

	return s.delete(ctx, traceID, userID)
}

// QueryByID returns the specified user from the database by the city id.
func (s Store) QueryByID(ctx context.Context, traceID string, userID string) (User, error) {
	query := fmt.Sprintf(`
query {
	getUser(id: %q) {
		id
		name
		email
		role
		password_hash
		date_created
		date_updated
	}
}`, userID)

	s.log.Printf("%s: %s: %s", traceID, "user.QueryByID", data.Log(query))

	var result struct {
		GetUser User `json:"getUser"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if result.GetUser.ID == "" {
		return User{}, ErrNotFound
	}

	return result.GetUser, nil
}

// QueryByEmail returns the specified user from the database by email.
func (s Store) QueryByEmail(ctx context.Context, traceID string, email string) (User, error) {
	query := fmt.Sprintf(`
query {
	queryUser(filter: { email: { eq: %q } }) {
		id
		name
		email
		role
		password_hash
		date_created
		date_updated
	}
}`, email)

	s.log.Printf("%s: %s: %s", traceID, "user.QueryByEmail", data.Log(query))

	var result struct {
		QueryUser []User `json:"queryUser"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryUser) != 1 {
		return User{}, ErrNotFound
	}

	return result.QueryUser[0], nil
}

// =============================================================================

func (s Store) add(ctx context.Context, traceID string, usr User) (User, error) {
	var result id
	mutation := fmt.Sprintf(`
	mutation {
		resp: addUser(input: [{
			name: %q
			email: %q
			role: %s
			password_hash: %q
			date_created: %q
			date_updated: %q
		}])
		%s
	}`, usr.Name, usr.Email, usr.Role, usr.PasswordHash,
		usr.DateCreated.UTC().Format(time.RFC3339),
		usr.DateUpdated.UTC().Format(time.RFC3339),
		result.document())

	s.log.Printf("%s: %s: %s", traceID, "user.Add", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, &result); err != nil {
		return User{}, errors.Wrap(err, "failed to add user")
	}

	if len(result.Resp.Entities) != 1 {
		return User{}, errors.New("user id not returned")
	}

	usr.ID = result.Resp.Entities[0].ID
	return usr, nil
}

func (s Store) update(ctx context.Context, traceID string, usr User) error {
	var result result
	mutation := fmt.Sprintf(`
	mutation {
		resp: updateUser(input: {
			filter: {
			id: [%q]
			},
			set: {
				name: %q
				email: %q
				role: %s
				password_hash: %q
				date_created: %q
				date_updated: %q
			}
		})
		%s
	}`, usr.ID, usr.Name, usr.Email, usr.Role, usr.PasswordHash,
		usr.DateCreated.UTC().Format(time.RFC3339),
		usr.DateUpdated.UTC().Format(time.RFC3339),
		result.document())

	s.log.Printf("%s: %s: %s", traceID, "user.Update", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if result.Resp.NumUids != 1 {
		msg := fmt.Sprintf("failed to update user: NumUids: %d  Msg: %s", result.Resp.NumUids, result.Resp.Msg)
		return errors.New(msg)
	}

	return nil
}

func (s Store) delete(ctx context.Context, traceID string, userID string) error {
	var result result
	mutation := fmt.Sprintf(`
	mutation {
		resp: deleteUser(filter: { id: [%q] })
		%s
	}`, userID, result.document())

	s.log.Printf("%s: %s: %s", traceID, "user.Delete", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if result.Resp.NumUids != 0 {
		msg := fmt.Sprintf("failed to delete user: NumUids: %d  Msg: %s", result.Resp.NumUids, result.Resp.Msg)
		return errors.New(msg)
	}

	return nil
}
