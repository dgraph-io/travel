$.ajaxSetup({
    contentType: "application/json; charset=utf-8"
});

function loadData() {
    drawchart(document.getElementById("cityselection").value);
    const d = { type: "city" };
    showInfo(d);
}

function circleMouseOver(d, index, circles) {
    const circle = circles[index];
    const radius = parseInt(circle.getAttribute("r"));
    circle.setAttribute("r", radius+4);
}

function circleMouseOut(d, index, circles) {
    const circle = circles[index];
    const radius = parseInt(circle.getAttribute("rorg"));
    circle.setAttribute("r", radius);
}

function showInfo(d, index, circles) {
    const cell = document.getElementById("data");
    const name = document.getElementById("cityselection").value;
    switch (d.type) {
        case "city":
            $.post(Dgraph,
            queryCity(name),
            function (o, status) {
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<div class=\"bluedot\"></div><div class=\"dotlabel\">City</div>";
                innerHTML += "<table><tr><td><dl>";              
                innerHTML += "<dt>ID: " + o.data.queryCity[0].id + "</dt>";
                innerHTML += "<dt>Name: " + o.data.queryCity[0].name + "</dt>";
                innerHTML += "<dt>Lat: " + o.data.queryCity[0].lat + "</dt>";
                innerHTML += "<dt>Lng: " + o.data.queryCity[0].lng + "</dt>";
                innerHTML += "</dl></td></tr></table>";
                cell.innerHTML = innerHTML;
            });
            break;
        case "advisory":
            $.post(Dgraph,
            queryAdvisory(name),
            function(o, status){
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<div class=\"reddot\"></div><div class=\"dotlabel\">Advisory</div>";
                innerHTML += "<table><tr><td><dl>";
                innerHTML += "<dt>ID: " + o.data.queryCity[0].advisory.id + "</dt>";
                innerHTML += "<dt>Country: " + o.data.queryCity[0].advisory.country + "</dt>";
                innerHTML += "<dt>Country Code: " + o.data.queryCity[0].advisory.country_code + "</dt>";
                innerHTML += "<dt>Continent: " + o.data.queryCity[0].advisory.continent + "</dt>";
                innerHTML += "<dt>Score: " + o.data.queryCity[0].advisory.score + "</dt>";
                innerHTML += "<dt>Message: " + o.data.queryCity[0].advisory.message + "</dt>";
                innerHTML += "</dl></td></tr></table>";
                cell.innerHTML = innerHTML;
            });
            break;
        case "weather":
            $.post(Dgraph,
            queryWeather(name),
            function(o, status){
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<div class=\"orangedot\"></div><div class=\"dotlabel\">Weather</div>";
                innerHTML += "<table><tr><td><dl>";
                innerHTML += "<dt>ID: " + o.data.queryCity[0].weather.id + "</dt>";
                innerHTML += "<dt>City Name: " + o.data.queryCity[0].weather.city_name + "</dt>";
                innerHTML += "<dt>Visibility: " + o.data.queryCity[0].weather.visibility + "</dt>";
                innerHTML += "<dt>Description: " + o.data.queryCity[0].weather.description + "</dt>";
                innerHTML += "<dt>Temp: " + convertKelvin(o.data.queryCity[0].weather.temp) + "F</dt>";
                innerHTML += "<dt>Feels Like: " + convertKelvin(o.data.queryCity[0].weather.feels_like) + "F</dt>";
                innerHTML += "<dt>Min Temp: " + convertKelvin(o.data.queryCity[0].weather.temp_min) + "F</dt>";
                innerHTML += "<dt>Max Temp: " + convertKelvin(o.data.queryCity[0].weather.temp_max) + "F</dt>";
                innerHTML += "<dt>Pressure: " + o.data.queryCity[0].weather.pressure + "</dt>";
                innerHTML += "<dt>Humidity: " + o.data.queryCity[0].weather.humidity + "</dt>";
                innerHTML += "<dt>Wind Speed: " + o.data.queryCity[0].weather.wind_speed + "</dt>";
                innerHTML += "<dt>Wind Direction: " + o.data.queryCity[0].weather.wind_direction + "</dt>";
                innerHTML += "</dl></td></tr></table>";
                cell.innerHTML = innerHTML;
            });
            break;
        case "place":
            $.post(Dgraph,
            queryPlaceByCategory(name, d.id),
            function(o, status){
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<div class=\"dot\" style=\"background-color:" + d.color + "\"></div><div class=\"dotlabel\">" + d.id + "</div>";
                innerHTML += "<table>";
                for (i = 0; i < o.data.queryCity[0].places.length; i++) {
                    innerHTML += "<tr><td><dl>";
                    innerHTML += "<dt>ID: " + o.data.queryCity[0].places[i].id + "</dt>";
                    innerHTML += "<dt>City: " + o.data.queryCity[0].places[i].city_name + "</dt>";
                    innerHTML += "<dt>Name: " + o.data.queryCity[0].places[i].name.split(":")[0] + "</dt>";
                    innerHTML += "<dt>Address: " + o.data.queryCity[0].places[i].address + "</dt>";
                    innerHTML += "<dt>Avg User Rating: " + o.data.queryCity[0].places[i].avg_user_rating + "</dt>";
                    innerHTML += "</dl></td></tr>";
                }
                innerHTML += "</table>";
                cell.innerHTML = innerHTML;
            });
            break;
        default:
            $.post(Dgraph,
            queryPlaceByName(d.id),
            function(o, status){
                if (typeof o.data === "undefined") {
                    cell.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<div class=\"dot\" style=\"background-color:" + d.color + "\"></div><div class=\"dotlabel\">" + d.type + "</div>";
                innerHTML += "<table><tr><td><dl>";
                innerHTML += "<dt>ID: " + o.data.queryPlace[0].id + "</dt>";
                innerHTML += "<dt>City: " + o.data.queryPlace[0].city_name + "</dt>";
                innerHTML += "<dt>Name: " + o.data.queryPlace[0].name.split(":")[0] + "</dt>";
                innerHTML += "<dt>Address: " + o.data.queryPlace[0].address + "</dt>";
                innerHTML += "<dt>Avg User Rating: " + o.data.queryPlace[0].avg_user_rating + "</dt>";
                innerHTML += "</dl></td></tr></table>";
                cell.innerHTML = innerHTML;
            });
            break;
    }
}

function convertKelvin(k) {
    const num = k * 9 / 5 - 459.67;
    return Math.round((num + Number.EPSILON) * 100) / 100;
}