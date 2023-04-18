function Punchcard(data, {
    x = ([x]) => x, // given d in data, returns the (temporal) x-value
    y = ([, y]) => y, // given d in data, returns the (quantitative) y-value
    title, // given d in data, returns the title text
    // width = 928, // width of the chart, in pixels
    cellSize = 17, // width and height of an individual day, in pixels
    cellPadding=1,
    weekday = "monday", // either: weekday, sunday, or monday
    formatDay = i => "SMTWTFS"[i], // given a day number in [0, 6], the day-of-week label
    formatHour = "%I %p", // format specifier string for months (above the chart)
    // yFormat, // format specifier string for values (in the title)
    color = undefined,
    emptyColor = undefined,
  } = {}) {
    // Compute values.

    const countDay = weekday === "sunday" ? i => i : i => (i + 6) % 7;
    const values = d3.rollup(data, v=>d3.sum(v, d=>y(d)), d=>countDay(x(d).getUTCDay()), d=>x(d).getUTCHours())
    const maxV = d3.max(values.values(), v => d3.max(v.values()))
    color = d3.scaleSequential([1,maxV], color)

    const radiusScale = d3.scaleSqrt().domain([1,maxV]).range([1.5,cellSize/2])

    const weekDays = weekday === "weekday" ? 5 : 7;
    const height = 25 + ((cellSize + cellPadding) * weekDays);
    const width = 25 + ((cellSize + cellPadding) * 24) + (cellPadding*2)
    
    // Construct formats.
  
  
    const svg = d3.create("svg")
        .attr("width", width)
        // .attr("height", height * years.length)
        .attr("height", height)
        // .attr("viewBox", [0, 0, width, height * years.length])
        .attr("viewBox", [0, 0, width, height])
        .attr("style", "max-width: 100%; height: auto; height: intrinsic;")
        .attr("font-family", "sans-serif")
        .attr("font-size", 10);
  
    svg.append("g")
      .attr("text-anchor", "end")
      .attr("transform", "translate(10,25)")
      .selectAll("text")
      .data(weekday === "weekday" ? d3.range(1, 6) : d3.range(7))
      .join("text")
        // .attr("x", -5)
        .attr("y", i => (countDay(i) + 0.5) * (cellSize + cellPadding))
        .attr("dy", "0.31em")
        .text(formatDay);
  
    const cell = svg.append("g")
      .attr("transform", "translate(15,25)")
      // .attr("transform", (d, i) => `translate(15,${height * i + cellSize * 1.5})`);
      .selectAll("circle")
      .data(d3.utcHours(new Date("1970-01-04T00:00Z"), new Date("1970-01-11T00:00Z")))
      .join("circle")
        .attr("class", "circle")
        .attr("width", cellSize)
        .attr("height", cellSize)
        .attr("cx", d => d.getUTCHours() * (cellSize + cellPadding) + (cellSize/2))
        .attr("cy", d => countDay(d.getUTCDay()) * (cellSize + cellPadding) + (cellSize/2))
        .attr("r", d => {
          let v = values.get(countDay(d.getUTCDay()))?.get(d.getUTCHours())
          let r = v == undefined ? 1 : radiusScale(v)
          return r
        })
        .attr("fill", d => {
          let v = values.get(countDay(d.getUTCDay()))?.get(d.getUTCHours())
          return v == undefined ? emptyColor : color(v)
        });

        if (title) cell.append("title")
        .text(title);

      let timeScale = d3.scaleUtc().domain([new Date("1970-01-04T00:00Z"), new Date("1970-01-04T23:00Z")])
        .range([cellSize/2, (cellSize + cellPadding) * 23.5])
      const axis = svg.append("g")
        .attr("transform", "translate(15,20)")

      axis.append("g").call(d3.axisTop(timeScale).ticks(8, formatHour))
  
    return Object.assign(svg.node(), {scales: {color}});
}
export {Punchcard as default}