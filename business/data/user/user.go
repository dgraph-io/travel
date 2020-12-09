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

// User manages the set of API's for user access.
type User struct {
	log *log.Logger
	gql *graphql.GraphQL
}

// New constructs a User for api access.
func New(log *log.Logger, gql *graphql.GraphQL) User {
	return User{
		log: log,
		gql: gql,
	}
}

// Add adds a new user to the database. If the user already exists
// this function will fail but the found user is returned. If the user is
// being added, the user with the id from the database is returned.
func (u User) Add(ctx context.Context, traceID string, nu NewUser, now time.Time) (Info, error) {
	if usr, err := u.QueryByEmail(ctx, traceID, nu.Email); err == nil {
		return usr, ErrExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return Info{}, errors.Wrap(err, "generating password hash")
	}

	usr := Info{
		Name:         nu.Name,
		Email:        nu.Email,
		Role:         nu.Role,
		PasswordHash: string(hash),
		DateCreated:  now,
		DateUpdated:  now,
	}

	usr, err = u.add(ctx, traceID, usr)
	if err != nil {
		return Info{}, errors.Wrap(err, "adding user to database")
	}

	return usr, nil
}

// Update updates a user in the database by its ID. If the user doesn't
// already exist, this function will fail.
func (u User) Update(ctx context.Context, traceID string, usr Info) error {
	if _, err := u.QueryByID(ctx, traceID, usr.ID); err != nil {
		return ErrNotExists
	}

	if err := u.update(ctx, traceID, usr); err != nil {
		return errors.Wrap(err, "updating user in database")
	}

	return nil
}

// Delete removes a user from the database by its ID. If the user doesn't
// already exist, this function will fail.
func (u User) Delete(ctx context.Context, traceID string, userID string) error {
	if _, err := u.QueryByID(ctx, traceID, userID); err != nil {
		return ErrNotExists
	}

	if err := u.delete(ctx, traceID, userID); err != nil {
		return errors.Wrap(err, "deleting user in database")
	}

	return nil
}

// QueryByID returns the specified user from the database by the city id.
func (u User) QueryByID(ctx context.Context, traceID string, userID string) (Info, error) {
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

	u.log.Printf("%s: %s: %s", traceID, "user.QueryByID", data.Log(query))

	var result struct {
		GetUser Info `json:"getUser"`
	}
	if err := u.gql.Query(ctx, query, &result); err != nil {
		return Info{}, errors.Wrap(err, "query failed")
	}

	if result.GetUser.ID == "" {
		return Info{}, ErrNotFound
	}

	return result.GetUser, nil
}

// QueryByEmail returns the specified user from the database by email.
func (u User) QueryByEmail(ctx context.Context, traceID string, email string) (Info, error) {
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

	u.log.Printf("%s: %s: %s", traceID, "user.QueryByEmail", data.Log(query))

	var result struct {
		QueryUser []Info `json:"queryUser"`
	}
	if err := u.gql.Query(ctx, query, &result); err != nil {
		return Info{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryUser) != 1 {
		return Info{}, ErrNotFound
	}

	return result.QueryUser[0], nil
}

// =============================================================================

func (u User) add(ctx context.Context, traceID string, usr Info) (Info, error) {
	mutation, result := prepareAdd(usr)
	u.log.Printf("%s: %s: %s", traceID, "user.Add", data.Log(mutation))

	if err := u.gql.Query(ctx, mutation, &result); err != nil {
		return Info{}, errors.Wrap(err, "failed to add user")
	}

	if len(result.Resp.Entities) != 1 {
		return Info{}, errors.New("user id not returned")
	}

	usr.ID = result.Resp.Entities[0].ID
	return usr, nil
}

func (u User) update(ctx context.Context, traceID string, usr Info) error {
	if usr.ID == "" {
		return errors.New("user missing id")
	}

	mutation, result := prepareUpdate(usr)
	u.log.Printf("%s: %s: %s", traceID, "user.Update", data.Log(mutation))

	if err := u.gql.Query(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if result.Resp.NumUids != 1 {
		msg := fmt.Sprintf("failed to update user: NumUids: %d  Msg: %s", result.Resp.NumUids, result.Resp.Msg)
		return errors.New(msg)
	}

	return nil
}

func (u User) delete(ctx context.Context, traceID string, userID string) error {
	if userID == "" {
		return errors.New("missing user id")
	}

	mutation, result := prepareDelete(userID)
	u.log.Printf("%s: %s: %s", traceID, "user.Delete", data.Log(mutation))

	if err := u.gql.Query(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if result.Resp.NumUids != 0 {
		msg := fmt.Sprintf("failed to delete user: NumUids: %d  Msg: %s", result.Resp.NumUids, result.Resp.Msg)
		return errors.New(msg)
	}

	return nil
}

// =============================================================================

func prepareAdd(usr Info) (string, id) {
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

	return mutation, result
}

func prepareUpdate(usr Info) (string, result) {
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

	return mutation, result
}

func prepareDelete(userID string) (string, result) {
	var result result
	mutation := fmt.Sprintf(`
mutation {
	resp: deleteUser(filter: { id: [%q] })
	%s
}`, userID, result.document())

	return mutation, result
}
