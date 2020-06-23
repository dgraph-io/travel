package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/user"
	"github.com/pkg/errors"
)

// AddUser handles the creation of users.
func AddUser(gqlConfig data.GraphQLConfig, newUser user.NewUser) error {
	if newUser.Name == "" || newUser.Email == "" || newUser.Password == "" || newUser.Role == "" {
		fmt.Println("help: adduser <name> <email> <password> <role>")
		return ErrHelp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gql := data.NewGraphQL(gqlConfig)

	u, err := user.Add(ctx, gql, newUser, time.Now())
	if err != nil {
		return errors.Wrap(err, "adding user")
	}

	fmt.Println("user id:", u.ID)
	return nil
}
