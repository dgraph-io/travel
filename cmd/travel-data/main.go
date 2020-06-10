package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/conf"
	"github.com/dgraph-io/travel/cmd/travel-data/internal/feed"
	"github.com/dgraph-io/travel/internal/data"
	"github.com/pkg/errors"
)

type city struct {
	CountryCode string
	Name        string
	Lat         float64
	Lng         float64
}

// These are the currently cities supported.
var cities = []city{
	{"US", "miami", 25.7617, -80.1918},
	{"US", "new york", 40.730610, -73.935242},
	{"AU", "sydney", -33.865143, 151.209900},
}

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	log := log.New(os.Stdout, "DATA : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if err := run(log); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	// =========================================================================
	// Configuration

	var cfg struct {
		conf.Version
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
		CustomFunctions struct {
			SendEmailURL string `conf:"default:http://travel-api:3000/v1/email"`
		}
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
	// App Starting

	log.Printf("main: Application initializing: version %q", build)
	defer log.Println("main: Completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: Config:\n%v\n", out)

	// =========================================================================
	// Process the feed

	dbConfig := data.DBConfig{
		URL:            cfg.Dgraph.URL,
		AuthHeaderName: cfg.Dgraph.AuthHeaderName,
		AuthToken:      cfg.Dgraph.AuthToken,
	}

	schemaConfig := data.SchemaConfig{
		SendEmailURL: cfg.CustomFunctions.SendEmailURL,
	}

	keys := feed.Keys{
		MapKey:     cfg.APIKeys.MapsKey,
		WeatherKey: cfg.APIKeys.WeatherKey,
	}

	url := feed.URL{
		Advisory: cfg.URL.Advisory,
		Weather:  cfg.URL.Weather,
	}

	for _, city := range cities {
		search := feed.Search{
			CityName:    city.Name,
			CountryCode: city.CountryCode,
			Lat:         city.Lat,
			Lng:         city.Lng,
			Categories:  cfg.Search.Categories,
			Radius:      uint(cfg.Search.Radius),
		}
		if err := feed.Work(log, dbConfig, schemaConfig, search, keys, url); err != nil {
			return err
		}
	}

	return nil
}
