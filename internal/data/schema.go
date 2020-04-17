package data

var gQLSchema = `type Weather {
	id: ID!
	city_name: String!
	visibility: String
	description: String
	temp: Float
	feels_like: Float
	temp_min: Float
	temp_max: Float
	pressure: Int
	humidity: Int
	wind_speed: Float
	wind_direction: Int
	sunrise: Int
	sunset: Int
}

type City {
	id: ID!
	name: String @search(by: [term])
	lat: Float!
	lng: Float!
	weather: Weather
}`

var goSchema = []Schema{
	{"City.lat", "float", false, nil, false},
	{"City.lng", "float", false, nil, false},
	{"City.name", "string", true, []string{"term"}, false},
	{"City.weather", "uid", false, nil, false},

	{"Weather.city_name", "string", false, nil, false},
	{"Weather.description", "string", false, nil, false},
	{"Weather.feels_like", "float", false, nil, false},
	{"Weather.humidity", "int", false, nil, false},
	{"Weather.pressure", "int", false, nil, false},
	{"Weather.sunrise", "int", false, nil, false},
	{"Weather.sunset", "int", false, nil, false},
	{"Weather.temp", "float", false, nil, false},
	{"Weather.temp_max", "float", false, nil, false},
	{"Weather.temp_min", "float", false, nil, false},
	{"Weather.visibility", "string", false, nil, false},
	{"Weather.wind_direction", "int", false, nil, false},
	{"Weather.wind_speed", "float", false, nil, false},

	{"dgraph.graphql.schema", "string", false, nil, false},
	{"dgraph.type", "string", true, []string{"exact"}, false},
}

// GrapQLSchema is used to keep the guidelines set for access
// a package level variable only in the file it is defined in.
func GrapQLSchema() (string, []Schema) {
	return gQLSchema, goSchema
}

// Schema represents information per predicate set in the schema.
type Schema struct {
	Predicate string   `json:"predicate"`
	Type      string   `json:"type"`
	Index     bool     `json:"index"`
	Tokenizer []string `json:"tokenizer"`
	Upsert    bool     `json:"upsert"`
}
