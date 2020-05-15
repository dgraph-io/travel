package handlers

import (
	"context"
	"io"
	"net/http"
)

func index(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	io.WriteString(w, indexHTML)
	return nil
}

var indexHTML = `<!DOCTYPE html>
<html lang="en">
	<head>
		<title>Travel Graph</title>
		<meta content="City Graph" name="description">
		<meta charset="utf-8">
		<script src="https://d3js.org/d3.v5.min.js"></script>
		<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.5.1/jquery.min.js"></script>
	</head>
	<style>
	.graphbox {
		width: 500px;
		heigth: 600px;
		border: 1px solid #333;
		box-shadow: 8px 8px 5px #444;
		padding: 8px 12px;
		background-image: linear-gradient(180deg, #fff, #ddd 40%, #ccc);
	}
	.databox {
		width: 500px;
		heigth: 600px;
		border: 2px solid #003B62;
  		font-family: verdana;
  		background-color: #B5CFE0;
  		padding-left: 5px;
	}
	td {
		width: 500px;
		padding: 10px;
  		text-align: left;
		vertical-align: top;
	}
	.bluedot {
		height: 25px;
		width: 25px;
		background-color: blue;
		border-radius: 50%;
		border: 1px solid #000;
	}
	.reddot {
		height: 25px;
		width: 25px;
		background-color: red;
		border-radius: 50%;
		border: 1px solid #000;
	}
	.orangedot {
		height: 25px;
		width: 25px;
		background-color: orange;
		border-radius: 50%;
		border: 1px solid #000;
	}
	.purpledot {
		height: 25px;
		width: 25px;
		background-color: purple;
		border-radius: 50%;
		border: 1px solid #000;
	}
	</style>
	<body>
		<table>
			<tr>
				<td><div class="graphbox"></div></td>
				<td>
					<div class="databox">
						<table>
							<tr>
								<td>
									<div class="bluedot"></div>
									<div>City</div>
								</td>
								<td>
									<div class="reddot"></div>
									<div>Advisory</div>
								</td>
								<td>
									<div class="orangedot"></div>
									<div>Weather</div>
								</td>
								<td>
									<div class="purpledot"></div>
									<div>Place</div>
								</td>
							</tr>
						</table>
						<table>
							<tr><td id="data"></td></tr>
						</table>
					</div>
				</td>
			</tr>
		</table>
		<script>
			var width = 500;	
			var height = 600;
			
			color = (function(){
			  const scale = d3.scaleOrdinal(d3.schemeCategory10);
			  return d => scale(d.group);
			})();
			
			var drag = simulation => {
	  
			  function dragstarted(d) {
				if (!d3.event.active) simulation.alphaTarget(0.3).restart();
				d.fx = d.x;
				d.fy = d.y;
			  }
	  
			  function dragged(d) {
				d.fx = d3.event.x;
				d.fy = d3.event.y;
			  }
	  
			  function dragended(d) {
				if (!d3.event.active) simulation.alphaTarget(0);
				d.fx = null;
				d.fy = null;
			  }
	  
			  return d3.drag()
				  .on("start", dragstarted)
				  .on("drag", dragged)
				  .on("end", dragended);
			}
			
			d3.json("/data").then(function(data) {
        		var chart = (function(){
          		const links = data.links.map(d => Object.create(d));
				  const nodes = data.nodes.map(d => Object.create(d));
				  
				const manyBody = d3.forceManyBody()
					.strength(-200);

          		const simulation = d3.forceSimulation(nodes)
              		.force("link", d3.forceLink(links).id(d => d.id))
					.force("charge", manyBody)
              		.force("center", d3.forceCenter((width / 2)+(width / 8), (height / 2)+(height / 6)));

          		const svg = d3.create("svg")
              		.attr("viewBox", [10, 10, width, height]);

          		const link = svg.append("g")
              		.attr("stroke", "#999")
              		.attr("stroke-opacity", 0.6)
            		.selectAll("line")
            		.data(links)
            		.join("line")
              		.attr("stroke-width", d => Math.sqrt(d.width));

          		const node = svg.append("g")
              		.attr("stroke-width", 1.5)
					.attr("stroke","black")
            		.selectAll("circle")
            		.data(nodes)
            		.join("circle")
              		.attr("r", d => d.radius)
              		.attr("fill", d => d.color)
					.on("click", showInfo)
              		.call(drag(simulation));

          		node.append("title")
              		.text(d => d.id);

          		simulation.on("tick", () => {
            		link
                		.attr("x1", d => d.source.x)
                		.attr("y1", d => d.source.y)
                		.attr("x2", d => d.target.x)
                		.attr("y2", d => d.target.y);

            		node
                		.attr("cx", d => d.x)
                		.attr("cy", d => d.y);
          		});
          
          	return svg.node();
        })();
		document.querySelector("div.graphbox").appendChild(chart);
	})

	$.ajaxSetup({
		contentType: "application/json; charset=utf-8"
	});

	function convertKelvin(k) {
		var num = k * 9/5 - 459.67
		return Math.round((num + Number.EPSILON) * 100) / 100
	}

	function showInfo(d, i) {
		var cell = document.getElementById("data");
		switch (d.type) {
			case "city":
				$.post("http://localhost:8080/graphql",
				'{"query":"query { getCity(id: \\"0x02\\") { id name lat lng } }","variables":null}',
				function(o, status){
					if (typeof o.data === "undefined") {
						cell.innerText = "ERROR: " + o.errors[0].message;
						return;
					}
					var innerHTML = "<table width=\"70%\">";
					innerHTML += "<tr><td><div class=\"bluedot\"></div></td><td>City</td></tr>";
					innerHTML += "<tr><td>ID:</td><td>" + o.data.getCity.id + "</td></tr>";
					innerHTML += "<tr><td>Name:</td><td>" + o.data.getCity.name + "</td></tr>";
					innerHTML += "<tr><td>Lat:</td><td>" + o.data.getCity.lat + "</td></tr>";
					innerHTML += "<tr><td>Lng:</td><td>" + o.data.getCity.lng + "</td></tr>";
					innerHTML += "</table>";
					cell.innerHTML = innerHTML;
				});
				break;
			case "advisory":
				$.post("http://localhost:8080/graphql",
				'{"query":"query { getCity(id: \\"0x02\\") { advisory { id continent country country_code last_updated message score source }} }","variables":null}',
				function(o, status){
					if (typeof o.data === "undefined") {
						cell.innerText = "ERROR: " + o.errors[0].message;
						return;
					}
					var innerHTML = "<table width=\"70%\">";
					innerHTML += "<tr><td><div class=\"reddot\"></div></td><td>Advisory</td></tr>";
					innerHTML += "<tr><td>ID:</td><td>" + o.data.getCity.advisory.id + "</td></tr>";
					innerHTML += "<tr><td>Country:</td><td>" + o.data.getCity.advisory.country + "</td></tr>";
					innerHTML += "<tr><td>Country Code:</td><td>" + o.data.getCity.advisory.country_code + "</td></tr>";
					innerHTML += "<tr><td>Continent:</td><td>" + o.data.getCity.advisory.continent + "</td></tr>";
					innerHTML += "<tr><td>Score:</td><td>" + o.data.getCity.advisory.score + "</td></tr>";
					innerHTML += "<tr><td>Message:</td><td>" + o.data.getCity.advisory.message + "</td></tr>";
					innerHTML += "</table>";
					cell.innerHTML = innerHTML;
				});
				break;
			case "weather":
				$.post("http://localhost:8080/graphql",
				'{"query":"query { getCity(id: \\"0x02\\") { weather { id city_name description feels_like humidity pressure sunrise sunset temp temp_min temp_max visibility wind_direction wind_speed }} }","variables":null}',
				function(o, status){
					if (typeof o.data === "undefined") {
						cell.innerText = "ERROR: " + o.errors[0].message;
						return;
					}
					var innerHTML = "<table width=\"70%\">";
					innerHTML += "<tr><td><div class=\"orangedot\"></div></td><td>Weather</td></tr>";
					innerHTML += "<tr><td>ID:</td><td>" + o.data.getCity.weather.id + "</td></tr>";
					innerHTML += "<tr><td>City Name:</td><td>" + o.data.getCity.weather.city_name + "</td></tr>";
					innerHTML += "<tr><td>Visibility:</td><td>" + o.data.getCity.weather.visibility + "</td></tr>";
					innerHTML += "<tr><td>Description:</td><td>" + o.data.getCity.weather.description + "</td></tr>";
					innerHTML += "<tr><td>Temp:</td><td>" + convertKelvin(o.data.getCity.weather.temp) + "F</td></tr>";
					innerHTML += "<tr><td>Feels Like:</td><td>" + convertKelvin(o.data.getCity.weather.feels_like) + "F</td></tr>";
					innerHTML += "<tr><td>Min Temp:</td><td>" + convertKelvin(o.data.getCity.weather.temp_min) + "F</td></tr>";
					innerHTML += "<tr><td>Max Temp:</td><td>" + convertKelvin(o.data.getCity.weather.temp_max) + "F</td></tr>";
					innerHTML += "<tr><td>Pressure:</td><td>" + o.data.getCity.weather.pressure + "</td></tr>";
					innerHTML += "<tr><td>Humidity:</td><td>" + o.data.getCity.weather.humidity + "</td></tr>";
					innerHTML += "<tr><td>Wind Speed:</td><td>" + o.data.getCity.weather.wind_speed + "</td></tr>";
					innerHTML += "<tr><td>Wind Direction:</td><td>" + o.data.getCity.weather.wind_direction + "</td></tr>";
					innerHTML += "</table>";
					cell.innerHTML = innerHTML;
				});
				break;
			case "place":
				$.post("http://localhost:8080/graphql",
				'{"query":"query { queryPlace(filter: { name: { eq: \\"' + d.id + '\\" } }) { id address avg_user_rating city_name gmaps_url lat lng location_type name no_user_rating place_id photo_id } }","variables":null}',
				function(o, status){
					if (typeof o.data === "undefined") {
						cell.innerText = "ERROR: " + o.errors[0].message;
						return;
					}
					var innerHTML = "<table width=\"70%\">";
					innerHTML += "<tr><td><div class=\"purpledot\"></div></td><td>Place</td></tr>";
					innerHTML += "<tr><td>ID:</td><td>" + o.data.queryPlace[0].id + "</td></tr>";
					innerHTML += "<tr><td>Name:</td><td>" + o.data.queryPlace[0].name + "</td></tr>";
					innerHTML += "<tr><td>Address:</td><td>" + o.data.queryPlace[0].address + "</td></tr>";
					innerHTML += "<tr><td>Avg User Rating:</td><td>" + o.data.queryPlace[0].avg_user_rating + "</td></tr>";
					innerHTML += "</table>";
					cell.innerHTML = innerHTML;
				});
				break;
		}
	}
	</script>
  </body>
</html>`
