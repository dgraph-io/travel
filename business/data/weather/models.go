package weather

// Info contains the weather data points captured from the API.
type Info struct {
	ID            string  `json:"id,omitempty"`
	City          City    `json:"city"`
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

// City is used to capture the city id in relationships.
type City struct {
	ID string `json:"id"`
}

type addResult struct {
	AddWeather struct {
		Weather []struct {
			ID string `json:"id"`
		} `json:"weather"`
	} `json:"addWeather"`
}

func (addResult) document() string {
	return `{
		weather {
			id
		}
	}`
}

type deleteResult struct {
	DeleteWeather struct {
		Msg     string
		NumUids int
	} `json:"deleteWeather"`
}

func (deleteResult) document() string {
	return `{
		msg,
		numUids,
	}`
}
