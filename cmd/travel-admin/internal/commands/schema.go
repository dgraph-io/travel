package commands

import (
	"fmt"
	"log"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/loader"
)

// Schema handles the updating of the schema.
func Schema(dbConfig data.DBConfig, schemaConfig data.SchemaConfig) error {
	if err := loader.UpdateSchema(dbConfig, schemaConfig); err != nil {
		return err
	}

	fmt.Println("schema updated")
	return nil
}

// Seed handles loading the databse with city data.
func Seed(log *log.Logger, dbConfig data.DBConfig, config loader.Config) error {
	var cities = []struct {
		CountryCode string
		Name        string
		Lat         float64
		Lng         float64
	}{
		{"US", "miami", 25.7617, -80.1918},
		{"US", "new york", 40.730610, -73.935242},
		{"AU", "sydney", -33.865143, 151.209900},
	}

	for _, city := range cities {
		search := loader.Search{
			CityName:    city.Name,
			CountryCode: city.CountryCode,
			Lat:         city.Lat,
			Lng:         city.Lng,
		}

		log.Println("main: Updating data for city:", search.CityName)
		if err := loader.UpdateData(log, dbConfig, config, search); err != nil {
			return err
		}
	}

	fmt.Println("data seeded")
	return nil
}
