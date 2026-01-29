// d3-graph.js
// Graph editor: edit nodes, add nodes, add links between nodes
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
    .attr("markerWidth", 6)
    .attr("markerHeight", 6)
    .attr("orient", "auto")
    .append("path")
    .attr("d", "M0,-5L10,0L0,5")
    .attr("fill", "#999");

  let simulation;
  let graphData;

  let link, linkLabel, node, label;

  // Link-creation state
  let linkSourceNode = null;

  // ---------- LOAD GRAPH ----------
  fetch("/api/getGraph")
    .then(res => res.json())
    .then(data => {
      graphData = data;
      initGraph(graphData);
    });

  // ---------- INIT GRAPH ----------
  function initGraph(graph) {
    svg.selectAll(".graph-layer").remove();

    applyLayerLayout(graph);

    simulation = d3.forceSimulation(graph.nodes)
      .force(
        "link",
        d3.forceLink(graph.links)
          .id(d => d.id)
          .distance(180)
          .strength(0.2)
      )
      .force("x", d3.forceX(d => d.targetX).strength(1))
      .force("y", d3.forceY(d => d.targetY).strength(1))
      .force("charge", d3.forceManyBody().strength(-30))
      .alpha(1)  // full energy
      .alphaDecay(0.05);

    // ---------- LINKS ----------
    link = svg.append("g")
      .attr("class", "graph-layer")
      .attr("stroke", "#999")
      .attr("stroke-opacity", 0.6)
      .selectAll("line")
      .data(graph.links)
      .enter()
      .append("line")
      .attr("stroke-width", d => Math.sqrt(d.weight))
      .attr("marker-end", "url(#arrow)")
      .on("click", (event, d) => {
        event.stopPropagation();
        openLinkEditor(d);
      });

    // ---------- LINK LABELS ----------
    linkLabel = svg.append("g")
      .attr("class", "graph-layer")
      .selectAll("text")
      .data(graph.links)
      .enter()
      .append("text")
      .text(d => d.weight)
      .attr("font-size", "11px")
      .attr("text-anchor", "middle")
      .attr("pointer-events", "none");

    // ---------- NODES ----------
    node = svg.append("g")
      .attr("class", "graph-layer")
      .selectAll("circle")
      .data(graph.nodes)
      .enter()
      .append("circle")
      .attr("r", 12)
      .attr("fill", d => colorByGroup(d.group))
      .attr("stroke", "#333")
      .attr("stroke-width", d => d === linkSourceNode ? 3 : 1)
      .on("click", (event, d) => {
        event.stopPropagation();
        // 1️⃣ Shift + Click → add child node
        if (event.shiftKey) {
          addNodeWithLink(d);
          return;
        }
        // 2️⃣ Cmd (mac) / Ctrl (win) + Click → link mode
        if (event.metaKey || event.ctrlKey) {
          handleLinkClick(d);
          return;
        }
        // 3️⃣ Normal click → edit node
        openNodeEditor(d);      
      })
      .call(drag(simulation));

    // ---------- NODE LABELS ----------
    label = svg.append("g")
      .attr("class", "graph-layer")
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

  // ---------- LINK CREATION ----------
  function handleLinkClick(nodeData) {
    if (!linkSourceNode) {
      linkSourceNode = nodeData;
      initGraph(graphData); // redraw highlight
      return;
    }

    if (linkSourceNode.id === nodeData.id) {
      linkSourceNode = null;
      initGraph(graphData);
      return;
    }

    const newLink = {
      source: linkSourceNode.id,
      target: nodeData.id,
      weight: 1
    };

    graphData.links.push(newLink);
    saveNewLink(newLink);

    linkSourceNode = null;
    initGraph(graphData);
  }

  function saveNewLink(link) {
    fetch("/api/link/add", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(link)
    })
    .catch(err => console.error("Add link error:", err));
  }

  // ---------- ADD NODE ----------
  function addNodeWithLink(parentNode) {
    const newNode = {
      id: "Node_" + Date.now(),
      group: parentNode.group,
      layer: parentNode.layer + 1
    };

    const newLink = {
      source: parentNode.id,
      target: newNode.id,
      weight: 1
    };

    graphData.nodes.push(newNode);
    graphData.links.push(newLink);

    fetch("/api/node/add", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ node: newNode, link: newLink })
    });

    initGraph(graphData);
  }

  // ---------- EDIT NODE ----------
