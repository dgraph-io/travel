package data

import "time"

// User represents someone with access to the system.
type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Roles        []string  `json:"roles"`
	PasswordHash string    `json:"password_hash"`
	DateCreated  time.Time `json:"date_created"`
	DateUpdated  time.Time `json:"date_updated"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	Roles           []string `json:"roles"`
	Password        string   `json:"password"`
	PasswordConfirm string   `json:"password_confirm"`
}

// EmailRequest is the data that will be sent with a sendemail request.
type EmailRequest struct {
	UserID   string `json:"userid"`
	NodeType string `json:"nodetype"`
	NodeID   string `json:"nodeid"`
	Email    string `json:"email"`
}

// EmailResponse is the response for the custom sendemail function.
type EmailResponse struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

// City represents a city and its coordinates.
type City struct {
	ID   string  `json:"id,omitempty"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// CityID is used to capture the city id in relationships.
type CityID struct {
	ID string `json:"id"`
}

// Place contains the place data points captured from the API.
type Place struct {
	ID               string   `json:"id,omitempty"`
	Category         string   `json:"category"`
	CityID           CityID   `json:"city"`
	PlaceID          string   `json:"place_id"`
	CityName         string   `json:"city_name"`
	Name             string   `json:"name"`
	Address          string   `json:"address"`
	Lat              float64  `json:"lat"`
	Lng              float64  `json:"lng"`
	LocationType     []string `json:"location_type"`
	AvgUserRating    float32  `json:"avg_user_rating"`
	NumberOfRatings  int      `json:"no_user_rating"`
	GmapsURL         string   `json:"gmaps_url"`
	PhotoReferenceID string   `json:"photo_id"`
}

// Advisory contains the travel advisory result captured for a city.
type Advisory struct {
	ID          string  `json:"id,omitempty"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Continent   string  `json:"continent"`
	Score       float64 `json:"score"`
	LastUpdated string  `json:"last_updated"`
	Message     string  `json:"message"`
	Source      string  `json:"source"`
}

// Weather contains the weather data points captured from the API.
type Weather struct {
	ID            string  `json:"id,omitempty"`
	CityName      string  `json:"city_name"`
	Visibility    string  `json:"visibility"`
	Desc          string  `json:"description"`
	Temp          float64 `json:"temp"`
	FeelsLike     float64 `json:"feels_like"`
	MinTemp       float64 `json:"temp_min"`
	MaxTemp       float64 `json:"temp_max"`
	Pressure      int     `json:"pressure"`
	Humidity      int     `json:"humidity"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection int     `json:"wind_direction"`
	Sunrise       int     `json:"sunrise"`
	Sunset        int     `json:"sunset"`
}
