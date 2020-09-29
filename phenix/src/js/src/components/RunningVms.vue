<!--
File: RunningVms.vue
This component is the main view of the state of health dashboard.
The State of Health (SoH) dashboard displays a topology view of all
vms in an experiment, based on the state of each vm the dashboard 
will display a distinctive icon (running, error, not deployed, not booted).
Users' options:
  Filter displayed vms
  View more information for each vm
-->
<template>
  <div class="content">
    <!--
      Popup modal @click event on each vm.
      The popup will display detail information of the vm's state
    -->
    <b-modal :active.sync="expModal.active" :on-cancel="resetExpModal" has-modal-card>
      <div class="modal-card" style="width:25em">
        <header class="modal-card-head">
          <p class="modal-card-title"> {{ expModal.vm.name ? expModal.vm.name : "unknown" }} </p>
        </header>
        <section class="modal-card-body">
          <p>Host: {{ expModal.vm.host }} </p>
          <p>IPv4: {{ expModal.vm.ipv4 | stringify }} </p>
          <p>CPU(s): {{ expModal.vm.cpus }} </p>
          <p>Memory: {{ expModal.vm.ram | ram }} </p>
          <p>Disk: {{ expModal.vm.disk }} </p>
          <p>Uptime: {{ expModal.vm.uptime | uptime }} </p>
          <p>Network(s): {{ expModal.vm.networks | stringify }} </p>
          <p>Taps: {{expModal.vm.taps }} </p>
        </section>
        <footer class="modal-card-foot buttons is-right">
          <template v-if="expModal.vm.host">
            <div>
              <!--
                Inside the popup users will have the option to 
                start/pause/kill a vm
              -->
              <template v-if="!expModal.vm.running">
                <b-tooltip label="start a vm" type="is-light">
                  <b-button class="button is-success" icon-left="play" @click="startVm ( expModal.vm.name )">
                  </b-button>
                </b-tooltip>
              </template>
              <template v-else>
                <b-tooltip label="pause a vm" type="is-light">
                  <b-button class="button is-danger" icon-left="pause" @click="pauseVm ( expModal.vm.name )">
                  </b-button>
                </b-tooltip>
              </template>
            </div>
            <div>
              &nbsp;
              <b-tooltip label="kill a vm" type="is-light">
                <b-button class="button is-danger" icon-left="trash" @click="killVm ( expModal.vm.name )">
                </b-button>
              </b-tooltip>
            </div>
          </template>
          <template v-else>
            <div>
              <p>VM not deployed!!!</p>
            </div>
          </template>
        </footer>
      </div>
    </b-modal>
    <hr>
    <b-field position="is-left">
      <p class="control">
        <h3>State of Health Board</h3>
      </p>
    </b-field>
    <div class="columns is-vcentered">
      <div class="column">
        <b-radio v-model="radioButton" native-value="running" type="is-light">Running Nodes</b-radio>
      </div>
      <div class="column">
        <b-radio v-model="radioButton" native-value="notrunning" type="is-light">Error Nodes</b-radio>
      </div>
      <div class="column">
        <b-radio v-model="radioButton" native-value="notdeploy" type="is-light">Not Deployed Nodes</b-radio>
      </div>
      <div class="column">
        <b-radio v-model="radioButton" native-value="notboot" type="is-light">Not Booted Nodes</b-radio>
      </div>
      <div class="column">
        <b-button @click="resetNetwork" type="is-light">Refresh Network</b-button>
      </div>
    </div>

    <!--
        Main display of the experiment's topology
        the topology is populated through the network variable
          network.nodes = experiment vms
          network.edge = connection to each vm
          network.options = topology view formatting
          getInfo() function enables the popup
      -->
    <network
      class="network"
      ref="network"
      :nodes="network.nodes"
      :edges="network.edges"
      :options="network.options"
      @select-node="getInfo( $event )"
    >
    </network>
    
    <!--
      Network-diagram key 
    -->
    <div class="columns is-vcentered">
      <div class="column" />
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <img :src="switchImg" style="width:25px;height:25px;" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">VLAN Segment</span>
          </div>
        </div>
      </div>
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <img :src="runningImg" style="width:25px;height:25px;" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">Running</span>
          </div>
        </div>
      </div>
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <img :src="nrunningImg" style="width:25px;height:25px;" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">In Error-state</span>
          </div>
        </div>
      </div>
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <img :src="notdeployImg" style="width:25px;height:25px;" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">Not deployed</span>
          </div>
        </div>
      </div>
      <div class="column">
        <div class="columns is-variable is-1">
          <div class="column is-one-fifth has-text-right">
            <img :src="notbootImg" style="width:25px;height:25px;" />
          </div>
          <div class="column">
            <span style="color: whitesmoke;">Not booted</span>
          </div>
        </div>
      </div>
      <div class="column" />
    </div> 
  </div>
