$.ajaxSetup({
    contentType: "application/json; charset=utf-8"
});

function loadData() {
    drawchart(document.getElementById("cityselection").value);
    const d = { type: "city" };
    showInfo(d);
}

function showTab(which) {
    const data = document.querySelector("div.databox");
    const code = document.querySelector("div.codebox");
    if (which == "data") {
        code.style.display = "none";
        data.style.display = "block";
        return;
    }
    code.style.display = "block";
    data.style.display = "none";
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

function showQueryResponse(query, resp) {
    let display = "<b>Query:</b><br /><br />" + query.replace(/\\n/g, "<br />");
    display += "<br /><br /><b>Response:</b><br /><br />" + JSON.stringify(resp);
    return display;
}

function showInfo(d, index, circles) {
    const data = document.getElementById("data");
    const code = document.getElementById("code");
    const name = document.getElementById("cityselection").value;

    switch (d.type) {
        
        case "city":
            var query = queryCity(name);
            $.post(Dgraph, query, function (o, status) {
                if (typeof o.data === "undefined") {
                    data.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerData = "<div class=\"bluedot\"></div><div class=\"dotlabel\">City</div>";
                innerData += "<table><tr><td><dl>";              
                innerData += "<dt>ID: " + o.data.queryCity[0].id + "</dt>";
                innerData += "<dt>Name: " + o.data.queryCity[0].name + "</dt>";
                innerData += "<dt>Lat: " + o.data.queryCity[0].lat + "</dt>";
                innerData += "<dt>Lng: " + o.data.queryCity[0].lng + "</dt>";
                innerData += "</dl></td></tr></table>";
                data.innerHTML = innerData;                
                code.innerHTML = showQueryResponse(query, o);
            });
            break;

        case "advisory":
            var query = queryAdvisory(name);
            $.post(Dgraph, query, function(o, status) {
                if (typeof o.data === "undefined") {
                    data.innerText = "ERROR: " + o.errors[0].message;
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
                data.innerHTML = innerHTML;
                code.innerHTML = showQueryResponse(query, o);
            });
            break;

        case "weather":
            var query = queryWeather(name);
            $.post(Dgraph, query, function(o, status) {
                if (typeof o.data === "undefined") {
                    data.innerText = "ERROR: " + o.errors[0].message;
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
                data.innerHTML = innerHTML;
                code.innerHTML = showQueryResponse(query, o);
            });
            break;

        case "place":
            var query = queryPlaceByCategory(name, d.id);
            $.post(Dgraph, query, function(o, status) {
                if (typeof o.data === "undefined") {
                    data.innerText = "ERROR: " + o.errors[0].message;
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
                data.innerHTML = innerHTML;
                code.innerHTML = showQueryResponse(query, o);
            });
            break;

        default:
            var query = queryPlaceByName(d.id);
            $.post(Dgraph, query, function(o, status) {
                if (typeof o.data === "undefined") {
                    data.innerText = "ERROR: " + o.errors[0].message;
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
                data.innerHTML = innerHTML;
                code.innerHTML = showQueryResponse(query, o);
            });
            break;
    }
}

function convertKelvin(k) {
    const num = k * 9 / 5 - 459.67;
    return Math.round((num + Number.EPSILON) * 100) / 100;
}