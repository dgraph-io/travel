package loader

import (
	"github.com/dgraph-io/travel/business/data/advisory"
	"github.com/dgraph-io/travel/business/data/place"
	"github.com/dgraph-io/travel/business/data/weather"
	advisoryfeed "github.com/dgraph-io/travel/business/feeds/advisory"
	placesfeed "github.com/dgraph-io/travel/business/feeds/places"
	weatherfeed "github.com/dgraph-io/travel/business/feeds/weather"
)

// marshalPlace marshals a Place value from the places package into
// a data Place value.
func marshalPlace(feedData placesfeed.Place, cityID string, category string) place.Info {
	return place.Info{
		PlaceID:          feedData.PlaceID,
		Category:         category,
		City:             place.City{ID: cityID},
		CityName:         feedData.CityName,
		Name:             feedData.Name,
		Address:          feedData.Address,
		Lat:              feedData.Lat,
		Lng:              feedData.Lng,
		LocationType:     feedData.LocationType,
		AvgUserRating:    feedData.AvgUserRating,
		NumberOfRatings:  feedData.NumberOfRatings,
		GmapsURL:         feedData.GmapsURL,
		PhotoReferenceID: feedData.PhotoReferenceID,
	}
}

// marshalAdvisory marshals a Advisory value from the advisory package into
// a data Advisory value.
func marshalAdvisory(feedData advisoryfeed.Advisory, cityID string) advisory.Info {
	return advisory.Info{
		City:        advisory.City{ID: cityID},
		Country:     feedData.Country,
		CountryCode: feedData.CountryCode,
		Continent:   feedData.Continent,
		Score:       feedData.Score,
		LastUpdated: feedData.LastUpdated,
		Message:     feedData.Message,
		Source:      feedData.Source,
	}
}

// marshalWeather marshals a Weather value from the weather package into
// a data Weather value.
func marshalWeather(feedData weatherfeed.Weather, cityID string) weather.Info {
	return weather.Info{
		City:          weather.City{ID: cityID},
		CityName:      feedData.CityName,
		Visibility:    feedData.Visibility,
		Desc:          feedData.Desc,
		Temp:          feedData.Temp,
		FeelsLike:     feedData.FeelsLike,
		MinTemp:       feedData.MinTemp,
		MaxTemp:       feedData.MaxTemp,
		Pressure:      feedData.Pressure,
		Humidity:      feedData.Humidity,
		WindSpeed:     feedData.WindSpeed,
		WindDirection: feedData.WindDirection,
		Sunrise:       feedData.Sunrise,
		Sunset:        feedData.Sunset,
	}
}
