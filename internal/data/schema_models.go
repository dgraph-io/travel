package data

// This is the schema for the application. This could be kept in a file
// and maintained for wider use. In these cases I would use gogenerate
// to hardcode the contents into the binary.
const gQLSchema = `
enum Roles {
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
	id: ID!
	email: String
	subject: String
	err: String
}

type Query{
	sendEmail(email: String!, subject: String!): EmailResponse @custom(http:{
		url: "http://0.0.0.0:3000/v1/email",
		method: "POST",
		body: "{ email: $email, subject: $subject }"
	})
}`

/*
# Dgraph.Authorization X-My-App-Auth {"roles":["ADMIN","MUTATE","QUERY"]} RS256 "-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAr6F9KGsyH6kbGlLiLSehw6eSeIxyV1lesOoYP0qDPCBLPxPlRIyS
GP1LOPSXOz1OaJIU5cHzK00dtvEnW2HCOkFpVxYBx0l0Ro/m+2rOxXLJLJrd4+Dw
fCdT91TtAHgaklkAYfMbxctdpXXFu3FHjMRW/lopaTDx6w/dnrOiw5P86mxgCABV
LRBvQ1vDxp57lwVDZ55xu8UxXtw30ukA7nioiiu/+MslxPbnsjWAvyfbMpjUz9S3
Mbe/FeMif5t1Cr05+M07r9ZN54/MX8zZJ651OFK96HpiIkxwK06jccprmCixYsYt
BIwXtt2J8WJ13gfqViFMj9MpdMMwVjwQQQIDAQAB
-----END RSA PUBLIC KEY-----"
*/
