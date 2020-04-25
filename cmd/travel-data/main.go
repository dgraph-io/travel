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
	GraphQL
		Remove the need to say mutation {} and query {}
		Remove extra logging in the graphql package
		Not sure having the city type in places is good.
		Not sure if data shouldn't have its own type system.

	Bugs
		AvgUserRating is being saved as a float32. I have validated the JSON has
			a floating point number (4.5) but on the query I am getting back (4).
			Look at store_test.go storePlace test.

	General
		We are only storing 1 result of places at this time.
		Decide on UI, Ratel or some CLI tooling.
		May not want to use the default Client in the Advisory/Weather feeds.
		Review all comments for necessity.

	Building/Testing
		Write integration tests.
		Running in Kind with Ready checks
		Working with Circle CI
		traval-data binary gets git information as well.

	Data
		Add schema predicates for weather and advisory.
		Do more to validate the Schema by checking actual predicate values.
		Review the use of foriegn key kinds of relationship.
		Write more tests.

	Place Feed
		Establish the relationship by creating an edge with the city node.
		Validate upserts are working.
		Write more tests.

	Advisory Feed
		Establish the relationship by creating an edge with the city node.
		Validate upserts are working.
		Write more tests.

	Weather Feed
		Validate upserts are working.
		Just connect the weather node with city node via an edge.
		Instead, check whether the City node has it's weather information available
			via the `weather` edge. Also update the weather info if its last udpated time
			is more than 24 hours.
		We need to flip the feed fetching. First find whether the weather info
			exists and its not outdated and then go fetch from the Feed only if required.
		Write more tests.

	Flight Feed
		Incorporate this feed into the data.
		Write more tests.
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
		City struct {
			CountryCode string  `conf:"default:AU"`
			Name        string  `conf:"default:sydney"`
			Lat         float64 `conf:"default:-33.865143"`
			Lng         float64 `conf:"default:151.209900"`
		}
		Search struct {
			Keyword string `conf:"default:hotels"`
			Radius  int    `conf:"default:5000"`
		}
		APIKeys struct {
			MapsKey    string `conf:"default:AIzaSyBR0-ToiYlrhPlhidE7DA-Zx7EfE7FnUek"`
			WeatherKey string `conf:"default:b2302a48062dc1da72430c612557498d"`
		}
		Dgraph struct {
			DBHost  string `conf:"default:localhost:9080"`
			APIHost string `conf:"default:localhost:8080"`
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

	dgraph := feed.Dgraph{
		DBHost:  cfg.Dgraph.DBHost,
		APIHost: cfg.Dgraph.APIHost,
	}

	search := feed.Search{
		CountryCode: cfg.City.CountryCode,
		CityName:    cfg.City.Name,
		Lat:         cfg.City.Lat,
		Lng:         cfg.City.Lng,
		Keyword:     cfg.Search.Keyword,
		Radius:      uint(cfg.Search.Radius),
	}

	keys := feed.Keys{
		MapKey:     cfg.APIKeys.MapsKey,
		WeatherKey: cfg.APIKeys.WeatherKey,
	}

	if err := feed.Work(log, dgraph, search, keys); err != nil {
		return err
	}

	return nil
}
