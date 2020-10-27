package main

import (
	"expvar"
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/conf"
	"github.com/dgraph-io/travel/app/travel-admin/commands"
	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/schema"
	"github.com/dgraph-io/travel/business/data/user"
	"github.com/dgraph-io/travel/business/loader"
	"github.com/pkg/errors"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	log := log.New(os.Stdout, "ADMIN : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if err := run(log); err != nil {
		if errors.Cause(err) != commands.ErrHelp {
			log.Printf("error: %s", err)
		}
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	// =========================================================================
	// Configuration

	var cfg struct {
		conf.Version
		Args   conf.Args
		Dgraph struct {
			URL            string `conf:"default:http://0.0.0.0:8080"`
			AuthHeaderName string `conf:"default:X-Travel-Auth"`
			AuthToken      string
		}
		CustomFunctions struct {
			UploadFeedURL string `conf:"default:http://0.0.0.0:3000/v1/feed/upload"`
		}
		Search struct {
			Categories []string `conf:"default:restaurant;bar;supermarket"`
			Radius     int      `conf:"default:5000"`
		}
		APIKeys struct {
			// You need to generate a Google Key to support Places API and JS Maps.
			// Once you have the key it's best to export it.
			// export UI_API_KEYS_MAPS_KEY=<KEY HERE>
			MapsKey    string
			WeatherKey string `conf:"default:5b68961dd2602c2f722f02448d2de823"`
		}
		URL struct {
			Advisory string `conf:"default:https://www.travel-advisory.info/api"`
			Weather  string `conf:"default:http://api.openweathermap.org/data/2.5/weather"`
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

	// For convenience with the training material, an ADMIN token is provided.
	if cfg.Dgraph.AuthToken == "" {
		cfg.Dgraph.AuthToken = schema.AdminJWT
	}

	// =========================================================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	log.Printf("main : Started : Application initializing : version %q", build)
	defer log.Println("main: Completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: Config:\n%v\n", out)

	// =========================================================================
	// Commands

	gqlConfig := data.GraphQLConfig{
		URL:            cfg.Dgraph.URL,
		AuthHeaderName: cfg.Dgraph.AuthHeaderName,
		AuthToken:      cfg.Dgraph.AuthToken,
	}

	switch cfg.Args.Num(0) {
	case "schema":
		config := schema.Config{
			CustomFunctions: schema.CustomFunctions{
				UploadFeedURL: cfg.CustomFunctions.UploadFeedURL,
			},
		}

		if err := commands.Schema(gqlConfig, config); err != nil {
			return errors.Wrap(err, "updating schema")
		}

	case "seed":
		config := loader.Config{
			Filter: loader.Filter{
				Categories: cfg.Search.Categories,
				Radius:     uint(cfg.Search.Radius),
			},
			Keys: loader.Keys{
				MapKey:     cfg.APIKeys.MapsKey,
				WeatherKey: cfg.APIKeys.WeatherKey,
			},
			URL: loader.URL{
				Advisory: cfg.URL.Advisory,
				Weather:  cfg.URL.Weather,
			},
		}

		if err := commands.Seed(log, gqlConfig, config); err != nil {
			return errors.Wrap(err, "seeding database")
		}

	case "adduser":
		newUser := user.NewUser{
			Name:     cfg.Args.Num(1),
			Email:    cfg.Args.Num(2),
			Password: cfg.Args.Num(3),
			Role:     cfg.Args.Num(4),
		}

		if err := commands.AddUser(log, gqlConfig, newUser); err != nil {
			return errors.Wrap(err, "adding user")
		}

	case "getuser":
		email := cfg.Args.Num(1)
		if err := commands.GetUser(log, gqlConfig, email); err != nil {
			return errors.Wrap(err, "getting user")
		}

	case "keygen":
		if err := commands.KeyGen(); err != nil {
			return errors.Wrap(err, "generating keys")
		}

	case "gentoken":
		email := cfg.Args.Num(1)
		privateKeyFile := cfg.Args.Num(2)
		algorithm := cfg.Args.Num(3)
		if err := commands.GenToken(log, gqlConfig, email, privateKeyFile, algorithm); err != nil {
			return errors.Wrap(err, "generating token")
		}

	default:
		fmt.Println("adduser: add a new user to the system")
		fmt.Println("getuser: retrieve information about a user")
		fmt.Println("keygen: generate a set of private/public key files")
		fmt.Println("gentoken: generate a JWT for a user with claims")
		fmt.Println("provide a command to get more help.")
		return commands.ErrHelp
	}

	return nil
}
