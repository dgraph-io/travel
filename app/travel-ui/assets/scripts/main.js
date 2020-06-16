var map;
var markers = [];
var currentMap;

$.ajaxSetup({
    contentType: "application/json; charset=utf-8",
    beforeSend: function (xhr) {
        if (AuthHeaderName != "") {
            xhr.setRequestHeader(AuthHeaderName, AuthToken);
        }
    }
});

function OnLoad() {
    loadCitySelections(); 
}

function loadCitySelections() {
    $('#cityselection').children().remove().end();

    var query = queryCityNames();
    $.post(Dgraph, query, function (o, status) {
        if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {
            window.alert("ERROR: " + o.errors[0].message);
            return;
        }

        const selection = document.getElementById("cityselection");
        for (var i = 0; i < o.data.queryCity.length; i++) {
            selection.options[i] = new Option(o.data.queryCity[i].name, o.data.queryCity[i].name);
        }

        loadData();
    });
}

function loadData() {
    document.getElementById("node").innerHTML = "";
    document.getElementById("query").innerHTML = "";
    showInfoTab("node");
    showGraphTab("graph");

    let city = document.getElementById("cityselection").value;
    drawchart(city);
    const d = { type: "city" };
    showNodeData(d);
}

function initMap() {
    map = new google.maps.Map(document.getElementById("map"), {
        center: {lat: 25.7617, lng: -80.1918},
        zoom: 8
    });
}

function loadMap() {
    const name = document.getElementById("cityselection").value;
    if (name == currentMap) {
        return;
    }
    currentMap = name;
    var query = queryCityPlaces(name);
    $.post(Dgraph, query, function (o, status) {
        if (typeof o.data === "undefined") {
            nodeBox.innerText = "ERROR: " + o.errors[0].message;
            return;
        }
        for (var i = 0; i < markers.length; i++) {
            markers[i].setMap(null);
        }
        markers = [];
        var bounds = new google.maps.LatLngBounds();
        for (i = 0; i < o.data.queryCity[0].places.length; i++) {
            var name = o.data.queryCity[0].places[i].name.split(":")[0];
            var latLng = new google.maps.LatLng(o.data.queryCity[0].places[i].lat, o.data.queryCity[0].places[i].lng);
            var marker = new google.maps.Marker({
                position: latLng,
                title: name,
                map: map
            });
            markers.push(marker);
            bounds.extend(latLng);
        }
        map.fitBounds(bounds);
    });
}

function loadSchema() {
    const schemaBox = document.getElementById("schema");
    let dgraph = Dgraph.replace("graphql", "admin")
    schema.innerHTML = "fetching schema ...";

    $.post(dgraph, querySchema(), function (o, status) {
        if (typeof o.data === "undefined") {
            schema.innerText = "ERROR: " + o.errors[0].message;
            return;
        }
        let doc = JSON.stringify(o).replace(/\\n/g, "<br />");
        doc = doc.replace(/\\t/g, "&nbsp;&nbsp;&nbsp;&nbsp;");
        schemaBox.innerHTML = "<b>Schema:</b><br /><br />" + doc;
    });
}

function showInfoTab(which) {
    const nodeBox = document.querySelector("div.nodebox");
    const queryBox = document.querySelector("div.querybox");
    const schemaBox = document.querySelector("div.schemabox");

    const nodeBut = document.getElementById("nodebutton");
    const queryBut = document.getElementById("querybutton");
    const schemaBut = document.getElementById("schemabutton");

    switch (which) {
        case "node":
            nodeBox.style.display = "block";
            queryBox.style.display = "none";
            schemaBox.style.display = "none";
            nodeBut.style.backgroundColor = "#d9d8d4";
            queryBut.style.backgroundColor = "#faf9f5";
            schemaBut.style.backgroundColor = "#faf9f5";
            break;
        case "query":
            nodeBox.style.display = "none";
            queryBox.style.display = "block";
            schemaBox.style.display = "none";
            nodeBut.style.backgroundColor = "#faf9f5";
            queryBut.style.backgroundColor = "#d9d8d4";
            schemaBut.style.backgroundColor = "#faf9f5";
            break;
        case "schema":
            nodeBox.style.display = "none";
            queryBox.style.display = "none";
            schemaBox.style.display = "block";
            nodeBut.style.backgroundColor = "#faf9f5";
            queryBut.style.backgroundColor = "#faf9f5";
            schemaBut.style.backgroundColor = "#d9d8d4";
            loadSchema();
            break;
    }
}

