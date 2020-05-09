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
		<link rel="stylesheet" href="/static/css/main.css" rel='stylesheet' type='text/css'/>
	</head>
	<style>
	.graphbox {
		width: 400px;
		heigth: 500px;
		border: 1px solid #333;
		box-shadow: 8px 8px 5px #444;
		padding: 8px 12px;
		background-image: linear-gradient(180deg, #fff, #ddd 40%, #ccc);
	}
	.databox {
		width: 400px;
		heigth: 500px;
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
	</style>
	<body>
		<table>
			<tr>
				<td><div class="graphbox"></div></td>
				<td>
					<div class="databox">
						<table>
							<tr><td><div>City: Sydney</div></td></tr>
							<tr><td id="data"></td></tr>
						</table>
					</div>
				</td>
			</tr>
		</table>
		<script>
			var width = 400;	
			var height = 500;
			
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

			function showInfo(d, i) {
				var cell = document.getElementById("data");
				cell.innerText = d.id;
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
	</script>
  </body>
</html>`
