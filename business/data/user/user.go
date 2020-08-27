// Package user provides support for managing users in the database.
package user

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotExists = errors.New("user does not exist")
	ErrExists    = errors.New("user exists")
	ErrNotFound  = errors.New("user not found")
)

// Add adds a new user to the database. If the user already exists
// this function will fail but the found user is returned. If the user is
// being added, the user with the id from the database is returned.
func Add(ctx context.Context, gql *graphql.GraphQL, nu NewUser, now time.Time) (User, error) {
	if u, err := QueryByEmail(ctx, gql, nu.Email); err == nil {
		return u, ErrExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "generating password hash")
	}

	u := User{
		Name:         nu.Name,
		Email:        nu.Email,
		Role:         nu.Role,
		PasswordHash: string(hash),
		DateCreated:  now,
		DateUpdated:  now,
	}

	u, err = add(ctx, gql, u)
	if err != nil {
		return User{}, errors.Wrap(err, "adding user to database")
	}

	return u, nil
}

// Update updates a user in the database by its ID. If the user doesn't
// already exist, this function will fail.
func Update(ctx context.Context, gql *graphql.GraphQL, u User) error {
	if _, err := QueryByID(ctx, gql, u.ID); err != nil {
		return ErrNotExists
	}

	if err := update(ctx, gql, u); err != nil {
		return errors.Wrap(err, "updating user in database")
	}

	return nil
}

// Delete removes a user from the database by its ID. If the user doesn't
// already exist, this function will fail.
func Delete(ctx context.Context, gql *graphql.GraphQL, userID string) error {
	if _, err := QueryByID(ctx, gql, userID); err != nil {
		return ErrNotExists
	}

	if err := delete(ctx, gql, userID); err != nil {
		return errors.Wrap(err, "deleting user in database")
	}

	return nil
}

// QueryByID returns the specified user from the database by the city id.
func QueryByID(ctx context.Context, gql *graphql.GraphQL, userID string) (User, error) {
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

	var result struct {
		GetUser User `json:"getUser"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if result.GetUser.ID == "" {
		return User{}, ErrNotFound
	}

	return result.GetUser, nil
}

// QueryByEmail returns the specified user from the database by email.
func QueryByEmail(ctx context.Context, gql *graphql.GraphQL, email string) (User, error) {
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

	var result struct {
		QueryUser []User `json:"queryUser"`
	}
	if err := gql.Query(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryUser) != 1 {
		return User{}, ErrNotFound
	}

	return result.QueryUser[0], nil
}

// =============================================================================

func add(ctx context.Context, gql *graphql.GraphQL, user User) (User, error) {
	mutation, result := prepareAdd(user)
	if err := gql.Query(ctx, mutation, &result); err != nil {
		return User{}, errors.Wrap(err, "failed to add user")
	}

	if len(result.AddUser.User) != 1 {
		return User{}, errors.New("user id not returned")
	}

	user.ID = result.AddUser.User[0].ID
	return user, nil
}

func update(ctx context.Context, gql *graphql.GraphQL, user User) error {
	if user.ID == "" {
		return errors.New("user missing id")
	}

	mutation, result := prepareUpdate(user)
	if err := gql.Query(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if result.UpdateUser.NumUids != 1 {
		msg := fmt.Sprintf("failed to update user: NumUids: %d  Msg: %s", result.UpdateUser.NumUids, result.UpdateUser.Msg)
		return errors.New(msg)
	}

	return nil
}

func delete(ctx context.Context, gql *graphql.GraphQL, userID string) error {
	if userID == "" {
		return errors.New("missing user id")
	}

	mutation, result := prepareDelete(userID)
	if err := gql.Query(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if result.DeleteUser.NumUids != 0 {
		msg := fmt.Sprintf("failed to delete user: NumUids: %d  Msg: %s", result.DeleteUser.NumUids, result.DeleteUser.Msg)
		return errors.New(msg)
	}

	return nil
}

// =============================================================================

func prepareAdd(user User) (string, addResult) {
	var result addResult
	mutation := fmt.Sprintf(`
mutation {
	addUser(input: [{
		name: %q
		email: %q
		role: %s
		password_hash: %q
		date_created: %q
		date_updated: %q
	}])
	%s
}`, user.Name, user.Email, user.Role, user.PasswordHash,
		user.DateCreated.UTC().Format(time.RFC3339),
		user.DateUpdated.UTC().Format(time.RFC3339),
		result.document())

	return mutation, result
}

func prepareUpdate(user User) (string, updateResult) {
	var result updateResult
	mutation := fmt.Sprintf(`
mutation {
	updateUser(input: {
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
}`, user.ID, user.Name, user.Email, user.Role, user.PasswordHash,
		user.DateCreated.UTC().Format(time.RFC3339),
		user.DateUpdated.UTC().Format(time.RFC3339),
		result.document())

	return mutation, result
}

func prepareDelete(userID string) (string, deleteResult) {
	var result deleteResult
	mutation := fmt.Sprintf(`
mutation {
	deleteUser(filter: { id: [%q] })
	%s
}`, userID, result.document())

	return mutation, result
}
