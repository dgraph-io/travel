package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/conf"
	"github.com/dgraph-io/travel/cmd/travel-data/internal/feed"
	"github.com/pkg/errors"
)

// TODO
/*
	We are only storing 1 result of places at this time.
	Finish tests for data, places.
	Write integration tests.
	Finish Advisory and Flight feeds.
	Decide on UI, Ratel or some CLI tooling.
	Validate upserts are working for weather and places.
	Need to apply proper times on the Client.Do calls in the feeds.
	Running in Kind with Ready checks
	Working with Circle CI
	Review the use of foriegn key kinds of relationship.

	Place Store
		Establish the relationship by creating an edge with the city node.

	Advisory Store
		Establish the relationship by creating an edge with the city node.

	Weather Store
		Just connect the weather node with city node via an edge.
		Instead, check whether the City node has it's weather information available
			via the `weather` edge. Also update the weather info if its last udpated time
			is more than 24 hours.
		We need to flip the feed fetching. First find whether the weather info
			exists and its not outdated and then go fetch from the Feed only if required.

	Advisory Store

*/

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	if err := run(); err != nil {
		log.Println("main : ERROR : ", err)
		os.Exit(1)
	}
}

func run() error {

	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "DATA : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// =========================================================================
	// Configuration

	var cfg struct {
		Search struct {
			Name    string  `conf:"default:sydney"`
			Lat     float64 `conf:"default:-33.865143"`
			Lng     float64 `conf:"default:151.209900"`
			Keyword string  `conf:"default:hotels"`
			Radius  int     `conf:"default:5000"`
		}
		APIKeys struct {
			MapsKey    string `conf:"default:AIzaSyBR0-ToiYlrhPlhidE7DA-Zx7EfE7FnUek"`
			WeatherKey string `conf:"default:b2302a48062dc1da72430c612557498d"`
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

	search := feed.Search{
		Name:    cfg.Search.Name,
		Lat:     cfg.Search.Lat,
		Lng:     cfg.Search.Lng,
		Keyword: cfg.Search.Keyword,
		Radius:  uint(cfg.Search.Radius),
	}

	keys := feed.Keys{
		MapKey:     cfg.APIKeys.MapsKey,
		WeatherKey: cfg.APIKeys.WeatherKey,
	}

	if err := feed.Work(log, search, keys, cfg.DB.Host); err != nil {
		return err
	}

	return nil
}
