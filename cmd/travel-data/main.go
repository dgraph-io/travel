package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/conf"
	"github.com/dgraph-io/travel/cmd/travel-data/internal/feed"
	"github.com/pkg/errors"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	// Quick way to run Dgraph
	// docker run -it -p 8080:8080 -p 9080:9080 -p 8080:8080dgraph/standalone:v20.03.0
	// Go to localhost:8000 on your Browser for Ratel UI
	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "PLACES : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// =========================================================================
	// Configuration

	var cfg struct {
		API struct {
			Key string `conf:"default:AIzaSyBR0-ToiYlrhPlhidE7DA-Zx7EfE7FnUek"`
		}
		DB struct {
			Host string `conf:"default:localhost:9080"`
		}
	}

	if err := conf.Parse(os.Args[1:], "PLACES", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("PLACES", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// =========================================================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars.
	log.Printf("main : Started : Application initializing : version %q", build)
	defer log.Println("main : Completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main : Config :\n%v\n", out)

	// =========================================================================
	// Process the feed

	if err := feed.Pull(log, cfg.API.Key, cfg.DB.Host); err != nil {
		return err
	}

	return nil
}
