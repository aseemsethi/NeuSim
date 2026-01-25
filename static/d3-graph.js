// Wait for the DOM to be fully loaded
document.addEventListener('DOMContentLoaded', (event) => {
    const svg = d3.select("svg"),
        margin = {top: 20, right: 20, bottom: 30, left: 40},
        width = +svg.attr("width") - margin.left - margin.right,
        height = +svg.attr("height") - margin.top - margin.bottom;

    const g = svg.append("g")
        .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    // Define scales
    const x = d3.scaleBand()
        .rangeRound([0, width])
        .padding(0.1);

    const y = d3.scaleLinear()
        .rangeRound([height, 0]);

    // Fetch the JSON data from the Go backend
    d3.json("/api/data").then(data => {
        // Ensure data is correctly accessed (based on the Go struct "Points")
        const points = data.points;

        // Set the domain of the scales
        x.domain(points.map(d => d.label));
        y.domain([0, d3.max(points, d => d.value)]);

        // Add the X axis
        g.append("g")
            .attr("class", "axis axis--x")
            .attr("transform", "translate(0," + height + ")")
            .call(d3.axisBottom(x));

        // Add the Y axis
        g.append("g")
            .attr("class", "axis axis--y")
            .call(d3.axisLeft(y))
            .append("text")
            .attr("transform", "rotate(-90)")
            .attr("y", 6)
            .attr("dy", "0.71em")
            .attr("text-anchor", "end")
            .text("Value");

        // Add the bars
        g.selectAll(".bar")
            .data(points)
            .enter().append("rect")
            .attr("class", "bar")
            .attr("x", d => x(d.label))
            .attr("y", d => y(d.value))
            .attr("width", x.bandwidth())
            .attr("height", d => height - y(d.value));
            
    }).catch(error => {
        console.error("Error fetching data:", error);
    });
});
