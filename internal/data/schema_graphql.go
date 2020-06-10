package data

var schema = struct {
	publicKey string
	document  string
}{
	publicKey: _publicKey,
	document:  _document,
}

var _publicKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnZ/BW/tuLr0uxZFw1Q5m
P1JpIksU46o+kIaqIXZjSAduma18m+oSgd1L19Fs9otAjfAlkyU8HF1hJNj/PVv8
MY72vhIWv60xBB4caXuLmflAiJEtvxHfw3WtVR9npQqEowcwrsf7MSSfdHwM4S+F
bMmcl/mE9c7DUrYJBUgu1IbdI7vrEoPE65GFafjZQHkPLUX8OaRXOt4rkT6HfYv+
XqaCs6Ie+dt6xL5HiQpO90/89CAJhi2q8AXvhfxqCVVfLxxd3jNJVq2olkCOLJRE
uJ29Bb460yKOAiDigEUobUpmvT6ggUZNrX71yP0GZxQFBhq9j1IRgPVg4CDA0Pw5
FQIDAQAB
-----END RSA PUBLIC KEY-----`

var _document = `enum Roles {
	ADMIN
	EMAIL
	MUTATE
	QUERY
}

type User {
	id: ID!
	email: String! @search(by: [exact])
	name: String!
	roles: [Roles]!
	password_hash: String!
	date_created: DateTime!
	date_updated: DateTime!
}

type City {
	id: ID!
	name: String! @search(by: [exact])
	lat: Float!
	lng: Float!
	places: [Place] @hasInverse(field: city)
	advisory: Advisory
	weather: Weather
}

type Advisory {
	id: ID!
	continent: String!
	country: String!
	country_code: String!
	last_updated: String
	message: String
	score: Float!
	source: String
}

type Place {
	id: ID!
	address: String
	avg_user_rating: Float
	category: String @search(by: [exact])
	city: City!
	city_name: String!
	gmaps_url: String
	lat: Float!
	lng: Float!
	location_type: [String]
	name: String! @search(by: [exact])
	no_user_rating: Int
	place_id: String!
	photo_id: String
}

type Weather {
	id: ID!
	city_name: String!
	description: String
	feels_like: Float
	humidity: Int
	pressure: Int
	sunrise: Int
	sunset: Int
	temp: Float
	temp_min: Float
	temp_max: Float
	visibility: String
	wind_direction: Int
	wind_speed: Float
}

type EmailResponse @remote {
	user_id: ID
	message: String
}

type Query{
	sendEmail(userID: String!, nodeType: String!, nodeID: String!, email: String!): EmailResponse @custom(http:{
		url: "{{.SendEmailURL}}",
		method: "POST",
		body: "{userid: $userID, nodetype: $nodeType, nodeid: $nodeID, email: $email}"
	})
}

# Dgraph.Authorization X-Travel-Auth Auth RS256 "{{.PublicKey}}"
`
