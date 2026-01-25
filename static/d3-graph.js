// d3-graph.js
// Requires d3.js v6+

document.addEventListener("DOMContentLoaded", () => {

  let width = window.innerWidth;
  let height = window.innerHeight;

  const svg = d3
    .select("#graph")
    .append("svg")
    .attr("viewBox", `0 0 ${width} ${height}`)
    .style("width", "100%")
    .style("height", "100%");

  let simulation;

  fetch("/api/test")
    .then(res => res.json())
    .then(data => initGraph(data))
    .catch(err => console.error(err));

  function initGraph(graph) {

    simulation = d3
      .forceSimulation(graph.nodes)
      .force(
        "link",
        d3.forceLink(graph.links)
          .id(d => d.id)
          .distance(120)
      )
      .force("charge", d3.forceManyBody().strength(-300))
      .force("center", d3.forceCenter(width / 2, height / 2))
      .alphaDecay(0.05); // cool down reasonably fast

    // Links
    const link = svg.append("g")
      .attr("stroke", "#999")
      .attr("stroke-opacity", 0.6)
      .selectAll("line")
      .data(graph.links)
      .enter()
      .append("line")
      .attr("stroke-width", d => Math.sqrt(d.value));

    // Nodes
    const node = svg.append("g")
      .selectAll("circle")
      .data(graph.nodes)
      .enter()
      .append("circle")
      .attr("r", 12)
      .attr("fill", d => colorByGroup(d.group))
      .call(drag(simulation));

    // Labels
    const label = svg.append("g")
      .selectAll("text")
      .data(graph.nodes)
      .enter()
      .append("text")
      .text(d => d.id)
      .attr("dx", 15)
      .attr("dy", ".35em")
      .attr("font-size", "12px");

    simulation.on("tick", () => {
      link
        .attr("x1", d => d.source.x)
        .attr("y1", d => d.source.y)
        .attr("x2", d => d.target.x)
        .attr("y2", d => d.target.y);

      node
        .attr("cx", d => d.x)
        .attr("cy", d => d.y);

      label
        .attr("x", d => d.x)
        .attr("y", d => d.y);
    });

    // Let simulation settle, then pause
    setTimeout(() => {
      simulation.alpha(0);
    }, 1500);
  }

  // 🔁 Drag behavior: re-enable movement anytime
  function drag(simulation) {

    function dragstarted(event, d) {
      simulation.alphaTarget(0.3).restart();
      d.fx = d.x;
      d.fy = d.y;
    }

    function dragged(event, d) {
      d.fx = event.x;
      d.fy = event.y;
    }

    function dragended(event, d) {
      d.fx = event.x;
      d.fy = event.y;
      simulation.alphaTarget(0); // cool back down
    }

    return d3.drag()
      .on("start", dragstarted)
      .on("drag", dragged)
      .on("end", dragended);
  }

  function colorByGroup(group) {
    const colors = {
      1: "#1f77b4",
      2: "#ff7f0e",
      3: "#2ca02c"
    };
    return colors[group] || "#999";
  }

  // Resize SVG only — no physics reset
  window.addEventListener("resize", () => {
    width = window.innerWidth;
    height = window.innerHeight;
    svg.attr("viewBox", `0 0 ${width} ${height}`);
  });

});
