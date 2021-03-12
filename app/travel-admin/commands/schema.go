package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/schema"
	"github.com/dgraph-io/travel/business/data/user"
	"github.com/dgraph-io/travel/business/loader"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Schema handles the updating of the schema.
func Schema(gqlConfig data.GraphQLConfig, config schema.Config) error {
	if err := loader.UpdateSchema(gqlConfig, config); err != nil {
		return err
	}

	fmt.Println("schema updated")
	return nil
}

// Seed handles loading the databse with a user and city data.
func Seed(log *log.Logger, gqlConfig data.GraphQLConfig, config loader.Config) error {
	if os.Getenv("TRAVEL_API_KEYS_MAPS_KEY") == "" {
		return errors.New("TRAVEL_API_KEYS_MAPS_KEY is not set with map key")
	}

	newUser := user.NewUser{
		Name:     "Bill Kennedy",
		Email:    "bill@ardanlabs.com",
		Password: "gopher",
		Role:     "ADMIN",
	}

	log.Println("main: Adding User:", newUser.Name)
	if err := AddUser(log, gqlConfig, newUser); err != nil {
		if errors.Cause(err) != user.ErrExists {
			return errors.Wrap(err, "adding user")
		}
	}

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

		log.Println("main: Adding city:", search.CityName)
		traceID := uuid.New().String()
		if err := loader.UpdateData(log, gqlConfig, traceID, config, search); err != nil {
			return err
		}
	}

	fmt.Println("main: Data seeded")
	return nil
}
