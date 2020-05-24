$.ajaxSetup({
    contentType: "application/json; charset=utf-8"
});

function loadData() {
    drawchart(document.getElementById("cityselection").value);
}

function convertKelvin(k) {
    let num = k * 9/5 - 459.67
    return Math.round((num + Number.EPSILON) * 100) / 100
}

function showInfo(d, i) {
    let cell = document.getElementById("data");
    let name = document.getElementById("cityselection").value;
    switch (d.type) {
        case "city":
            $.post(Dgraph,
            '{"query":"query { queryCity(filter: { name: { eq: \\"' + name + '\\" } }) { id name lat lng } }","variables":null}',
            function (o, status) {
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<table width=\"70%\">";
                innerHTML += "<tr><td><div class=\"bluedot\"></div></td><td>City</td></tr>";
                innerHTML += "<tr><td>ID:</td><td>" + o.data.queryCity[0].id + "</td></tr>";
                innerHTML += "<tr><td>Name:</td><td>" + o.data.queryCity[0].name + "</td></tr>";
                innerHTML += "<tr><td>Lat:</td><td>" + o.data.queryCity[0].lat + "</td></tr>";
                innerHTML += "<tr><td>Lng:</td><td>" + o.data.queryCity[0].lng + "</td></tr>";
                innerHTML += "</table>";
                cell.innerHTML = innerHTML;
            });
            break;
        case "advisory":
            $.post(Dgraph,
            '{"query":"query { queryCity(filter: { name: { eq: \\"' + name + '\\" } }) { advisory { id continent country country_code last_updated message score source }} }","variables":null}',
            function(o, status){
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<table width=\"70%\">";
                innerHTML += "<tr><td><div class=\"reddot\"></div></td><td>Advisory</td></tr>";
                innerHTML += "<tr><td>ID:</td><td>" + o.data.queryCity[0].advisory.id + "</td></tr>";
                innerHTML += "<tr><td>Country:</td><td>" + o.data.queryCity[0].advisory.country + "</td></tr>";
                innerHTML += "<tr><td>Country Code:</td><td>" + o.data.queryCity[0].advisory.country_code + "</td></tr>";
                innerHTML += "<tr><td>Continent:</td><td>" + o.data.queryCity[0].advisory.continent + "</td></tr>";
                innerHTML += "<tr><td>Score:</td><td>" + o.data.queryCity[0].advisory.score + "</td></tr>";
                innerHTML += "<tr><td>Message:</td><td>" + o.data.queryCity[0].advisory.message + "</td></tr>";
                innerHTML += "</table>";
                cell.innerHTML = innerHTML;
            });
            break;
        case "weather":
            $.post(Dgraph,
            '{"query":"query { queryCity(filter: { name: { eq: \\"' + name + '\\" } }) { weather { id city_name description feels_like humidity pressure sunrise sunset temp temp_min temp_max visibility wind_direction wind_speed }} }","variables":null}',
            function(o, status){
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<table width=\"70%\">";
                innerHTML += "<tr><td><div class=\"orangedot\"></div></td><td>Weather</td></tr>";
                innerHTML += "<tr><td>ID:</td><td>" + o.data.queryCity[0].weather.id + "</td></tr>";
                innerHTML += "<tr><td>City Name:</td><td>" + o.data.queryCity[0].weather.city_name + "</td></tr>";
                innerHTML += "<tr><td>Visibility:</td><td>" + o.data.queryCity[0].weather.visibility + "</td></tr>";
                innerHTML += "<tr><td>Description:</td><td>" + o.data.queryCity[0].weather.description + "</td></tr>";
                innerHTML += "<tr><td>Temp:</td><td>" + convertKelvin(o.data.queryCity[0].weather.temp) + "F</td></tr>";
                innerHTML += "<tr><td>Feels Like:</td><td>" + convertKelvin(o.data.queryCity[0].weather.feels_like) + "F</td></tr>";
                innerHTML += "<tr><td>Min Temp:</td><td>" + convertKelvin(o.data.queryCity[0].weather.temp_min) + "F</td></tr>";
                innerHTML += "<tr><td>Max Temp:</td><td>" + convertKelvin(o.data.queryCity[0].weather.temp_max) + "F</td></tr>";
                innerHTML += "<tr><td>Pressure:</td><td>" + o.data.queryCity[0].weather.pressure + "</td></tr>";
                innerHTML += "<tr><td>Humidity:</td><td>" + o.data.queryCity[0].weather.humidity + "</td></tr>";
                innerHTML += "<tr><td>Wind Speed:</td><td>" + o.data.queryCity[0].weather.wind_speed + "</td></tr>";
                innerHTML += "<tr><td>Wind Direction:</td><td>" + o.data.queryCity[0].weather.wind_direction + "</td></tr>";
                innerHTML += "</table>";
                cell.innerHTML = innerHTML;
            });
            break;
        case "place":
            $.post(Dgraph,
            '{"query":"query { queryPlace(filter: { category: { eq: \\"' + d.id + '\\" } }) { id address avg_user_rating category city{ id } city_name gmaps_url lat lng location_type name no_user_rating place_id photo_id } }","variables":null}',
            function(o, status){
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<table width=\"70%\">";
                innerHTML += "<tr><td><div class=\"dot\" style=\"background-color:" + d.color + "\";></div></td><td>" + d.id + "</td></tr>";
                for (i = 0; i < o.data.queryPlace.length; i++) {
                    innerHTML += "<tr><td>ID:</td><td>" + o.data.queryPlace[i].id + "</td></tr>";
                    innerHTML += "<tr><td>City:</td><td>" + o.data.queryPlace[i].city_name + "</td></tr>";
                    innerHTML += "<tr><td>Name:</td><td>" + o.data.queryPlace[i].name.split(":")[0] + "</td></tr>";
                    innerHTML += "<tr><td>Address:</td><td>" + o.data.queryPlace[i].address + "</td></tr>";
                    innerHTML += "<tr><td>Avg User Rating:</td><td>" + o.data.queryPlace[i].avg_user_rating + "</td></tr>";
                    innerHTML += "<tr><td colspan=\"2\">==============================</td></tr>";
                }
                innerHTML += "</table>";
                cell.innerHTML = innerHTML;
            });
            break;
        default:
            $.post(Dgraph,
            '{"query":"query { queryPlace(filter: { name: { eq: \\"' + d.id + '\\" } }) { id address avg_user_rating category city{ id } city_name gmaps_url lat lng location_type name no_user_rating place_id photo_id } }","variables":null}',
            function(o, status){
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<table width=\"70%\">";
                innerHTML += "<tr><td><div class=\"dot\" style=\"background-color:" + d.color + "\";></div></td><td>" + d.type +"</td></tr>";
                innerHTML += "<tr><td>ID:</td><td>" + o.data.queryPlace[0].id + "</td></tr>";
                innerHTML += "<tr><td>City:</td><td>" + o.data.queryPlace[0].city_name + "</td></tr>";
                innerHTML += "<tr><td>Name:</td><td>" + o.data.queryPlace[0].name.split(":")[0] + "</td></tr>";
                innerHTML += "<tr><td>Address:</td><td>" + o.data.queryPlace[0].address + "</td></tr>";
                innerHTML += "<tr><td>Avg User Rating:</td><td>" + o.data.queryPlace[0].avg_user_rating + "</td></tr>";
                innerHTML += "</table>";
                cell.innerHTML = innerHTML;
            });
            break;
    }
}