function openNodeEditor(nodeData) {
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
    <input id="edit-node-id" value="${nodeData.id}" /><br><br>

    Group:<br>
    <input id="edit-node-group" type="number" value="${nodeData.group}" /><br><br>

    Layer:<br>
    <input id="edit-node-layer" type="number" value="${nodeData.layer}" /><br><br>

    Value:<br>
    <input id="edit-node-value" type="number" value="${nodeData.value}" /><br><br>

    ActivFn:<br>
    <input id="edit-node-activFn" value="${nodeData.activFn}" /><br><br>

    <button id="save-node">Save</button>
    <button id="cancel-node">Cancel</button>
  `);

  // ---- SAVE ----
  d3.select("#save-node").on("click", () => {
    const oldId = nodeData.id;

    nodeData.id = document.getElementById("edit-node-id").value.trim();
    nodeData.group = +document.getElementById("edit-node-group").value;
    nodeData.layer = +document.getElementById("edit-node-layer").value;
    nodeData.value = +document.getElementById("edit-node-value").value;
    nodeData.activFn = document.getElementById("edit-node-activFn").value;

    // Update links if ID changed
    if (oldId !== nodeData.id) {
      graphData.links.forEach(l => {
        if (typeof l.source === "object") {
          if (l.source.id === oldId) l.source.id = nodeData.id;
        } else if (l.source === oldId) {
          l.source = nodeData.id;
        }

        if (typeof l.target === "object") {
          if (l.target.id === oldId) l.target.id = nodeData.id;
        } else if (l.target === oldId) {
          l.target = nodeData.id;
        }
      });
    }

    // Recompute layout
    applyLayerLayout(graphData);

    // Re-render graph
    initGraph(graphData);

    // Persist ONLY this node
    saveNode(nodeData);

    editor.remove();
  });

  // ---- CANCEL ----
  d3.select("#cancel-node").on("click", () => editor.remove());
}

function saveNode(nodeData) {
  fetch("/api/node", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      id: nodeData.id,
      group: nodeData.group,
      layer: nodeData.layer,
      value: nodeData.value,
      activFn: nodeData.activFn
    })
  })
  .catch(err => console.error("Save node error:", err));
}

// ---------- LINK EDITOR ----------
  function openLinkEditor(linkData) {
    d3.select("#link-editor").remove();

    const editor = d3.select("body")
      .append("div")
      .attr("id", "link-editor")
      .style("position", "fixed")
      .style("top", "20px")
      .style("left", "20px")
      .style("background", "#fff")
      .style("border", "1px solid #ccc")
      .style("padding", "12px")
      .style("font-size", "13px")
      .style("box-shadow", "0 4px 10px rgba(0,0,0,0.15)")
      .style("z-index", 2000);

    const src = typeof linkData.source === "object" ? linkData.source.id : linkData.source;
    const tgt = typeof linkData.target === "object" ? linkData.target.id : linkData.target;

    editor.html(`
      <strong>Edit Link</strong><br><br>
      Source: <b>${src}</b><br>
      Target: <b>${tgt}</b><br><br>
      Weight:<br>
      <input id="edit-weight" type="number" value="${linkData.weight}" /><br><br>
      <button id="save-link">Save</button>
      <button id="cancel-link">Cancel</button>
    `);

    d3.select("#save-link").on("click", () => {
      linkData.weight = +document.getElementById("edit-weight").value;

      link
        .attr("stroke-width", d => Math.sqrt(d.weight));
      linkLabel
        .text(d => d.weight);

      saveLink(linkData);

      editor.remove();
    });

    d3.select("#cancel-link").on("click", () => editor.remove());
  }
    // ---------- SAVE LINK ----------
  function saveLink(linkData) {
    const cleanLink = {
      source: typeof linkData.source === "object" ? linkData.source.id : linkData.source,
      target: typeof linkData.target === "object" ? linkData.target.id : linkData.target,
      weight: linkData.weight
    };

    fetch("/api/link", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(cleanLink)
    })
    .catch(err => console.error("Save link error:", err));
  }

  // ---------- LAYOUT ----------
  // function applyLayerLayout(graph) {
  //   const nodesByLayer = d3.group(graph.nodes, d => d.layer);
  //   const maxLayer = d3.max(graph.nodes, d => d.layer);

  //   const xScale = d3.scalePoint()
  //     .domain(d3.range(1, maxLayer + 1))
  //     .range([120, width - 120]);

  //   nodesByLayer.forEach(nodes => {
  //     const top = 80;
  //     const bottom = height - 80;
  //     const step = (bottom - top) / (nodes.length + 1);

  //     nodes.forEach((n, i) => {
  //       n.targetX = xScale(n.layer);
  //       n.targetY = top + (i + 1) * step;
  //     });
  //   });
  // }

  function applyLayerLayout(graph) {
  const nodesByLayer = d3.group(graph.nodes, d => d.layer);
  const maxLayer = d3.max(graph.nodes, d => d.layer);

  const xScale = d3.scalePoint()
    .domain(d3.range(1, maxLayer + 1))
    .range([140, width - 140]);

  nodesByLayer.forEach(nodes => {
    const top = 80;
    const bottom = height - 80;
    const step = (bottom - top) / (nodes.length + 1);

    nodes.forEach((n, i) => {
      n.targetX = xScale(n.layer);
      n.targetY = top + (i + 1) * step;

      // 🔒 HARD RESET X/Y to target positions
      n.x = n.targetX;
      n.y = n.targetY;
      n.fx = null;
      n.fy = null;
    });
  });
}


  // ---------- DRAG ----------
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

  function colorByGroup(group) {
    const colors = {
      1: "#1f77b4",
      2: "#ff7f0e",
      3: "#2ca02c",
      4: "#d62728"
    };
    return colors[group] || "#999";
  }

  // ---------- RESIZE ----------
  window.addEventListener("resize", () => {
    width = window.innerWidth;
    height = window.innerHeight;
    svg.attr("viewBox", `0 0 ${width} ${height}`);
    initGraph(graphData);
  });

  function relayoutGraph() {
    if (!graphData || !simulation) return;

    // Recompute target positions strictly by layer
    applyLayerLayout(graphData);

    // Release any manual drag locks
    graphData.nodes.forEach(n => {
      n.fx = null;
      n.fy = null;
    });

    // Force nodes back into columns + rows
    simulation
      .force("x", d3.forceX(d => d.targetX).strength(1))
      .force("y", d3.forceY(d => d.targetY).strength(1))
      .alpha(1)          // full energy
      .restart();

    // Let it settle, then freeze again
    setTimeout(() => simulation.alpha(0), 800);
  }

  document.getElementById("relayout-btn")
  .addEventListener("click", () => {
    relayoutGraph();
  });

});
