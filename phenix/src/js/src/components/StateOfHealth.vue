<!--
File: StateOfHealth.vue
This component will display and overview of the state of health
of each experiment. The component will display a list of all
experiments (running/not running), and an overview of the number
of vms running and not running. The user can select an experiment
to view in more details its state.
 -->

<template>
  <div class="content">
    <template v-if="experiments.length == 0">
      <section class="hero is-light is-bold is-large">
        <div class="hero-body">
          <div class="container" style="text-align: center">
            <h1 class="title">
              There are no experiments!
            </h1>
          </div>
        </div>
      </section>
    </template>
    <!-- If there are experiment display them in a table -->
    <template v-else>
      <hr>
      <b-field position="is-right">
        <b-autocomplete v-model="searchName"
                        placeholder="Find an Experiment"
                        icon="search"
                        :data="filteredData"
                        @select="option => filtered = option">
          <template slot="empty">
            No results found
          </template>
        </b-autocomplete>
        <p class='control'>
          <button class='button' style="color:#686868" @click="searchName = ''">
            <b-icon icon="window-close"></b-icon>
          </button>
        </p>
      </b-field>
      <div>
        <b-table
          :data="filteredExperiments"
          :paginated="table.isPaginated && paginationNeeded"
          :per-page="table.perPage"
          :current-page.sync="table.currentPage"
          :pagination-simple="table.isPaginationSimple"
          :pagination-size="table.paginationSize"
          :default-sort-direction="table.defaultSortDirection"
          default-sort="name">
          <template slot="empty">
            <section class="section">
              <div class="content has-text-white has-text-centered">
                Your search turned up empty!
              </div>
            </section>
          </template>
          <template slot-scope="props">
            <b-table-column field="name" label="Name" width="200" sortable>
              <template v-if="updating( props.row.status )">
                {{ props.row.name }}
              </template>
              <template v-else>
                <router-link class="navbar-item" :to="{ name: 'soh', params: { id: props.row.name }}">
                  {{ props.row.name }}
                </router-link>
              </template>
            </b-table-column>
            <b-table-column field="status" label="Status" width="100" sortable centered>
              <template v-if="props.row.status == 'starting'">
                <section>
                  <b-progress size="is-medium" type="is-warning" show-value :value=props.row.percent format="percent"></b-progress>
                </section>
              </template>
              <template v-else-if="adminUser()">
                <span class="tag is-medium" :class="decorator( props.row.status )">
                  <div class="field">
                    <div class="field" @click="( props.row.running ) ? stop( props.row.name, props.row.status ) : start( props.row.name, props.row.status )">
                      {{ props.row.status }}
                    </div>
                  </div>
                </span>
              </template>
              <template v-else>
                <span class="tag is-medium" :class="decorator( props.row.status )">
                  {{ props.row.status }}
                </span>
              </template>
            </b-table-column>
            <b-table-column field="topology" label="Topology" width="200">
              {{ props.row.topology | lowercase }}
            </b-table-column>
            <b-table-column field="apps" label="Applications" width="200">
              {{ props.row.apps | stringify | lowercase }}
            </b-table-column>
            <b-table-column field="start_time" label="Start Time" width="250" sortable>
              {{ props.row.start_time }}
            </b-table-column>
            <b-table-column field="vm_count" label="# of Running VMs" width="100" centered>
              {{ props.row.running_count }} / {{ props.row.total_count }}
            </b-table-column>
            <b-table-column field="vm_count" label="# of Not Running VMs" width="100" centered>
              {{ props.row.notrunning }}
            </b-table-column>
            <b-table-column field="vm_count" label="# of Not Deployed VMs" width="100" centered>
              {{ props.row.notdeploy }}
            </b-table-column>
            <b-table-column field="vm_count" label="# of Not Booted VMs" width="100" centered>
              {{ props.row.notboot }}
            </b-table-column>
          </template>
        </b-table>
        <br>
        <b-field v-if="paginationNeeded" grouped position="is-right">
          <div class="control is-flex">
            <b-switch v-model="table.isPaginated" size="is-small" type="is-light">Pagenate</b-switch>
          </div>
        </b-field>
      </div>
    </template>
    <b-loading :is-full-page="true" :active.sync="isWaiting" :can-cancel="false"></b-loading>
  </div>
</template>

