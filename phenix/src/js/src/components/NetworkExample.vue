<template>
  <div>
    <div id="graph" style="margin-top: 10px; border: 2px solid whitesmoke; background: #333;"></div>
    <div>This is the key!!!</div>
  </div>
</template>

<script>
import * as d3 from "d3";

import Linux    from "@/assets/linux.svg";
import CentOS   from "@/assets/centos.svg";
import RedHat   from "@/assets/redhat.svg";
import Windows  from "@/assets/windows.svg";
import Router   from "@/assets/router.svg";
import Firewall from "@/assets/firewall.svg";
import Printer  from "@/assets/printer.svg";
import VLAN     from "@/assets/vlan.svg";

export default {
  async created () {
    await this.updateNetwork();
    this.generateGraph();
  },

  methods: {
    async updateNetwork () {
      try {
        // let resp = await this.$http.get( 'experiments/' + this.$route.params.id + '/soh' );
        let resp = await this.$http.get( 'experiments/test-01/soh' );
        let state = await resp.json();

        this.nodes = state.nodes;
        this.edges = state.edges;

        // TODO: remove this once server-side is updated.
        this.edges.forEach(e => {
          e.source = e.from;
          e.target = e.to;
        });
      } catch {
        this.$buefy.toast.open ({
          message: 'Getting Network Failed',
          type: 'is-danger',
          duration: 4000
        });
      } finally {
        this.isWaiting = false;
      }
    },

    updateNodeImage(node) {
      if ( node.image == 'interface' ) {
        return "url(#" + node.image + ")";
      }

      // TODO: remove this once server-side is updated.
      return "url(#linux)";
    },

    updateNodeColor(node) {
      if (!this.running) {
        return
      }

      const colors = {
        "running":    "green",
        "notrunning": "red",
        "notboot":    "blue",
        "notdeploy":  "yellow",
      }

      return colors[node.status];
    },

    generateGraph() {
      const links = this.edges.map(d => Object.create(d));
      const nodes = this.nodes.map(d => Object.create(d));
      const width = 600;
      const height = 300;

      const simulation = d3.forceSimulation(nodes)
        .force("link", d3.forceLink(links).id(d => d.id))
        .force("charge", d3.forceManyBody())
        .force("center", d3.forceCenter(width / 2, height / 2));

      const svg = d3.select("#graph").append("svg")
        .attr("viewBox", [0, 0, width, height]);

      const link = svg.append("g")
        .attr("stroke", "#999")
        .attr("stroke-opacity", 0.6)
        .selectAll("line")
        .data(links)
        .join("line")
        .attr("stroke-width", d => Math.sqrt(d.value));

      const defs = svg.append("svg:defs");

      defs.append("svg:pattern")
        .attr("id", "linux")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", Linux)
        .attr("width", 20)
        .attr("height", 20)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "centos")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", CentOS)
        .attr("width", 20)
        .attr("height", 20)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "rhel")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", RedHat)
        .attr("width", 20)
        .attr("height", 20)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "windows")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", Windows)
        .attr("width", 20)
        .attr("height", 20)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "router")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", Router)
        .attr("width", 20)
        .attr("height", 20)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "firewall")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", Firewall)
        .attr("width", 20)
        .attr("height", 20)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "printer")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", Printer)
        .attr("width", 20)
        .attr("height", 20)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "interface")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", VLAN)
        .attr("width", 20)
        .attr("height", 20)
        .attr("x", 0)
        .attr("y", 0);

      const node = svg.append("g")
        .selectAll("circle")
        .data(nodes)
        .join("circle")
        .attr("class", "circle")
        .attr("stroke", this.updateNodeColor)
        .attr("stroke-width", 1.5)
        .attr("r", 10)
        .attr("fill", this.updateNodeImage)
        .call(this.drag(simulation));

      node.append("title").text(d => d.label);

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
    },

    color(d) {
      const scale = d3.scaleOrdinal(d3.schemeCategory10);
      return d => scale(d.group);
    },

    drag(simulation) {
      function dragstarted(event) {
        if (!event.active) simulation.alphaTarget(0.3).restart();
        event.subject.fx = event.subject.x;
        event.subject.fy = event.subject.y;
      }
      
      function dragged(event) {
        event.subject.fx = event.x;
        event.subject.fy = event.y;
      }
      
      function dragended(event) {
        if (!event.active) simulation.alphaTarget(0);
        event.subject.fx = null;
        event.subject.fy = null;
      }
      
      return d3.drag()
          .on("start", dragstarted)
          .on("drag", dragged)
          .on("end", dragended);
    },
  },

name: "App",

data() {
  return {
    running: false,
    nodes: [],
    edges: [],
  };
}

}
</script>

// TODO: add button to experiment table; add button to running/stopped experiment component;
// move code to RunningVms and rename component; use StateOfHealth.vue with this content.