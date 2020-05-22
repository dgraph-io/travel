package feed

import (
	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/feeds/advisory"
	"github.com/dgraph-io/travel/internal/feeds/places"
	"github.com/dgraph-io/travel/internal/feeds/weather"
)

type m struct{}

// Marshal provides marshaling functions for the different
// feed values.
var marshal m

// Place marshals a Place value from the places package into
// a data Place value.
func (m) Place(place places.Place, cityID string, category string) data.Place {
	return data.Place{
		PlaceID:          place.PlaceID,
		Category:         category,
		CityID:           data.CityID{ID: cityID},
		CityName:         place.CityName,
		Name:             place.Name,
		Address:          place.Address,
		Lat:              place.Lat,
		Lng:              place.Lng,
		LocationType:     place.LocationType,
		AvgUserRating:    place.AvgUserRating,
		NumberOfRatings:  place.NumberOfRatings,
		GmapsURL:         place.GmapsURL,
		PhotoReferenceID: place.PhotoReferenceID,
	}
}

// Places marshals a collection of Place values from the
// places package into a collection of data Place values.
func (m) Places(places []places.Place, cityID string, category string) []data.Place {
	dataPlaces := make([]data.Place, len(places))
	for i, place := range places {
		dataPlaces[i] = marshal.Place(place, cityID, category)
	}
	return dataPlaces
}

// Advisory marshals a Advisory value from the advisory package into
// a data Advisory value.
func (m) Advisory(advisory advisory.Advisory) data.Advisory {
	return data.Advisory{
		Country:     advisory.Country,
		CountryCode: advisory.CountryCode,
		Continent:   advisory.Continent,
		Score:       advisory.Score,
		LastUpdated: advisory.LastUpdated,
		Message:     advisory.Message,
		Source:      advisory.Source,
	}
}

// Weather marshals a Weather value from the weather package into
// a data Weather value.
func (m) Weather(weather weather.Weather) data.Weather {
	return data.Weather{
		CityName:      weather.CityName,
		Visibility:    weather.Visibility,
		Desc:          weather.Desc,
		Temp:          weather.Temp,
		FeelsLike:     weather.FeelsLike,
		MinTemp:       weather.MinTemp,
		MaxTemp:       weather.MaxTemp,
		Pressure:      weather.Pressure,
		Humidity:      weather.Humidity,
		WindSpeed:     weather.WindSpeed,
		WindDirection: weather.WindDirection,
		Sunrise:       weather.Sunrise,
		Sunset:        weather.Sunset,
	}
}
