<template>
  <div>
    <b-modal :active.sync="detailsModal.active" :on-cancel="resetDetailsModal" has-modal-card>
      <div class="modal-card" style="width:25em">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ detailsModal.vm }} VM Details</p>
        </header>
        <section class="modal-card-body">
          <p>Hostname: {{ detailsModal.vm }}</p>
          <p>The quick brown fox jumped over the lazy dog.</p>
        </section>
        <footer class="modal-card-foot">
        </footer>
      </div>
    </b-modal>
    <hr>
    <div class="level is-vcentered">
      <div class="level-item">
        <span style="font-weight: bold; font-size: x-large;">State of Health Board for Experiment: {{ this.$route.params.id }}</span>&nbsp;
      </div>
    </div>
    <div class="columns is-vcentered">
      <div class="column" />
      <div class="column">
        <b-radio v-model="radioButton" native-value="running" type="is-light">Running</b-radio>
      </div>
      <div class="column">
        <b-radio v-model="radioButton" native-value="notrunning" type="is-light">Not running</b-radio>
      </div>
      <div class="column">
        <b-radio v-model="radioButton" native-value="notboot" type="is-light">Not booted</b-radio>
      </div>
      <div class="column">
        <b-radio v-model="radioButton" native-value="notdeploy" type="is-light">Not deployed</b-radio>
      </div>
      <div class="column">
        <b-button @click="resetNetwork" type="is-light">Refresh Network</b-button>
      </div>
      <div class="column" />
    </div>
    <div style="margin-top: 10px; border: 2px solid whitesmoke; background: #333;">
      <div v-if="nodes == null">
        <section class="hero is-light is-bold is-large">
          <div class="hero-body">
            <div class="container" style="text-align: center">
              <h1 class="title">
                There are no nodes matching your search criteria!
              </h1>
                <b-button type="is-success" outlined @click="resetNetwork()">Refresh Network</b-button>
            </div>
          </div>
        </section>
      </div>
      <div v-else id="graph"></div>
    </div>
    <br>
    <div class="columns is-vcentered">
      <div class="column" />
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <img :src="vlan" style="width:20px;height:20px;" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">VLAN Segment</span>
          </div>
        </div>
      </div>
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <b-icon icon="circle" style="color: green" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">Running</span>
          </div>
        </div>
      </div>
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <b-icon icon="circle" style="color: red" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">Not running</span>
          </div>
        </div>
      </div>
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <b-icon icon="circle" style="color: blue" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">Not booted</span>
          </div>
        </div>
      </div>
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <b-icon icon="circle" style="color: yellow" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">Not deployed</span>
          </div>
        </div>
      </div>
      <div class="column" />
    </div>
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
    async updateNetwork (filter = '') {
      let url = 'experiments/' + this.$route.params.id + '/soh';

      if (filter) {
        url = url + '?statusFilter=' + filter;
      }

      try {
        let resp = await this.$http.get( url );
        let state = await resp.json();

        this.running = state.started;
        this.nodes = state.nodes;
        this.edges = state.edges;
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
      return "url(#" + node.image + ")";
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
      if (this.nodes == null) {
        return;
      }

      const nodes = this.nodes.map(d => Object.create(d));
      const links = this.edges.map(d => Object.create(d));

      const width = 600;
      const height = 300;

      const simulation = d3.forceSimulation(nodes)
        .force("link", d3.forceLink(links).id(d => d.id))
        .force("charge", d3.forceManyBody())
        .force("center", d3.forceCenter(width / 2, height / 2));

      d3.select("#graph").select("svg").remove();

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
        .attr("width", 30)
        .attr("height", 30)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "centos")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", CentOS)
        .attr("width", 30)
        .attr("height", 30)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "rhel")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", RedHat)
        .attr("width", 30)
        .attr("height", 30)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "windows")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", Windows)
        .attr("width", 30)
        .attr("height", 30)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "Router")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", Router)
        .attr("width", 30)
        .attr("height", 30)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "Firewall")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", Firewall)
        .attr("width", 30)
        .attr("height", 30)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "Printer")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", Printer)
        .attr("width", 30)
        .attr("height", 30)
        .attr("x", 0)
        .attr("y", 0);

      defs.append("svg:pattern")
        .attr("id", "Switch")
        .attr("width", 50)
        .attr("height", 50)
        .append("svg:image")
        .attr("xlink:href", VLAN)
        .attr("width", 30)
        .attr("height", 30)
        .attr("x", 0)
        .attr("y", 0);

      const node = svg.append("g")
        .selectAll("circle")
        .data(nodes)
        .join("circle")
        .attr("class", "circle")
        .attr("stroke", this.updateNodeColor)
        .attr("stroke-width", 1.5)
        // .attr("r", 10)
        // .attr("fill", this.updateNodeImage)
        .attr("r", 5)
        .attr("fill", this.updateNodeColor)
        .on( 'mouseenter', this.entered)
        .on( 'mouseleave', this.exited)
        .on( 'click', this.clicked)
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

    entered(e, n) {
      let circle = d3.select(e.target);
      console.log(circle);
      circle
        .transition()
        .attr("r", 15)
        .attr("fill", () => this.updateNodeImage(n));
    },

    exited(e, n) {
      let circle = d3.select(e.target);
      circle
        .transition()
        .attr("r", 5)
        .attr("fill", () => this.updateNodeColor(n));
    },

    clicked(e, n) {
      this.detailsModal.active = true;
      this.detailsModal.vm = n.label;
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
  
    async resetNetwork () {
      this.radioButton = '';
      await this.updateNetwork();
      this.generateGraph();
    },

    resetDetailsModal () {
        this.detailsModal = {
          active: false,
          vm: ''
        }
      },
  },

  watch: {
    radioButton: async function ( filter ) {
      if ( filter != '' ) {
        await this.updateNetwork(filter);
        this.generateGraph();
      }
    }
  },

  data() {
    return {
      running: false,
      nodes: [],
      edges: [],
      radioButton: '',
      vlan: VLAN,
      detailsModal: {
        active: false,
        vm: ''
      }
    };
  }
}
</script>

<style scoped>
  label.radio:hover {
    color: whitesmoke;
  }
</style>