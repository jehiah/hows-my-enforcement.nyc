// Inspired from
// https://observablehq.com/@d3/calendar-view
function FYCalendar(data, {
    x = ([x]) => x, // given d in data, returns the (temporal) x-value
    y = ([, y]) => y, // given d in data, returns the (quantitative) y-value
    title, // given d in data, returns the title text
    width = undefined, // width of the chart, in pixels
    Xmin=undefined,
    Xmax=undefined,
    cellSize = 17, // width and height of an individual day, in pixels
    monthPadding = 6, // padding between months
    cellPadding=1,
    weekday = "monday", // either: weekday, sunday, or monday
    formatDay = i => "SMTWTFS"[i], // given a day number in [0, 6], the day-of-week label
    formatMonth = "%b", // format specifier string for months (above the chart)
    // yFormat, // format specifier string for values (in the title)
    color = undefined,
    emptyColor = undefined,
  } = {}) {
    // Compute values.

    // const X = d3.map(data, x);
    // const Y = d3.map(data, y);
    // const I = d3.range(X.length);
    
    const countDay = weekday === "sunday" ? i => i : i => (i + 6) % 7;
    const timeWeek = weekday === "sunday" ? d3.utcSunday : d3.utcMonday;
    // const timeMonth = d3.utcMonth
    const weekDays = weekday === "weekday" ? 5 : 7;
    const height = 25 + ((cellSize + cellPadding) * weekDays);
    if (width == undefined ) {
      width = 15 + (d3.utcMonth.count(d3.utcMonth(Xmin), Xmax) * monthPadding) + ((timeWeek.count(timeWeek(Xmin), Xmax) + 1) * ( cellSize + cellPadding ))
    }
    
    // Construct formats.
    formatMonth = d3.utcFormat(formatMonth);
  
    // // Compute titles.
    // if (title === undefined) {
    //   const formatDate = d3.utcFormat("%B %-d, %Y");
    //   const formatValue = color.tickFormat(100, yFormat);
    //   title = i => `${formatDate(X[i])}\n${formatValue(Y[i])}`;
    // } else if (title !== null) {
    //   const T = d3.map(data, title);
    //   title = i => T[i];
    // }

    // // Group the index by year, in reverse input order. (Assuming that the input is
    // // chronological, this will show years in reverse chronological order.)
    // const years = d3.groups(I, i => X[i].getUTCFullYear()).reverse();
    // // console.log("years", years)
  
    // function pathMonth(t) {
    //   const d = Math.max(0, Math.min(weekDays, countDay(t.getUTCDay())));
    //   const w = timeWeek.count(d3.utcYear(t), t);
    //   return `${d === 0 ? `M${w * cellSize},0`
    //       : d === weekDays ? `M${(w + 1) * cellSize},0`
    //       : `M${(w + 1) * cellSize},0V${d * cellSize}H${w * cellSize}`}V${weekDays * cellSize}`;
    // }
  
    const container = document.createElement("div")

    const svg = d3.create("svg")
        .attr("width", width)
        // .attr("height", height * years.length)
        .attr("height", height)
        // .attr("viewBox", [0, 0, width, height * years.length])
        .attr("viewBox", [0, 0, width, height])
        .attr("style", "max-width: 100%; height: auto; height: intrinsic;")
        .attr("font-family", "sans-serif")
        .attr("font-size", 10);
    container.appendChild(svg.node())
    const tooltip = d3.select( container.appendChild(document.createElement("div")))
  .attr("class", "tooltip")
  .style("position", "absolute")
  .style("pointer-events", "none")
  .style("opacity", 0);
  
    // const year = svg.selectAll("g")
    //   .data([[1, I]])
    //   .join("g")
    //     .attr("transform", (d, i) => `translate(15,${height * i + cellSize * 1.5})`);
  
    // year.append("text")
    //     .attr("x", -5)
    //     .attr("y", -5)
    //     .attr("font-weight", "bold")
    //     .attr("text-anchor", "end")
    //     .text(([key]) => key);
  
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
      .selectAll("rect")
      .data(data)
      .join("rect")
        .attr("class", "cell")
        .attr("width", cellSize)
        .attr("height", cellSize)
        .attr("x", d => (d3.utcMonth.count(d3.utcMonth(Xmin), x(d)) * monthPadding) + (timeWeek.count(timeWeek(Xmin), x(d)) * (cellSize + cellPadding)))
        .attr("y", d => countDay(x(d).getUTCDay()) * (cellSize + cellPadding))
        .attr("fill", d => y(d) == 0 ? emptyColor : color(y(d)))
        .on("mouseover", (event, d) => {
          event.target.classList.add("highlighted")

          if ( y(d) == 0) { 
            tooltip.transition()
            .duration(50)
            .style("opacity", 0);
            tooltip.html('')
            return
          }

          
          // d3.select(this).classed("highlighted", true);
          // console.log(d, y(d));
          // Show tooltip and update its content based on the hovered item
          tooltip.transition()
            .duration(100)
            .style("opacity", 0.9);
          tooltip.html(title(d))
            .style("left", (event.pageX + 10) + "px")
            .style("top", (event.pageY + 10) + "px");
        })
        .on("mousemove", _ =>  {
          // Update tooltip position on mousemove
          tooltip.style("left", (event.pageX + 10) + "px")
            .style("top", (event.pageY + 10) + "px");
        })
        .on("mouseout", event =>  {
          event.target.classList.remove("highlighted")
          // Hide tooltip on mouseout
          tooltip.transition()
            .duration(100)
            .style("opacity", 0);
        })
        if (title) cell.append("title")
        .text(title);


    // const cell = year.append("g")
    //   .selectAll("rect")
    //   .data(weekday === "weekday"
    //       ? ([, I]) => I.filter(i => ![0, 6].includes(X[i].getUTCDay()))
    //       : ([, I]) => I)
    //   .join("rect")
    //     .attr("class", "cell")
    //     .attr("width", cellSize - 1)
    //     .attr("height", cellSize - 1)
    //     .attr("x", i => timeWeek.count(d3.utcMonth(X[0]), X[i]) * cellSize + 0.5)
    //     .attr("y", i => countDay(X[i].getUTCDay()) * cellSize + 0.5)
    //     .attr("fill", i => color(Y[i]));

    // const XUTC = X.map(d => d3.utcDay(d).getTime())
    // const emptyDays = d3.utcDay.range(d3.utcMonth(X[0]), X[X.length-1]).filter(d => XUTC.indexOf(d.getTime()) == -1)

    // const cell2 = year.append("g")
    //     .selectAll("rect.cell-empty")
    //     .data(_ => emptyDays)
    //     .join("rect")
    //       .attr("class", "cell-empty")
    //       .attr("width", cellSize - 1)
    //       .attr("height", cellSize - 1)
    //       .attr("x", d => timeWeek.count(d3.utcMonth(X[0]), d) * cellSize + 0.5)
    //       .attr("y", d => countDay(d.getUTCDay()) * cellSize + 0.5)
    //       .attr("fill", "#eee");
    // cell2.append("title")
    //     .text(d3.utcFormat("%B %-d, %Y"))

    const month = svg.append("g")
    .attr("transform", "translate(15,15)")
    .selectAll("g")
      .data(d3.utcMonth.range(Xmin, Xmax))
      .join("g");

      // .attr("x", d => timeWeek.count(d3.utcMonth(X[0]), timeWeek.ceil(d)) * cellSize + 2)
      // month.filter((d, i) => i).append("text")
        // .attr("fill", "none")
        // .attr("stroke", "#fff")
        // .attr("stroke-width", 3)
        // .attr("d", pathMonth)
        // .text()  
    month.append("text")
    .attr("x", d => (d3.utcMonth.count(d3.utcMonth(Xmin), d) * monthPadding) + (timeWeek.count(timeWeek(Xmin), d) * (cellSize + cellPadding)))
    // .attr("x", d => timeWeek.count(d3.utcMonth(X[0]), timeWeek.ceil(d)) * cellSize + 2)
      // .attr("y", -5)
      .text(d => formatMonth(d));
  
    return Object.assign(container, {scales: {color}});
}
export {FYCalendar as default}