<template>
  <div>
    <b-modal :active.sync="detailsModal.active" :on-cancel="resetDetailsModal" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ detailsModal.vm }} VM Details</p>
        </header>
        <section class="modal-card-body">
          <template v-if="detailsModal.soh">
            <p>The following state of health has been reported for the {{ detailsModal.vm }} VM.</p>
            <br>
            <div v-if="detailsModal.soh.reachability">
              <p class="title is-5">Reachability</p>
              <b-table
                :data="detailsModal.soh.reachability"
                default-sort="host">
                <template slot-scope="props">
                  <b-table-column field="hostname" label="Host" sortable>
                    {{ props.row.hostname }}
                  </b-table-column>
                  <b-table-column field="timestamp" label="Timestamp" sortable>
                    {{ props.row.timestamp }}
                  </b-table-column>
                  <b-table-column field="error" label="Error" sortable>
                    {{ props.row.error }}
                  </b-table-column>
                </template>
              </b-table>
              <br>
            </div>
            <div v-if="detailsModal.soh.processes">
              <p class="title is-5">Processes</p>
              <b-table
                :data="detailsModal.soh.processes"
                default-sort="process">
                <template slot-scope="props">
                  <b-table-column field="process" label="Process" sortable>
                    {{ props.row.process }}
                  </b-table-column>
                  <b-table-column field="timestamp" label="Timestamp" sortable>
                    {{ props.row.timestamp }}
                  </b-table-column>
                  <b-table-column field="error" label="Error" sortable>
                    {{ props.row.error }}
                  </b-table-column>
                </template>
              </b-table>
              <br>
            </div>
            <div v-if="detailsModal.soh.listeners">
              <p class="title is-5">Listeners</p>
              <b-table
                :data="detailsModal.soh.listeners"
                default-sort="listener">
                <template slot-scope="props">
                  <b-table-column field="listener" label="Listener" sortable>
                    {{ props.row.listener }}
                  </b-table-column>
                  <b-table-column field="timestamp" label="Timestamp" sortable>
                    {{ props.row.timestamp }}
                  </b-table-column>
                  <b-table-column field="error" label="Error" sortable>
                    {{ props.row.error }}
                  </b-table-column>
                </template>
              </b-table>
              <br>
            </div>
          </template>
          <template v-else>
            <p>There is no state of health data available for the {{ detailsModal.vm }} VM.</p>
          </template>
        </section>
        <footer class="modal-card-foot">
        </footer>
      </div>
    </b-modal>
    <hr>
    <div class="level is-vcentered">
      <div class="level-item">
        <span style="font-weight: bold; font-size: x-large;">State of Health Board for Experiment: {{ this.$route.params.id }}</span>
      </div>
    </div>
    <div>
      <b-tabs>
        <b-tab-item label="Topology Graph">
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
            <div class="column">
              <b-button @click="execSoH" type="is-light" disabled>Manual Refresh</b-button>
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
            <div class="column is-one-fifth">
              <div class="columns is-variable is-1">
                <div class="column has-text-right">
                  <img :src="vlan" style="width:20px;height:20px;" />
                </div>
                <div class="column">
                  <span style="color: whitesmoke;">VLAN Segment</span>
                </div>
              </div>
            </div>
            <div class="column is-one-fifth">
              <div class="columns is-variable is-1">
                <div class="column has-text-right">
                  <b-icon icon="circle" style="color: #4F8F00" />
                </div>
                <div class="column">
                  <span style="color: whitesmoke;">Running</span>
                </div>
              </div>
            </div>
            <div class="column is-one-fifth">
              <div class="columns is-variable is-1">
                <div class="column has-text-right">
                  <b-icon icon="circle" style="color: #941100" />
                </div>
                <div class="column">
                  <span style="color: whitesmoke;">Not running</span>
                </div>
              </div>
            </div>
            <div class="column" />
          </div>
          <div class="columns is-vcentered">
            <div class="column" />
            <div class="column is-one-fifth">
              <div class="columns is-variable is-1">
                <div class="column has-text-right">
                  <b-icon icon="circle" style="color: #005493" />
                </div>
                <div class="column">
                  <span style="color: whitesmoke;">Not booted</span>
                </div>
              </div>
            </div>
            <div class="column is-one-fifth">
              <div class="columns is-variable is-1">
                <div class="column has-text-right">
                  <b-icon icon="circle" style="color: #FFD479" />
                </div>
                <div class="column">
                  <span style="color: whitesmoke;">Not deployed</span>
                </div>
              </div>
            </div>
            <div class="column is-one-fifth">
              <div class="columns is-variable is-1">
                <div class="column has-text-right">
                  <b-icon icon="circle" style="color: black" />
                </div>
                <div class="column">
                  <span style="color: whitesmoke;">Experiment stopped</span>
                </div>
              </div>
            </div>
            <div class="column" />
          </div>
        </b-tab-item>
        <b-tab-item label="Network Volume">
          <div class="columns is-centered">
            <div class="column is-one-quarter">
              <h3 class="title is-4">This is all mock data.</h3>
            </div>
          </div>
          <div style="margin-top: 10px; border: 2px solid whitesmoke; background: #333;">
            <div id="chord"></div>
          </div>
        </b-tab-item>
        <b-tab-item label="SoH Messages">
          <div v-if="messages" class="columns is-centered is-multiline">
            <div v-for="( n, index ) in nodes" :key="index">
              <div class="column is-one-half">
                <div v-if="n.soh">
                  <h3 class="title is-3">{{ n.label }}</h3>
                  <div v-if="n.soh.reachability">
                    <h3 class="title is-5">Reachability</h3>
                    <b-table
                      :data="n.soh.reachability"
                      default-sort="host">
                      <template slot-scope="props">
                        <b-table-column field="hostname" label="Host" sortable>
                          {{ props.row.hostname }}
                        </b-table-column>
                        <b-table-column field="timestamp" label="Timestamp" sortable>
                          {{ props.row.timestamp }}
                        </b-table-column>
                        <b-table-column field="error" label="Error" sortable>
                          {{ props.row.error }}
                        </b-table-column>
                      </template>
                    </b-table>
                    <br>
                  </div>
                  <div v-if="n.soh.processes">
                    <h3 class="title is-5">Processes</h3>
                    <b-table
                      :data="n.soh.processes"
                      default-sort="process">
                      <template slot-scope="props">
                        <b-table-column field="process" label="Process" sortable>
                          {{ props.row.process }}
                        </b-table-column>
                        <b-table-column field="timestamp" label="Timestamp" sortable>
                          {{ props.row.timestamp }}
                        </b-table-column>
                        <b-table-column field="error" label="Error" sortable>
                          {{ props.row.error }}
                        </b-table-column>
                      </template>
                    </b-table>
                    <br>
                  </div>
                  <div v-if="n.soh.listeners">
                    <h3 class="title is-5">Listeners</h3>
                    <b-table
                      :data="n.soh.listeners"
                      default-sort="listener">
                      <template slot-scope="props">
                        <b-table-column field="listener" label="Listener" sortable>
                          {{ props.row.listener }}
                        </b-table-column>
                        <b-table-column field="timestamp" label="Timestamp" sortable>
                          {{ props.row.timestamp }}
                        </b-table-column>
                        <b-table-column field="error" label="Error" sortable>
                          {{ props.row.error }}
                        </b-table-column>
                      </template>
                    </b-table>
                    <br>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div v-else>
            <section class="hero is-light is-bold is-large">
              <div class="hero-body">
                <div class="container" style="text-align: center">
                  <h1 class="title">
                    There are no state of health messages avaiable!
                  </h1>
                </div>
              </div>
            </section>
          </div>
        </b-tab-item>
      </b-tabs>
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
  async beforeDestroy () {
    this.$options.sockets.onmessage = null;
  },

  async created () {
    this.$options.sockets.onmessage = this.handler;
    await this.updateNetwork();
    this.generateGraph();
    this.generateChord();
  },

  methods: {
    handler ( event ) {
      event.data.split( /\r?\n/ ).forEach( m => {
        let msg = JSON.parse( m );
        this.handle( msg );
      });
    },

    handle ( msg ) {
      switch ( msg.resource.type ) {
        case 'experiment': {
          if ( msg.resource.name != this.$route.params.id ) {
            return;
          }

          switch ( msg.resource.action ) {
            case 'stop':
            case 'start': {
              this.resetNetwork();
            }
          }

          break;
        }

        case 'experiment/vm': {
          // exp_name/vm_name
          let resource = msg.resource.name.split( '/' );
          let expName  = resource[0];
          let vmName   = resource[1];

          // Ignore this broadcast if it's not for this experiment.
          if ( expName != this.$route.params.id ) {
            return;
          }

          switch ( msg.resource.action ) {
            case 'stop': {
              for ( let i = 0; i < this.nodes.length; i++ ) {
                if ( this.nodes[i].label == vmName ) {
                  this.nodes[i].status = 'notrunning';
                  d3.selectAll('circle').attr( "fill", this.updateNodeColor );
                }
              }

              break;
            }
            case 'start': {
              for ( let i = 0; i < this.nodes.length; i++ ) {
                if ( this.nodes[i].label == vmName ) {
                  this.nodes[i].status = 'running';
                  d3.selectAll('circle').attr( "fill", this.updateNodeColor );
                }
              }
              
              break;
            }
            case 'delete': {
              for ( let i = 0; i < this.nodes.length; i++ ) {
                if ( this.nodes[i].label == vmName ) {
                  this.nodes[i].status = 'notdeploy';
                  d3.selectAll('circle').attr( "fill", this.updateNodeColor );
                }
              }
              
              break;
            }
          }
        }
      }
    },

    async updateNetwork ( filter = '' ) {
      let url = 'experiments/' + this.$route.params.id + '/soh';

      if ( filter ) {
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

      // check if there are any SoH messages; set messages
      // true if so and break
      for ( let i = 0; i < this.nodes.length; i++ ) {
        if ( this.nodes[i].soh != null ) {
          this.messages = true;
          break;
        } 
      }
    },

    updateNodeImage( node ) {
      return "url(#" + node.image + ")";
    },

    updateNodeBorder( node ) {
      if ( node.soh ) {
        return '#FF9900'; // orange
      }

      return this.updateNodeColor( node );
    },

    updateNodeColor( node ) {
      if ( !this.running ) {
          if ( node.status == "ignore" ) {
          return "url(#Switch)";
        }

        return;
      }

      if ( node.status == "ignore" ) {
        return "url(#Switch)";
      }

      const colors = {
        "running":    "#4F8F00", // green
        "notrunning": "#941100", // red
        "notboot":    "#005493", // blue
        "notdeploy":  "#FFD479", // yellow
      }

      return colors[ node.status ];
    },

    generateGraph () {
      if ( this.nodes == null ) {
        return;
      }

      const nodes = this.nodes.map( d => Object.create( d ) );
      const links = this.edges.map( d => Object.create( d ) );

      const width = 600;
      const height = 400;

      const simulation = d3.forceSimulation( nodes )
        .force( "link", d3.forceLink( links ).id( d => d.id ) )
        .force( "charge", d3.forceManyBody() )
        .force( "center", d3.forceCenter( width / 2, height / 2 ) );

      d3.select( "#graph" ).select( "svg" ).remove();

      const svg = d3.select( "#graph" ).append( "svg" )
        .attr( "viewBox", [ 0, 0, width, height ] );

      const g = svg.append( "g" );

      svg.call( d3.zoom()
        .extent( [ [ 0, 0 ], [ width, height ] ] )
        .scaleExtent([  -5, 5 ] )
        .on( "zoom", function ( { transform } ) {
          g.attr( "transform", transform );
        })
      );

      const link = g.append( "g" )
        .attr( "stroke", "#999" )
        .attr( "stroke-opacity", 0.6 )
        .selectAll( "line" )
        .data( links )
        .join( "line" )
        .attr( "stroke-width", d => Math.sqrt( d.value ) );

      const defs = svg.append( "svg:defs" );

      defs.append( "svg:pattern" )
        .attr( "id", "linux" )
        .attr( "width", 50 )
        .attr( "height", 50 )
        .append( "svg:image" )
        .attr( "xlink:href", Linux )
        .attr( "width", 30 )
        .attr( "height", 30 )
        .attr( "x", 0 )
        .attr( "y", 0 );

      defs.append( "svg:pattern" )
        .attr( "id", "centos" )
        .attr( "width", 50 )
        .attr( "height", 50 )
        .append( "svg:image" )
        .attr( "xlink:href", CentOS )
        .attr( "width", 30 )
        .attr( "height", 30 )
        .attr( "x", 0 )
        .attr( "y", 0 );

      defs.append( "svg:pattern" )
        .attr( "id", "rhel" )
        .attr( "width", 50 )
        .attr( "height", 50 )
        .append( "svg:image" )
        .attr( "xlink:href", RedHat )
        .attr( "width", 30 )
        .attr( "height", 30 )
        .attr( "x", 0 )
        .attr( "y", 0 );

      defs.append( "svg:pattern" )
        .attr( "id", "windows" )
        .attr( "width", 50 )
        .attr( "height", 50 )
        .append( "svg:image" )
        .attr( "xlink:href", Windows )
        .attr( "width", 30 )
        .attr( "height", 30 )
        .attr( "x", 0 )
        .attr( "y", 0 );

      defs.append( "svg:pattern" )
        .attr( "id", "Router" )
        .attr( "width", 50 )
        .attr( "height", 50 )
        .append( "svg:image" )
        .attr( "xlink:href", Router )
        .attr( "width", 30 )
        .attr( "height", 30 )
        .attr( "x", 0 )
        .attr( "y", 0 );

      defs.append( "svg:pattern" )
        .attr( "id", "Firewall" )
        .attr( "width", 50 )
        .attr( "height", 50 )
        .append( "svg:image" )
        .attr( "xlink:href", Firewall )
        .attr( "width", 30 )
        .attr( "height", 30 )
        .attr( "x", 0 )
        .attr( "y", 0 );

      defs.append( "svg:pattern" )
        .attr( "id", "Printer" )
        .attr( "width", 50 )
        .attr( "height", 50 )
        .append( "svg:image" )
        .attr( "xlink:href", Printer )
        .attr( "width", 30 )
        .attr( "height", 30 )
        .attr( "x", 0 )
        .attr( "y", 0 );

      defs.append( "svg:pattern" )
        .attr( "id", "Switch" )
        .attr( "width", 50 )
        .attr( "height", 50 )
        .append( "svg:image" )
        .attr( "xlink:href", VLAN )
        .attr( "width", 10 )
        .attr( "height", 10 )
        .attr( "x", 0 )
        .attr( "y", 0 );

      const node = g.append( "g" )
        .selectAll( "circle" )
        .data( nodes )
        .join( "circle" )
        .attr( "class", "circle" )
        .attr( "stroke", this.updateNodeBorder )
        .attr( "stroke-width", 1.5 )
        .attr( "r", 5 )
        .attr( "fill", this.updateNodeColor )
        .attr( "width", 5 )
        .attr( "height", 5 )
        .on( 'mouseenter', this.entered )
        .on( 'mouseleave', this.exited )
        .on( 'click', this.clicked )
        .call( this.drag( simulation ) );

      node.append( "title" ).text( d => d.label );

      simulation.on( "tick", () => {
        link
          .attr( "x1", d => d.source.x )
          .attr( "y1", d => d.source.y )
          .attr( "x2", d => d.target.x )
          .attr( "y2", d => d.target.y );

        node
          .attr( "cx", d => d.x )
          .attr( "cy", d => d.y) ;
      });
    },

    entered ( e, n ) {
      if ( n.image == "Switch" ) {
        return;
      }

      let circle = d3.select( e.target );

      circle
        .transition()
        .attr( "r", 15 )
        .attr( "fill", () => this.updateNodeImage( n ) );
    },

    exited ( e, n ) {
      let circle = d3.select( e.target );

      circle
        .transition()
        .attr( "r", 5 )
        .attr( "fill", () => this.updateNodeColor( n ) );
    },

    clicked ( e, n ) {
      this.detailsModal.active = true;
      this.detailsModal.vm = n.label;
      this.detailsModal.soh = n.soh;
    },

    color ( d ) {
      const scale = d3.scaleOrdinal( d3.schemeCategory10 );
      return d => scale( d.group );
    },

    drag ( simulation ) {
      function dragstarted ( event ) {
        if ( !event.active ) simulation.alphaTarget( 0.3 ).restart();
        event.subject.fx = event.subject.x;
        event.subject.fy = event.subject.y;
      }
      
      function dragged ( event ) {
        event.subject.fx = event.x;
        event.subject.fy = event.y;
      }
      
      function dragended ( event ) {
        if ( !event.active ) simulation.alphaTarget( 0 );
        event.subject.fx = null;
        event.subject.fy = null;
      }
      
      return d3.drag()
        .on( "start", dragstarted )
        .on( "drag", dragged )
        .on( "end", dragended );
    },

    generateChord () {
      const names = this.mockData.names === undefined ? d3.range(this.mockData.length) : this.mockData.names;
      const colors = this.mockData.colors === undefined ? d3.quantize(d3.interpolateRainbow, names.length) : this.mockData.colors;
      const tickStep = d3.tickStep(0, d3.sum(this.mockData.flat()), 100);
      const formatValue = d3.format(".1~%");

      const height = 900;
      const width = 900;
      const outerRadius = Math.min(width, height) * 0.5 - 60;
      const innerRadius = outerRadius - 10;
      
      const chord = d3.chord()
        .padAngle(10 / innerRadius)
        .sortSubgroups(d3.descending)
        .sortChords(d3.descending);
      
      const arc = d3.arc()
        .innerRadius(innerRadius)
        .outerRadius(outerRadius);
      
      const ribbon = d3.ribbon()
        .radius(innerRadius - 1)
        .padAngle(1 / innerRadius);

      const color = d3.scaleOrdinal(names, colors);

      function ticks (startAngle, endAngle, value) {
        const k = (endAngle - startAngle) / value;
        return d3.range(0, value, tickStep).map(value => {
          return {value, angle: value * k + startAngle};
        });
      }

      d3.select( "#chord" ).select( "svg" ).remove();

      const svg = d3.select("#chord").append("svg")
        .attr("viewBox", [-width / 2, -height / 2, width, height]);

      const chords = chord(this.mockData);

      const group = svg.append("g")
        .attr("font-size", 10)
        .attr("font-family", "sans-serif")
        .selectAll("g")
        .data(chords.groups)
        .join("g");
      
      group.append("path")
        .attr("fill", d => color(names[d.index]))
        .attr("d", arc);

      group.append("title")
        .text(d => `${names[d.index]}
          ${formatValue(d.value)}`);

      const groupTick = group.append("g")
        .selectAll("g")
        .data(ticks)
        .join("g")
        .attr("transform", d => `rotate(${d.angle * 180 / Math.PI - 90}) translate(${outerRadius},0)`);

      groupTick.append("line")
        .attr("stroke", "currentColor")
        .attr("x2", 6);

      groupTick.append("text")
        .attr("x", 8)
        .attr("dy", "0.35em")
        .attr("fill", "whitesmoke")
        .attr("transform", d => d.angle > Math.PI ? "rotate(180) translate(-16)" : null)
        .attr("text-anchor", d => d.angle > Math.PI ? "end" : null)
        .text(d => formatValue(d.value));

      group.select("text")
        .attr("font-weight", "bold")
        .attr("fill", "whitesmoke")
        .text(function(d) {
          return this.getAttribute("text-anchor") === "end"
            ? `↑ ${names[d.index]}`
            : `${names[d.index]} ↓`;
        });

      svg.append("g")
        .attr("fill-opacity", 0.8)
        .selectAll("path")
        .data(chords)
        .join("path")
        .style("mix-blend-mode", "multiply")
        .attr("fill", d => color(names[d.source.index]))
        .attr("d", ribbon)
        .append("title")
        .text(d => `${formatValue(d.source.value)} ${names[d.target.index]} → ${names[d.source.index]}${d.source.index === d.target.index ? "" : `\n${formatValue(d.target.value)} ${names[d.source.index]} → ${names[d.target.index]}`}`)
        .on("mouseover", this.showToolTip)
        .on("mouseleave", this.hideToolTip);
    },

    showToolTip (d) {
      console.log(d);
    },

    hideToolTip (d) {
      console.log(d);
    },
  
    async resetNetwork () {
      this.radioButton = '';
      await this.updateNetwork();
      this.generateGraph();
      this.generateChord();
    },

    resetDetailsModal () {
      this.detailsModal = {
        active: false,
        vm: '',
        soh: null
      }
    },

    execSoH () {
      // TODO: add WS call to execute SoH refresh
      return;
    }
  },

  watch: {
    radioButton: async function ( filter ) {
      if ( filter != '' ) {
        await this.updateNetwork( filter );
        this.generateGraph();
        this.generateChord();
      }
    }
  },

  data() {
    return {
      running: false,
      messages: false,
      nodes: [],
      edges: [],
      radioButton: '',
      vlan: VLAN,
      detailsModal: {
        active: false,
        vm: '',
        soh: null
      },
      mockData: Object.assign([
        [.0, .008859, .000554, .004430, .025471, .024363, .005537, .025471],
        [.001107, .0, .000000, .004983, .011074, .010520, .002215, .004983],
        [.000554, .002769, .0, .002215, .003876, .008306, .000554, .003322],
        [.000554, .001107, .000554, .0, .011628, .006645, .004983, .010520],
        [.002215, .004430, .000000, .002769, .0, .012182, .004983, .028239],
        [.011628, .026024, .000000, .013843, .087486, .0, .017165, .055925],
        [.000554, .004983, .000000, .003322, .004430, .008859, .0, .004430],
        [.002215, .007198, .000000, .003322, .016611, .014950, .001107, .0]
      ], {
        names: ["generator", "plc-01", "plc-02", "plc-03", "plc-04", "plc-05", "plc-06", "plc-07"],
        colors: ["#c4c4c4", "#69b40f", "#ec1d25", "#c8125c", "#008fc8", "#10218b", "#134b24", "#737373"]
      })
    };
  }
}
</script>

<style scoped>
  label.radio:hover {
    color: whitesmoke;
  }
</style>