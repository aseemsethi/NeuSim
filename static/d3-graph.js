// d3-graph.js
// Layered columns with evenly spaced vertical nodes
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

  // Arrow marker
  svg.append("defs")
    .append("marker")
    .attr("id", "arrow")
    .attr("viewBox", "0 -5 10 10")
    .attr("refX", 22)
    .attr("refY", 0)
    .attr("markerWidth", 6)
    .attr("markerHeight", 6)
    .attr("orient", "auto")
    .append("path")
    .attr("d", "M0,-5L10,0L0,5")
    .attr("fill", "#999");

  let simulation;

  fetch("/api/test")
    .then(res => res.json())
    .then(data => initGraph(data))
    .catch(err => console.error(err));

  function initGraph(graph) {

    // ----- GROUP NODES BY LAYER -----
    const nodesByLayer = d3.group(graph.nodes, d => d.layer);
    const maxLayer = d3.max(graph.nodes, d => d.layer);

    // ----- X POSITION (COLUMNS) -----
    const layerScale = d3.scalePoint()
      .domain(d3.range(1, maxLayer + 1))
      .range([120, width - 120]);

    // ----- Y POSITION (STACKED PER LAYER) -----
    nodesByLayer.forEach(nodes => {
      const paddingTop = 80;
      const paddingBottom = height - 80;
      const step = (paddingBottom - paddingTop) / (nodes.length + 1);

      nodes.forEach((node, i) => {
        node.targetY = paddingTop + (i + 1) * step;
      });
    });

    simulation = d3.forceSimulation(graph.nodes)
      .force(
        "link",
        d3.forceLink(graph.links)
          .id(d => d.id)
          .distance(120)
      )
      // ⬅️ column placement
      .force(
        "x",
        d3.forceX(d => layerScale(d.layer)).strength(1)
      )
      // ⬆️ vertical placement (FIX)
      .force(
        "y",
        d3.forceY(d => d.targetY).strength(1)
      )
      .force("charge", d3.forceManyBody().strength(-150))
      .alphaDecay(0.05);

    // LINKS
    const link = svg.append("g")
      .attr("stroke", "#999")
      .attr("stroke-opacity", 0.6)
      .selectAll("line")
      .data(graph.links)
      .enter()
      .append("line")
      .attr("stroke-width", d => Math.sqrt(d.weight))
      .attr("marker-end", "url(#arrow)");

    // NODES
    const node = svg.append("g")
      .selectAll("circle")
      .data(graph.nodes)
      .enter()
      .append("circle")
      .attr("r", 12)
      .attr("fill", d => colorByGroup(d.group))
      .call(drag(simulation));

    // LABELS
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

    // Let it settle, then sleep
    setTimeout(() => simulation.alpha(0), 1200);
  }

  // DRAG (can reposition again)
  function drag(simulation) {
    return d3.drag()
      .on("start", (event, d) => {
        simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
      })
      .on("drag", (event, d) => {
        d.fx = event.x;
        d.fy = event.y;
      })
      .on("end", (event, d) => {
        d.fx = event.x;
        d.fy = event.y;
        simulation.alphaTarget(0);
      });
  }

  function colorByGroup(group) {
    const colors = {
      1: "#1f77b4",
      2: "#ff7f0e",
      3: "#2ca02c"
    };
    return colors[group] || "#999";
  }

  // Resize: recompute vertical spacing
  window.addEventListener("resize", () => {
    width = window.innerWidth;
    height = window.innerHeight;
    svg.attr("viewBox", `0 0 ${width} ${height}`);

    if (!simulation) return;

    const nodesByLayer = d3.group(simulation.nodes(), d => d.layer);

    nodesByLayer.forEach(nodes => {
      const paddingTop = 80;
      const paddingBottom = height - 80;
      const step = (paddingBottom - paddingTop) / (nodes.length + 1);

      nodes.forEach((node, i) => {
        node.targetY = paddingTop + (i + 1) * step;
      });
    });

    simulation
      .force("y", d3.forceY(d => d.targetY).strength(1))
      .alpha(0.3)
      .restart();
  });

});
