function querySchema() {
    return JSON.stringify({
        query: `query { getGQLSchema { schema }}`,
        variables: null
    });
}

function queryCityNames(cityName) {
    return JSON.stringify({
        query: `query {
            queryCity(filter: {}) {
                name
            }}`,
        variables: null
    });
}

function queryCity(cityName) {
    return JSON.stringify({
        query: `query {
            queryCity(filter: { name: { eq: "` + cityName + `" } }) {
                id
                name
                lat
                lng
            }}`,
        variables: null
    });
}

function queryCityPlaces(cityName) {
    return JSON.stringify({
        query: `query {
            queryCity(filter: { name: { eq: "` + cityName + `" } }) {
                places {
                    lat
                    lng
                    name
                }
            }}`,
        variables: null
    });
}

function queryAdvisory(cityName) {
    return JSON.stringify({
        query: `query {
            queryCity(filter: { name: { eq: "` + cityName + `" } }) {
                advisory {
                    id
                    continent
                    country
                    country_code
                    last_updated
                    message
                    score
                    source
                }
            }}`,
        variables: null
    });
}

function queryWeather(cityName) {
    return JSON.stringify({
        query: `query {
            queryCity(filter: { name: { eq: "` + cityName + `" } }) {
                weather {
                    id
                    city_name
                    description
                    feels_like
                    humidity
                    pressure
                    sunrise
                    sunset
                    temp
                    temp_min
                    temp_max
                    visibility
                    wind_direction
                    wind_speed
                }
            }
        }`,
        variables: null
    });
}

function queryPlaceByCategory(cityName, category) {
    return JSON.stringify({
        query: `query {
            queryCity(filter: { name: { eq: "` + cityName + `" } }) {
                places(order: { asc: name }, filter: {category: {eq: "` + category + `" } }) {
                    id
                    address
                    avg_user_rating
                    category
                    city{
                        id
                    }
                    city_name
                    gmaps_url
                    lat
                    lng
                    location_type
                    name
                    no_user_rating
                    place_id
                    photo_id
                }
            }
        }`,
        variables: null
    });
}

function queryPlaceByName(placeName) {
    return JSON.stringify({
        query: `query {
            queryPlace(filter: { name: { eq: "` + placeName + `" } }) {
                id
                address
                avg_user_rating
                category
                city{
                    id
                }
                city_name
                gmaps_url
                lat
                lng
                location_type
                name
                no_user_rating
                place_id
                photo_id
            }
        }`,
        variables: null
    });
}

function queryUploadFeed(countryCode, cityName, lat, lng) {
    return JSON.stringify({
        query: `query {
            uploadFeed(countryCode: "` + countryCode + `", cityName: "` + cityName + `", lat: ` + lat + `, lng: ` + lng + `) {
                country_code
                city_name
                lat
                lng
	            message
            }
        }`,
        variables: null
    })
}
