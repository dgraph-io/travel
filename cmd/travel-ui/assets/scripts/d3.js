color = (function(){
  const scale = d3.scaleOrdinal(d3.schemeCategory10);
  return d => scale(d.group);
})();

const drag = simulation => {
    const dragstarted = d => {
	    if (!d3.event.active) simulation.alphaTarget(0.3).restart();
	    d.fx = d.x;
	    d.fy = d.y;
    }

    const dragged = d => {
	    d.fx = d3.event.x;
	    d.fy = d3.event.y;
    }

    const dragended = d => {
	    if (!d3.event.active) simulation.alphaTarget(0);
	    d.fx = null;
	    d.fy = null;
    }

    return d3.drag()
	  .on("start", dragstarted)
	  .on("drag", dragged)
	  .on("end", dragended);
}

function makechart(data) {
    const f = function () {
        const width = document.querySelector("div.graphbox").clientWidth;
        const height = document.querySelector("div.graphbox").clientHeight;
        const links = data.links.map(d => Object.create(d));
        const nodes = data.nodes.map(d => Object.create(d));
        const manyBody = d3.forceManyBody()
            .strength(-200);
        const simulation = d3.forceSimulation(nodes)
            .force("link", d3.forceLink(links).id(d => d.id))
            .force("charge", manyBody)
            .force("center", d3.forceCenter((width / 2), (height / 2)));
        const svg = d3.create("svg")
            .attr("viewBox", [10, 10, width, height])
            .attr('preserveAspectRatio', 'xMinYMid');
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
            .attr("rorg", d => d.radius)
            .attr("fill", d => d.color)
            .on("click", showNodeData)
            .on("mouseover", circleMouseOver)
            .on("mouseout", circleMouseOut)
            .call(drag(simulation));
        node.append("title")
            .text(d => d.id.split(":")[0]);
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
    }

    const chart = f();
    document.querySelector("div.graphbox").appendChild(chart);
}

function drawchart(city) {
    document.querySelector("div.graphbox").innerHTML = "";
    
    const err = function(error) {
        document.querySelector("div.graphbox").innerHTML = "no data for city: " + city;
    }
    d3.json("/data/" + city).then(makechart).catch(err);
}