<script>
  export default {
    async beforeDestroy () {
      this.$options.sockets.onmessage = null;
    },

    async created () {
      this.updateExperiments();
    },

    computed: {
      /*
      Sort and parse the experiment(s) information to display them in the table.
      */
      filteredExperiments: function() {
        let experiments = this.experiments;
        
        var name_re = new RegExp( this.searchName, 'i' );
        var data = [];
        
        for ( let i in experiments ) {
          let exp = experiments[i];
          let running_vms = 0;

          if (this.soh_ready) {
            /*
            Get an overview of soh of each experiment
            */
            exp.running_count = this.soh[i][1];
            exp.total_count = this.soh[i][2];
            exp.notboot = this.soh[i][3];
            exp.notdeploy = this.soh[i][4];
            exp.notrunning = this.soh[i][5];
          }

          if ( exp.name.match( name_re ) ) {
            exp.start_time = exp.start_time == '' ? 'N/A' : exp.start_time;
            data.push( exp );
          }
        }

        return data;
      },
    
      filteredData () {
        let names = this.experiments.map( exp => { return exp.name; } );

        return names.filter(
          option => {
            return option
              .toString()
              .toLowerCase()
              .indexOf( this.searchName.toLowerCase() ) >= 0
          }
        )
      },

      /*
      Enable pagination at the table
      */
      paginationNeeded () {
        var experiments = this.experiments;
        if ( experiments.length <= 10 ) {
          return false;
        } else {
          return true;
        }
      }
    },
    
    methods: { 
      /*
      Get all experiments
      */
      async updateExperiments () {
        try {
          let resp = await this.$http.get('experiments');
          let state = await resp.json();
          this.experiments = state.experiments;

          for (let experiment in this.experiments) {
            let exp_soh = [];
            let soh_tmp = [];
            console.log('Getting soh for: ' + this.experiments[experiment].name + '/soh');
            resp = await this.$http.get('experiments/' + this.experiments[experiment].name + '/soh');
            state = await resp.json();
            soh_tmp = state;
            exp_soh.push(this.experiments[experiment].name);
            exp_soh.push(soh_tmp.running_count);
            exp_soh.push(soh_tmp.total_count);
            exp_soh.push(soh_tmp.notboot_count);
            exp_soh.push(soh_tmp.notdeploy_count);
            exp_soh.push(soh_tmp.notrunning_count);
            this.soh.push(exp_soh);
          }
          this.soh_ready = true;

          this.isWaiting = false;
        } catch {
          this.$buefy.toast.open ({
            message: 'Getting the Experiments Failed',
            type: 'is-danger',
            duration: 40000
          });
        } finally {
          this.isWaiting = false;
        }
      },
      
      /*
      Base on the role of the current user, enable options.  
      */
      globalUser () {
        return [ 'Global Admin' ].includes( this.$store.getters.role );
      },
      
      adminUser () {
        return [ 'Global Admin', 'Experiment Admin' ].includes( this.$store.getters.role );
      },
      
      experimentUser () {
        return [ 'Global Admin', 'Experiment Admin', 'Experiment User' ].includes( this.$store.getters.role );
      },

      update: function ( value ) {
        this.isMenuActive = true;
      },

      updating: function( status ) {
        return status === "starting" || status === "stopping";
      },
      
      decorator ( status ) {
        switch ( status ) {
          case 'started':
            return 'is-success';
          case 'starting':
          case 'stopping':
            return 'is-warning';
          case 'stopped':
            return 'is-danger';
        }
      }
    },
    
    directives: {
      focus: {
        inserted ( el ) {
          if ( el.tagName == 'INPUT' ) {
            el.focus()
          } else {
            el.querySelector( 'input' ).focus()
          }
        }
      }
    },

    data () {
      return {
        table: {
          isPaginated: true,
          perPage: 10,
          currentPage: 1,
          isPaginationSimple: true,
          paginationSize: 'is-small',
          defaultSortDirection: 'asc'
        },
        soh: [],
        soh_ready: false,
        experiments: [],
        topologies: [],
        applications: [],
        searchName: '',
        filtered: null,
        isMenuActive: false,
        action: null,
        rowName: null,
        isWaiting: true
      }
    }
  }
</script>

<style scoped>
  div.autocomplete >>> a.dropdown-item {
    color: #383838 !important;
  }
</style>
