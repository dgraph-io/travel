$.ajaxSetup({
    contentType: "application/json; charset=utf-8"
});

function loadData() {
    document.getElementById("node").innerHTML = "";
    document.getElementById("query").innerHTML = "";
    showTab("node");

    drawchart(document.getElementById("cityselection").value);
    const d = { type: "city" };
    showInfo(d);
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

function showTab(which) {
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
    const nodeBox = document.getElementById("node");
    const queryBox = document.getElementById("query");
    const name = document.getElementById("cityselection").value;

    switch (d.type) {
        case "city":
            var query = queryCity(name);
            $.post(Dgraph, query, function (o, status) {
                if (typeof o.data === "undefined") {
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
                if (typeof o.data === "undefined") {
                    nodeBox.innerText = "ERROR: " + o.errors[0].message;
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
                nodeBox.innerHTML = innerHTML;
                queryBox.innerHTML = showQueryResponse(query, o);
            });
            break;

        case "weather":
            var query = queryWeather(name);
            $.post(Dgraph, query, function(o, status) {
                if (typeof o.data === "undefined") {
                    nodeBox.innerText = "ERROR: " + o.errors[0].message;
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
                nodeBox.innerHTML = innerHTML;
                queryBox.innerHTML = showQueryResponse(query, o);
            });
            break;

        case "place":
            var query = queryPlaceByCategory(name, d.id);
            $.post(Dgraph, query, function(o, status) {
                if (typeof o.data === "undefined") {
                    nodeBox.innerText = "ERROR: " + o.errors[0].message;
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
                nodeBox.innerHTML = innerHTML;
                queryBox.innerHTML = showQueryResponse(query, o);
            });
            break;

        default:
            var query = queryPlaceByName(d.id);
            $.post(Dgraph, query, function(o, status) {
                if (typeof o.data === "undefined") {
                    nodeBox.innerText = "ERROR: " + o.errors[0].message;
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