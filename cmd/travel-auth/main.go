package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/conf"
	"github.com/dgraph-io/travel/cmd/travel-auth/internal/commands"
	"github.com/dgraph-io/travel/internal/data"
	"github.com/pkg/errors"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	if err := run(); err != nil {
		log.Printf("error: %s", err)
		os.Exit(1)
	}
}

func run() error {

	// =========================================================================
	// Configuration

	var cfg struct {
		conf.Version
		Args   conf.Args
		Dgraph struct {
			URL            string `conf:"default:http://0.0.0.0:8080"`
			AuthHeaderName string
			AuthToken      string
		}
	}

	cfg.Version.SVN = build
	cfg.Version.Desc = "copyright information here"

	const prefix = "TRAVEL"
	if err := conf.Parse(os.Args[1:], prefix, &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage(prefix, &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString(prefix, &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// =========================================================================
	// Commands

	dgraph := data.Dgraph{
		URL:            cfg.Dgraph.URL,
		AuthHeaderName: cfg.Dgraph.AuthHeaderName,
		AuthToken:      cfg.Dgraph.AuthToken,
	}

	switch cfg.Args.Num(0) {
	case "adduser":
		newUser := data.NewUser{
			Name:     cfg.Args.Num(1),
			Email:    cfg.Args.Num(2),
			Password: cfg.Args.Num(3),
			Roles:    []string{cfg.Args.Num(4)},
		}

		if err := commands.AddUser(dgraph, newUser); err != nil {
			return errors.Wrap(err, "adding user")
		}

	case "getuser":
		email := cfg.Args.Num(1)
		if err := commands.GetUser(dgraph, email); err != nil {
			return errors.Wrap(err, "getting user")
		}

	case "genkeys":
		if err := commands.GenerateKeys(); err != nil {
			return errors.Wrap(err, "generating keys")
		}

	case "gentoken":
		email := cfg.Args.Num(1)
		privateKeyFile := cfg.Args.Num(2)
		if err := commands.GenerateToken(dgraph, email, privateKeyFile); err != nil {
			return errors.Wrap(err, "generating token")
		}

	default:
		return errors.New("must specify a value command: [useradd,getuser,genkeys,gentoken]")
	}

	return nil
}
