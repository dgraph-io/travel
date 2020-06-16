package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/travel/business/data"
	"github.com/pkg/errors"
)

// GetUser returns information about a user by email.
func GetUser(dbConfig data.DBConfig, email string) error {
	if email == "" {
		fmt.Println("help: getuser <email>")
		return ErrHelp
	}

	db, err := data.NewDB(dbConfig)
	if err != nil {
		return errors.Wrap(err, "init database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u, err := db.Query.UserByEmail(ctx, email)
	if err != nil {
		return errors.Wrap(err, "getting user")
	}

	fmt.Printf("user: %#v\n", u)
	return nil
}
