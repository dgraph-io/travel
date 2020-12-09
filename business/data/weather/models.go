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

// =============================================================================

type id struct {
	Resp struct {
		Entities []struct {
			ID string `json:"id"`
		} `json:"entities"`
	} `json:"resp"`
}

func (id) document() string {
	return `{
		entities: weather {
			id
		}
	}`
}

type result struct {
	Resp struct {
		Msg     string
		NumUids int
	} `json:"resp"`
}

func (result) document() string {
	return `{
		msg,
		numUids,
	}`
}
