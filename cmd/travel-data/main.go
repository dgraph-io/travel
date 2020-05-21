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

// TODO
/*
	UI
		https://observablehq.com/@d3/force-directed-graph

	travel-api (coming soon)
		Identify an end point we need that dgraph can't provide.

	Kubernetes
		Running in Kind with Ready checks
*/

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
		City struct {
			CountryCode string  `conf:"default:AU"`
			Name        string  `conf:"default:sydney"`
			Lat         float64 `conf:"default:-33.865143"`
			Lng         float64 `conf:"default:151.209900"`
		}
		Search struct {
			Keywords []string `conf:"default:mcdonalds;dominos;kfc"`
			Radius   int      `conf:"default:5000"`
		}
		APIKeys struct {
			MapsKey    string `conf:"default:AIzaSyAKz3OhgUF-BO3dsFQWEwHsGmAh7BXe15w"`
			WeatherKey string `conf:"default:5b68961dd2602c2f722f02448d2de823"`
		}
		URL struct {
			Advisory string `conf:"default:https://www.travel-advisory.info/api"`
			Weather  string `conf:"default:http://api.openweathermap.org/data/2.5/weather"`
		}
		Dgraph struct {
			Protocol       string `conf:"default:http"`
			APIHost        string `conf:"default:0.0.0.0:8080"`
			BasicAuthToken string
		}
	}
	cfg.Version.SVN = build
	cfg.Version.Desc = "copyright information here"

	if err := conf.Parse(os.Args[1:], "DATA", &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage("DATA", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString("DATA", &cfg)
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

	log.Printf("main: Application initializing : version %q", build)
	defer log.Println("main: Completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main: Config:\n%v\n", out)

	// =========================================================================
	// Process the feed

	dgraph := data.Dgraph{
		Protocol:       cfg.Dgraph.Protocol,
		APIHostInside:  cfg.Dgraph.APIHost,
		BasicAuthToken: cfg.Dgraph.BasicAuthToken,
	}

	search := feed.Search{
		CountryCode: cfg.City.CountryCode,
		CityName:    cfg.City.Name,
		Lat:         cfg.City.Lat,
		Lng:         cfg.City.Lng,
		Keywords:    cfg.Search.Keywords,
		Radius:      uint(cfg.Search.Radius),
	}

	keys := feed.Keys{
		MapKey:     cfg.APIKeys.MapsKey,
		WeatherKey: cfg.APIKeys.WeatherKey,
	}

	url := feed.URL{
		Advisory: cfg.URL.Advisory,
		Weather:  cfg.URL.Weather,
	}

	if err := feed.Work(log, dgraph, search, keys, url); err != nil {
		return err
	}

	return nil
}
