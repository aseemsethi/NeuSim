// d3-graph.js
// Full graph editor with layered layout and link value labels
// Requires d3.js v6+

document.addEventListener("DOMContentLoaded", () => {

  let width = window.innerWidth;
  let height = window.innerHeight;

  const svg = d3.select("#graph")
    .append("svg")
    .attr("viewBox", `0 0 ${width} ${height}`)
    .style("width", "100%")
    .style("height", "100%");

  // ---------- Arrow marker ----------
  svg.append("defs")
    .append("marker")
    .attr("id", "arrow")
    .attr("viewBox", "0 -5 10 10")
    .attr("refX", 22)
    .attr("refY", 0)
    .attr("markerWidth", 8)
    .attr("markerHeight", 8)
    .attr("orient", "auto")
    .append("path")
    .attr("d", "M0,-5L10,0L0,5")
    .attr("fill", "#999");

  let simulation, graphData;
  let link, linkLabel, node, label;

  fetch("/api/test")
    .then(res => res.json())
    .then(data => {
      graphData = data;
      initGraph(data);
    })
    .catch(err => console.error(err));

  function initGraph(graph) {
    applyLayerLayout(graph);

    simulation = d3.forceSimulation(graph.nodes)
      .force(
        "link",
        d3.forceLink(graph.links)
          .id(d => d.id)
          .distance(120)
      )
      .force("x", d3.forceX(d => d.targetX).strength(1))
      .force("y", d3.forceY(d => d.targetY).strength(1))
      .force("charge", d3.forceManyBody().strength(-150))
      .alphaDecay(0.05);

    // ---------- LINKS ----------
    link = svg.append("g")
      .attr("stroke", "#999")
      .attr("stroke-opacity", 0.6)
      .selectAll("line")
      .data(graph.links)
      .enter()
      .append("line")
      .attr("stroke-width", d => Math.sqrt(d.value))
      .attr("marker-end", "url(#arrow)");

    // ---------- LINK LABELS (VALUE) ----------
    linkLabel = svg.append("g")
      .selectAll("text")
      .data(graph.links)
      .enter()
      .append("text")
      .text(d => d.weight)
      .attr("font-size", "11px")
      .attr("fill", "#333")
      .attr("text-anchor", "middle")
      .attr("pointer-events", "none"); // prevents blocking drag

    // ---------- NODES ----------
    node = svg.append("g")
      .selectAll("circle")
      .data(graph.nodes)
      .enter()
      .append("circle")
      .attr("r", 12)
      .attr("fill", d => colorByGroup(d.group))
      .on("click", (event, d) => {
        event.stopPropagation();
        openEditor(d);
      })
      .call(drag(simulation));

    // ---------- NODE LABELS ----------
    label = svg.append("g")
      .selectAll("text")
      .data(graph.nodes)
      .enter()
      .append("text")
      .text(d => d.id)
      .attr("dx", 15)
      .attr("dy", ".35em")
      .attr("font-size", "12px");

    // ---------- TICK ----------
    simulation.on("tick", () => {

      link
        .attr("x1", d => d.source.x)
        .attr("y1", d => d.source.y)
        .attr("x2", d => d.target.x)
        .attr("y2", d => d.target.y);

      // Position link value at midpoint
      linkLabel
        .attr("x", d => (d.source.x + d.target.x) / 2)
        .attr("y", d => (d.source.y + d.target.y) / 2);

      node
        .attr("cx", d => d.x)
        .attr("cy", d => d.y);

      label
        .attr("x", d => d.x)
        .attr("y", d => d.y);
    });

    setTimeout(() => simulation.alpha(0), 1200);
  }

  // ---------- Layered Layout ----------
  function applyLayerLayout(graph) {
    const nodesByLayer = d3.group(graph.nodes, d => d.layer);
    const maxLayer = d3.max(graph.nodes, d => d.layer);

    const xScale = d3.scalePoint()
      .domain(d3.range(1, maxLayer + 1))
      .range([120, width - 120]);

    nodesByLayer.forEach(nodes => {
      const top = 80;
      const bottom = height - 80;
      const step = (bottom - top) / (nodes.length + 1);

      nodes.forEach((n, i) => {
        n.targetX = xScale(n.layer);
        n.targetY = top + (i + 1) * step;
      });
    });
  }

  // ---------- Drag ----------
  function drag(sim) {
    return d3.drag()
      .on("start", (e, d) => {
        sim.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
      })
      .on("drag", (e, d) => {
        d.fx = e.x;
        d.fy = e.y;
      })
      .on("end", (e, d) => {
        d.fx = e.x;
        d.fy = e.y;
        sim.alphaTarget(0);
      });
  }

  // ---------- Node Editor ----------
  function openEditor(nodeData) {
    d3.select("#node-editor").remove();

    const editor = d3.select("body")
      .append("div")
      .attr("id", "node-editor")
      .style("position", "fixed")
      .style("top", "20px")
      .style("right", "20px")
      .style("background", "#fff")
      .style("border", "1px solid #ccc")
      .style("padding", "12px")
      .style("font-size", "13px")
      .style("box-shadow", "0 4px 10px rgba(0,0,0,0.15)")
      .style("z-index", 2000);

    editor.html(`
      <strong>Edit Node</strong><br><br>
      ID:<br>
      <input id="edit-id" value="${nodeData.id}" /><br><br>
      Group:<br>
      <input id="edit-group" type="number" value="${nodeData.group}" /><br><br>
      Layer:<br>
      <input id="edit-layer" type="number" value="${nodeData.layer}" /><br><br>
      <button id="save-node">Save</button>
      <button id="cancel-node">Cancel</button>
    `);

    d3.select("#save-node").on("click", () => {
      nodeData.id = document.getElementById("edit-id").value.trim();
      nodeData.group = +document.getElementById("edit-group").value;
      nodeData.layer = +document.getElementById("edit-layer").value;

      applyLayerLayout(graphData);

      label.text(d => d.id);
      node.attr("fill", d => colorByGroup(d.group));

      simulation
        .force("x", d3.forceX(d => d.targetX).strength(1))
        .force("y", d3.forceY(d => d.targetY).strength(1))
        .alpha(0.6)
        .restart();

      editor.remove();
    });

    d3.select("#cancel-node").on("click", () => editor.remove());
  }

  function colorByGroup(group) {
    const colors = {
      1: "#1f77b4",
      2: "#ff7f0e",
      3: "#2ca02c",
      4: "#d62728"
    };
    return colors[group] || "#999";
  }

  // ---------- Resize ----------
  window.addEventListener("resize", () => {
    width = window.innerWidth;
    height = window.innerHeight;
    svg.attr("viewBox", `0 0 ${width} ${height}`);

    applyLayerLayout(graphData);
    simulation
      .force("x", d3.forceX(d => d.targetX).strength(1))
      .force("y", d3.forceY(d => d.targetY).strength(1))
      .alpha(0.3)
      .restart();
  });

});
