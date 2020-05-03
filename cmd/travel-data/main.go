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
	UI
		https://observablehq.com/@d3/force-directed-graph

	Building/Testing
		Working with Circle CI

	travel-api
		Identify an end point we need that dgraph can't provide.

	Kubernetes
		Running in Kind with Ready checks
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
		URL struct {
			Advisory string `conf:"default:https://www.travel-advisory.info/api"`
			Weather  string `conf:"default:http://api.openweathermap.org/data/2.5/weather"`
		}
		Dgraph struct {
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

	url := feed.URL{
		Advisory: cfg.URL.Advisory,
		Weather:  cfg.URL.Weather,
	}

	if err := feed.Work(log, dgraph, search, keys, url); err != nil {
		return err
	}

	return nil
}
