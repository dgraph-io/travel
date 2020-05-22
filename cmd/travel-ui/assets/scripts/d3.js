var width = 1000;	
var height = 800;

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
    var chart = (function () {
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