function showGraphTab(which) {
    const graphBox = document.querySelector("div.graphbox");
    const mapBox = document.querySelector("div.mapbox");

    const graphBut = document.getElementById("graphbutton");
    const mapBut = document.getElementById("mapbutton");

    switch (which) {
        case "graph":
            graphBox.style.display = "block";
            mapBox.style.display = "none";
            graphBut.style.backgroundColor = "#d9d8d4";
            mapBut.style.backgroundColor = "#faf9f5";
            break;
        case "map":
            graphBox.style.display = "none";
            mapBox.style.display = "block";
            graphBut.style.backgroundColor = "#faf9f5";
            mapBut.style.backgroundColor = "#d9d8d4";
            loadMap();
            break;
    }
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

function showNodeData(d, index, circles) {
    const nodeBox = document.getElementById("node");
    const queryBox = document.getElementById("query");
    const name = document.getElementById("cityselection").value;

    switch (d.type) {
        case "city":
            var query = queryCity(name);
            $.post(Dgraph, query, function (o, status) {
                 if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {
                    nodeBox.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerData = "<div class=\"bluedot\"></div><div class=\"dotlabel\">City</div>";
                innerData += "<table><tr><td><dl>";              
                innerData += "<dt>ID: " + o.data.queryCity[0].id + "</dt>";
                innerData += "<dt>Name: " + o.data.queryCity[0].name + "</dt>";
                innerData += "<dt>Lat: " + o.data.queryCity[0].lat + "</dt>";
                innerData += "<dt>Lng: " + o.data.queryCity[0].lng + "</dt>";
                innerData += "</dl></td></tr></table>";
                nodeBox.innerHTML = innerData;
                queryBox.innerHTML = showQueryResponse(query, o);
            });
            break;

        case "advisory":
            var query = queryAdvisory(name);
            $.post(Dgraph, query, function(o, status) {
                if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {
                    nodeBox.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerData = "<div class=\"reddot\"></div><div class=\"dotlabel\">Advisory</div>";
                innerData += "<table><tr><td><dl>";
                innerData += "<dt>ID: " + o.data.queryCity[0].advisory.id + "</dt>";
                innerData += "<dt>Country: " + o.data.queryCity[0].advisory.country + "</dt>";
                innerData += "<dt>Country Code: " + o.data.queryCity[0].advisory.country_code + "</dt>";
                innerData += "<dt>Continent: " + o.data.queryCity[0].advisory.continent + "</dt>";
                innerData += "<dt>Score: " + o.data.queryCity[0].advisory.score + "</dt>";
                innerData += "<dt>Message: " + o.data.queryCity[0].advisory.message + "</dt>";
                innerData += "</dl></td></tr></table>";
                nodeBox.innerHTML = innerData;
                queryBox.innerHTML = showQueryResponse(query, o);
            });
            break;

        case "weather":
            var query = queryWeather(name);
            $.post(Dgraph, query, function(o, status) {
                if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {
                    nodeBox.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerData = "<div class=\"orangedot\"></div><div class=\"dotlabel\">Weather</div>";
                innerData += "<table><tr><td><dl>";
                innerData += "<dt>ID: " + o.data.queryCity[0].weather.id + "</dt>";
                innerData += "<dt>City Name: " + o.data.queryCity[0].weather.city_name + "</dt>";
                innerData += "<dt>Visibility: " + o.data.queryCity[0].weather.visibility + "</dt>";
                innerData += "<dt>Description: " + o.data.queryCity[0].weather.description + "</dt>";
                innerData += "<dt>Temp: " + convertKelvin(o.data.queryCity[0].weather.temp) + "F</dt>";
                innerData += "<dt>Feels Like: " + convertKelvin(o.data.queryCity[0].weather.feels_like) + "F</dt>";
                innerData += "<dt>Min Temp: " + convertKelvin(o.data.queryCity[0].weather.temp_min) + "F</dt>";
                innerData += "<dt>Max Temp: " + convertKelvin(o.data.queryCity[0].weather.temp_max) + "F</dt>";
                innerData += "<dt>Pressure: " + o.data.queryCity[0].weather.pressure + "</dt>";
                innerData += "<dt>Humidity: " + o.data.queryCity[0].weather.humidity + "</dt>";
                innerData += "<dt>Wind Speed: " + o.data.queryCity[0].weather.wind_speed + "</dt>";
                innerData += "<dt>Wind Direction: " + o.data.queryCity[0].weather.wind_direction + "</dt>";
                innerData += "</dl></td></tr></table>";
                nodeBox.innerHTML = innerData;
                queryBox.innerHTML = showQueryResponse(query, o);
            });
            break;

        case "place":
            var query = queryPlaceByCategory(name, d.id);
            $.post(Dgraph, query, function (o, status) {
                if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {
                    nodeBox.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerData = "<div class=\"dot\" style=\"background-color:" + d.color + "\"></div><div class=\"dotlabel\">" + d.id + "</div>";
                innerData += "<table>";
                for (i = 0; i < o.data.queryCity[0].places.length; i++) {
                    innerData += "<tr><td><dl>";
                    innerData += "<dt>ID: " + o.data.queryCity[0].places[i].id + "</dt>";
                    innerData += "<dt>City: " + o.data.queryCity[0].places[i].city_name + "</dt>";
                    innerData += "<dt>Name: " + o.data.queryCity[0].places[i].name.split(":")[0] + "</dt>";
                    innerData += "<dt>Address: " + o.data.queryCity[0].places[i].address + "</dt>";
                    innerData += "<dt>Lat: " + o.data.queryCity[0].places[i].lat + "</dt>";
                    innerData += "<dt>Lng: " + o.data.queryCity[0].places[i].lng + "</dt>";
                    innerData += "<dt>Avg User Rating: " + o.data.queryCity[0].places[i].avg_user_rating + "</dt>";
                    innerData += "</dl></td></tr>";
                }
                innerData += "</table>";
                nodeBox.innerHTML = innerData;
                queryBox.innerHTML = showQueryResponse(query, o);
            });
            break;

        default:
            var query = queryPlaceByName(d.id);
            $.post(Dgraph, query, function (o, status) {
                if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {
                    nodeBox.innerText = "ERROR: " + o.errors[0].message;
                    return;
                }
                let innerHTML = "<div class=\"dot\" style=\"background-color:" + d.color + "\"></div><div class=\"dotlabel\">" + d.type + "</div>";
                innerHTML += "<table><tr><td><dl>";
                innerHTML += "<dt>ID: " + o.data.queryPlace[0].id + "</dt>";
                innerHTML += "<dt>City: " + o.data.queryPlace[0].city_name + "</dt>";
                innerHTML += "<dt>Name: " + o.data.queryPlace[0].name.split(":")[0] + "</dt>";
                innerHTML += "<dt>Address: " + o.data.queryPlace[0].address + "</dt>";
                innerHTML += "<dt>Lat: " + o.data.queryPlace[0].lat + "</dt>";
                innerHTML += "<dt>Lng: " + o.data.queryPlace[0].lng + "</dt>";
                innerHTML += "<dt>Avg User Rating: " + o.data.queryPlace[0].avg_user_rating + "</dt>";
                innerHTML += "</dl></td></tr></table>";
                nodeBox.innerHTML = innerHTML;
                queryBox.innerHTML = showQueryResponse(query, o);
            });
            break;
    }
}

function convertKelvin(k) {
    const num = k * 9 / 5 - 459.67;
    return Math.round((num + Number.EPSILON) * 100) / 100;
}

function showNewCityModal() {
    const message = document.getElementById("modalmessage");
    const modal = document.getElementById("newcitymodal");
    message.innerText = "";
    modal.style.display = "block";
}

function closeNewCityModal() {
    const modal = document.getElementById("newcitymodal");
    modal.style.display = "none";
}

window.onclick = function(event) {
    const modal = document.getElementById("newcitymodal");
    if (event.target == modal) {
      modal.style.display = "none";
    }
}

function addNewCity() {
    const message = document.getElementById("modalmessage");
    const countryCode = document.getElementById("countrycode");
    if (countryCode == "") {
        message.innerText = "country code is required";
        return;
    }
    const cityName = document.getElementById("cityname");
    if (cityName == "") {
        message.innerText = "city name is required";
        return;
    }
    const lat = document.getElementById("lat");
    if (lat == "") {
        message.innerText = "latitude is required";
        return;
    }
    const lng = document.getElementById("lng");
    if (lng == "") {
        message.innerText = "longitude is required";
        return;
    }

    var query = queryUploadFeed(countryCode.value, cityName.value, lat.value, lng.value);
    $.post(Dgraph, query, function (o, status) {
        if ((typeof o.errors != "undefined") && (o.errors.length > 0)) {
            message.style.color = "red";
            message.innerText = "ERROR: " + o.errors[0].message;
            return;
        }
        
        message.style.color = "green";
        message.innerText = o.data.uploadFeed.message;
        
        const selection = document.getElementById("cityselection");
        var option = document.createElement("option");
        option.text = cityName.value;
        selection.add(option);
    });
}