</template>

<script>
import Network from "./Network.vue";
/*
We need to import the images that we use
throug the viewer.
*/
import Switch from "@/assets/Switch.png";
import Running from "@/assets/running.png";
import NotRunning from "@/assets/notrunning.png";
import NotBoot from "@/assets/notboot.png";
import NotDeploy from "@/assets/notdeploy.png";
import Options from "@/assets/options.png"

export default {
  components: {
    Network
  },

  async beforeDestroy () {
    this.$options.sockets.onmessage = null;
  },

  async created () {
    this.updateNetwork();
  },

  methods: {
    /*
    Function: updateNetwork
    Params: none
    return: experiment vms and experiment overall information

    First this function fetch vms' state, 
    the api 'experiments/<experiment>/soh' is use to get the overall state of 
    health of the experiment. For each vm in the experiment we need to know
    if the vm is runnin/in-error/not-deployed/not-booted, as well as the how/who
    the vm is connected.  

    Second we fetch the overall information of the vm through the api 
    'experiments'<experiment-name>/vm'. The second call will fetch necessary
    information to populated the popup. 
    */
    async updateNetwork () {
      try {
        let resp = await this.$http.get( 'experiments/' + this.$route.params.id + '/soh' );
        let state = await resp.json();
        this.network = state;
        this.onMemNetwork = JSON.parse( JSON.stringify ( this.network ) );

        resp = await this.$http.get( 'experiments/' + this.$route.params.id + '/vms' );
        state = await resp.json();
        this.experiment = state;

        /*
        update the path of the icon for each vm
        */
        this.updateImage();
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

    /*
    Function: updateImage
    Params: none
    Return: none

    Vue is not playing nice and it doesn't like 
    when you pass the full path from the api call. To overpass
    this "feature" we need to update the path of the image with
    the imports above.
    */
    updateImage () {
      for ( var node = 0; node < this.network.nodes.length; node++ ) {
        if ( this.network.nodes[ node ].image == 'interface' ) {
          this.network.nodes[ node ].image = Switch;
          this.onMemNetwork.nodes[ node ].image = Switch;
        }
        else if ( this.network.nodes[ node ].image == 'running' ) {
          this.network.nodes[ node ].image = Running;
          this.onMemNetwork.nodes[ node ].image = Running;
        }
        else if ( this.network.nodes[ node ].image == 'notrunning' ) {
          this.network.nodes[ node ].image = NotRunning;
          this.onMemNetwork.nodes[ node ].image = NotRunning;
        }
        else if ( this.network.nodes[ node ].image == 'notboot' ) {
          this.network.nodes[ node ].image = NotBoot;
          this.onMemNetwork.nodes[ node ].image = NotBoot;
        }
        else if ( this.network.nodes[ node ].image == 'notdeploy' ) {
          this.network.nodes[ node ].image = NotDeploy;
          this.onMemNetwork.nodes[ node ].image = NotDeploy;
        }
      }
    },

    /*
    Function: getInfo()
    Params: event click from vm
    Return: by reference vm info

    For each vm we get the overall information, host, ips, etc.
    After fetching the information we enable and populated the popup.
    */
    getInfo( e ) {
       for ( var vm = 0; vm < this.experiment.vms.length; vm++ ) {
        if ( this.experiment.vms[ vm ].id == e.nodes[ 0 ] ) {
          this.expModal.vm = this.experiment.vms[ vm ]
          this.expModal.active = true;
          break;
        }
      }
    },

    /*
    Start a VM in error state
    */
    startVm ( name ) {
      this.$buefy.dialog.confirm({
        title: 'Start the VM',
        message: 'This will start the ' + name + ' VM.',
        cancelText: 'Cancel',
        confirmText: 'Start',
        type: 'is-sucess',
        hasIcon: true,
        onConfirm: () => {
          this.isWaiting = true;
          this.$http.post(
            'experiments/' + this.$route.params.id + '/vms/' + name + '/start'
          ).then(
            response => {
              this.resetNetwork();
              this.isWaiting = false;
            }, response => {
              this.$beufy.toast.open({
                message: 'Start the ' + name + ' VM failed with ' + response.status + ' status.',
                type: 'is-danger',
                duration: 4000
              });
              this.isWaiting = false;
            }
          );
        }
      })

      this.expModal.active = false;
      this.resetExpModal();
    },

    /*
    Pause a running VM
    */
    pauseVm ( name ) {
      this.$buefy.dialog.confirm({
        title: 'Pause the VM',
        message: 'This will pause the ' + name + ' VM.',
        cancelText: 'Cancel',
        confirmText: 'Pause',
        type: 'is-waiting',
        hasIcon: true,
        onConfirm: () => {
          this.isWaiting = true;
          this.$http.post(
            'experiments/' + this.$route.params.id + '/vms/' + name + '/stop'
          ).then(
            response => {
              this.resetNetwork();
              this.isWaiting = false;
            }, response => {
              this.$buefy.toast.open({
                message: 'Pausing the ' + name + ' VM failed with ' + response.status + ' status.',
                type: 'is-danger',
                duration: 4000
              });
              this.isWaiting = false;
            }
          );
        }
      })
      
      this.expModal.active = false;
      this.resetExpModal();
    },

    /*
    Kill a running VM
    */
    killVm ( name ) {
      this.$buefy.dialog.confirm({
        title: 'Kill the VM',
        message: 'This will kill the '
        + name
        + ' VM. You will not be able to restore this VM until you restart the '
        + this.$route.params.id
        + ' experiment!',
        cancelText: 'Cancel',
        confirmText: 'KILL IT',
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          this.isWaiting = true;
          this.$http.delete(
            'experiments/' + this.$route.params.id + '/vms/' + name
          ).then(
            response => {
              this.resetNetwork();
              this.isWaiting = false;
          }, response => {
            this.$buefy.toast.open({
              message: 'Killing the ' + name + ' VM failed with ' + response.status + ' status.',
              type: 'is-danger',
              duration: 4000
            });
            this.isWaiting = false;
          }
         );
        }
      })
      
      this.expModal.active = false;
      this.resetExpModal();
    },

    /*
    Reset the popup information
    */
    resetExpModal () {
      this.expModal = {
        active: false,
        vm: [],
        snapshots: false
      }
    },

    /*
    Refresh topology view
    */
    resetNetwork () {
      this.radioButton = '';
      this.network = [];
      this.onMemNetwork = [];
      this.updateNetwork();
    },

    /*
    Function: filterNetwork
    Paramms: user's filter
    Description: filter the main network to only show requested VMs
    */
    filterNetwork ( filter ) {
      let nodes = [];
      
      1/*
      traverse the network and select VMs that match user's filter
      include all interfaces
      */
      this.onMemNetwork.nodes.forEach( function( node ) {
        if ( node.status == filter ) {
          nodes.push( node );
        }
        if ( node.status == "interface" ) {
          nodes.push( node );
        }
      });
      
      /*
      Reset nodes in the network and update with user's filter
      */
      this.network.nodes = [];
      this.network.nodes = nodes;
    }
  },
  
  watch: {
    radioButton: function ( val ) {
      if ( val != '' ) {
        this.filterNetwork( val )  
      }
    }
  },

  data () {
    return {
      experiment: [],
      network: [],
      onMemNetwork: [],
      isWaiting: true,
      expModal : {
        active: false,
        vm: [],
        snapshots: false
      },
      isShow: false,
      radioButton: '',
      optionsImg: Options,
        switchImg: Switch,
        runningImg: Running,
        nrunningImg: NotRunning,
        notdeployImg: NotDeploy,
        notbootImg: NotBoot
    }
  }
}
</script>

<style lang="css">
  .wrapper {
    padding: 20px 50px;
    text-align: center;
  }

  .modal-card-body {
    text-align: left;
  }

  .network {
    height: 800px;
    border: 1px solid #ccc;
    margin: 5px 0;
    background: black;
  }

  .options {
    text-align: left;
    color: white;
    position: absolute;
    z-index: 1;
    top: 18%;
  }

  .child {
    list-style-type: none;
    margin: 0px;
    padding: 0px;
  }

  .key {
    position: relative;
    z-index: 1;
    top: 86%;
    right: 5%;
  }
  .key > p {
    display: inline-block;
    color: white;
  }
  .key > p > img {
    width: 30px;
  }

  label.radio:hover {
    color: whitesmoke;
  }
</style>
