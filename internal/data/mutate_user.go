package data

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

type mutateUser struct {
	marshal userMarshal
}

var mutUser mutateUser

func (mutateUser) add(ctx context.Context, graphql *graphql.GraphQL, user User) (User, error) {
	if user.ID != "" {
		return User{}, errors.New("user contains id")
	}

	mutation, result := mutUser.marshal.add(user)
	if err := graphql.Query(ctx, mutation, &result); err != nil {
		return User{}, errors.Wrap(err, "failed to add user")
	}

	if len(result.AddUser.User) != 1 {
		return User{}, errors.New("user id not returned")
	}

	user.ID = result.AddUser.User[0].ID
	return user, nil
}

func (mutateUser) update(ctx context.Context, graphql *graphql.GraphQL, user User) error {
	if user.ID == "" {
		return errors.New("user missing id")
	}

	mutation, result := mutUser.marshal.update(user)
	if err := graphql.Query(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if result.UpdateUser.NumUids != 1 {
		msg := fmt.Sprintf("failed to update user: NumUids: %d  Msg: %s", result.UpdateUser.NumUids, result.UpdateUser.Msg)
		return errors.New(msg)
	}

	return nil
}

func (mutateUser) delete(ctx context.Context, graphql *graphql.GraphQL, userID string) error {
	if userID == "" {
		return errors.New("missing user id")
	}

	mutation, result := mutUser.marshal.delete(userID)
	if err := graphql.Query(ctx, mutation, nil); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if result.DeleteUser.NumUids != 0 {
		msg := fmt.Sprintf("failed to delete user: NumUids: %d  Msg: %s", result.DeleteUser.NumUids, result.DeleteUser.Msg)
		return errors.New(msg)
	}

	return nil
}

type userMarshal struct{}

func (userMarshal) add(user User) (string, userIDResult) {
	var result userIDResult
	mutation := fmt.Sprintf(`
mutation {
	addUser(input: [{
		name: %q
		email: %q
		roles: [%s]
		password_hash: %q
		date_created: %q
		date_updated: %q
	}])
	%s
}`, user.Name, user.Email, strings.Join(user.Roles, ","), user.PasswordHash,
		user.DateCreated.UTC().Format(time.RFC3339),
		user.DateUpdated.UTC().Format(time.RFC3339),
		result.marshal())

	return mutation, result
}

func (userMarshal) update(user User) (string, userUpdateResult) {
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
  			roles: [%s]
  			password_hash: %q
  			date_created: %q
  			date_updated: %q
		}
	})
	%s
}`, user.ID, user.Name, user.Email, strings.Join(user.Roles, ","), user.PasswordHash,
		user.DateCreated.UTC().Format(time.RFC3339),
		user.DateUpdated.UTC().Format(time.RFC3339),
		result.marshal())

	return mutation, result
}

func (userMarshal) delete(userID string) (string, userDeleteResult) {
	var result userDeleteResult
	mutation := fmt.Sprintf(`
mutation {
	deleteUser(filter: { id: [%q] })
	%s
}`, userID, result.marshal())

	return mutation, result
}

type userIDResult struct {
	AddUser struct {
		User []struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"addUser"`
}

func (userIDResult) marshal() string {
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

func (userUpdateResult) marshal() string {
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

func (userDeleteResult) marshal() string {
	return `{
		msg,
		numUids,
	}`
}
