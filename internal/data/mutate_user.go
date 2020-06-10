package data

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// AddUser adds a new user to the database. If the user already exists
// this function will fail but the found user is returned. If the user is
// being added, the user with the id from the database is returned.
func (m mutate) AddUser(ctx context.Context, newUser NewUser, now time.Time) (User, error) {
	if user, err := m.query.UserByEmail(ctx, newUser.Email); err == nil {
		return user, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "generating password hash")
	}

	u := User{
		Name:         newUser.Name,
		Email:        newUser.Email,
		Role:         newUser.Role,
		PasswordHash: string(hash),
		DateCreated:  now,
		DateUpdated:  now,
	}

	u, err = user.add(ctx, m.graphql, u)
	if err != nil {
		return User{}, errors.Wrap(err, "adding user to database")
	}

	return u, nil
}

// UpdateUser updates a user in the database by its ID. If the user doesn't
// already exist, this function will fail.
func (m mutate) UpdateUser(ctx context.Context, usr User) error {
	if _, err := m.query.User(ctx, usr.ID); err != nil {
		return ErrUserNotExists
	}

	if err := user.update(ctx, m.graphql, usr); err != nil {
		return errors.Wrap(err, "updating user in database")
	}

	return nil
}

// DeleteUser removes a user from the database by its ID. If the user doesn't
// already exist, this function will fail.
func (m mutate) DeleteUser(ctx context.Context, userID string) error {
	if _, err := m.query.User(ctx, userID); err != nil {
		return ErrUserNotExists
	}

	if err := user.delete(ctx, m.graphql, userID); err != nil {
		return errors.Wrap(err, "deleting user in database")
	}

	return nil
}

// =============================================================================

type usr struct {
	prepare userPrepare
}

var user usr

func (u usr) add(ctx context.Context, graphql *graphql.GraphQL, user User) (User, error) {
	if user.ID != "" {
		return User{}, errors.New("user contains id")
	}

	mutation, result := u.prepare.add(user)
	if err := graphql.Query(ctx, mutation, &result); err != nil {
		return User{}, errors.Wrap(err, "failed to add user")
	}

	if len(result.AddUser.User) != 1 {
		return User{}, errors.New("user id not returned")
	}

	user.ID = result.AddUser.User[0].ID
	return user, nil
}

func (u usr) update(ctx context.Context, graphql *graphql.GraphQL, user User) error {
	if user.ID == "" {
		return errors.New("user missing id")
	}

	mutation, result := u.prepare.update(user)
	if err := graphql.Query(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if result.UpdateUser.NumUids != 1 {
		msg := fmt.Sprintf("failed to update user: NumUids: %d  Msg: %s", result.UpdateUser.NumUids, result.UpdateUser.Msg)
		return errors.New(msg)
	}

	return nil
}

func (u usr) delete(ctx context.Context, graphql *graphql.GraphQL, userID string) error {
	if userID == "" {
		return errors.New("missing user id")
	}

	mutation, result := u.prepare.delete(userID)
	if err := graphql.Query(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if result.DeleteUser.NumUids != 0 {
		msg := fmt.Sprintf("failed to delete user: NumUids: %d  Msg: %s", result.DeleteUser.NumUids, result.DeleteUser.Msg)
		return errors.New(msg)
	}

	return nil
}

// =============================================================================

type userPrepare struct{}

func (userPrepare) add(user User) (string, userAddResult) {
	var result userAddResult
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

func (userPrepare) update(user User) (string, userUpdateResult) {
	var result userUpdateResult
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

func (userPrepare) delete(userID string) (string, userDeleteResult) {
	var result userDeleteResult
	mutation := fmt.Sprintf(`
mutation {
	deleteUser(filter: { id: [%q] })
	%s
}`, userID, result.document())

	return mutation, result
}

type userAddResult struct {
	AddUser struct {
		User []struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"addUser"`
}

func (userAddResult) document() string {
	return `{
		user {
			id
		}
	}`
}

type userUpdateResult struct {
	UpdateUser struct {
		Msg     string
		NumUids int
	} `json:"updateUser"`
}

func (userUpdateResult) document() string {
	return `{
		msg,
		numUids,
	}`
}

type userDeleteResult struct {
	DeleteUser struct {
		Msg     string
		NumUids int
	} `json:"deleteUser"`
}

func (userDeleteResult) document() string {
	return `{
		msg,
		numUids,
	}`
}
