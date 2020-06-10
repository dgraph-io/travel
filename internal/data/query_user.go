package data

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// User returns the specified user from the database by the city id.
func (q *query) User(ctx context.Context, userID string) (User, error) {
	query := fmt.Sprintf(`
query {
	getUser(id: %q) {
		id
		name
		email
		roles
		password_hash
		date_created
		date_updated
	}
}`, userID)

	var result struct {
		GetUser User `json:"getUser"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if result.GetUser.ID == "" {
		return User{}, ErrUserNotFound
	}

	return result.GetUser, nil
}

// UserByEmail returns the specified user from the database by email.
func (q *query) UserByEmail(ctx context.Context, email string) (User, error) {
	query := fmt.Sprintf(`
query {
	queryUser(filter: { email: { eq: %q } }) {
		id
		name
		email
		roles
		password_hash
		date_created
		date_updated
	}
}`, email)

	var result struct {
		QueryUser []User `json:"queryUser"`
	}
	if err := q.graphql.Query(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryUser) != 1 {
		return User{}, ErrPlaceNotFound
	}

	return result.QueryUser[0], nil
}
