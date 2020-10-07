package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/user"
	"github.com/pkg/errors"
)

// GetUser returns information about a user by email.
func GetUser(gqlConfig data.GraphQLConfig, email string) error {
	if email == "" {
		fmt.Println("help: getuser <email>")
		return ErrHelp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gql := data.NewGraphQL(gqlConfig)
	u := user.New(gql)

	usr, err := u.QueryByEmail(ctx, email)
	if err != nil {
		return errors.Wrap(err, "getting user")
	}

	fmt.Printf("user: %#v\n", usr)
	return nil
}
