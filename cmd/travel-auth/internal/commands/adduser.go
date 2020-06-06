package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/pkg/errors"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")

// AddUser handles the creation of users.
func AddUser(dgraph data.Dgraph, newUser data.NewUser) error {
	if newUser.Name == "" || newUser.Email == "" || newUser.Password == "" || newUser.Roles == nil {
		fmt.Println("help: adduser <name> <email> <password> <role>")
		return ErrHelp
	}

	db, err := data.NewDB(dgraph)
	if err != nil {
		return errors.Wrap(err, "init database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u, err := db.Mutate.AddUser(ctx, newUser, time.Now())
	if err != nil {
		return errors.Wrap(err, "adding user")
	}

	fmt.Println("user id:", u.ID)
	return nil
}
