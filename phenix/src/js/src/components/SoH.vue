<!--
File: SoH.vue
This component will pass the state of health view for each
running experiment. It will include a specific component 
rendering based on whether the experiment is running or not
 -->


<template>
  <component :is="component"></component>
</template>

<script>
  import Vue from 'vue'

  export default {
    /*
    These are the components available to the main experiment 
    component based on whether or not an experiment is running.
    */
    components: {
      running: () => import( './RunningVms.vue' ),
      notrunning: () => import( './NotRunningVms.vue' )
    },

    async beforeRouteEnter ( to, _, next ) {
      try {
        let resp = await Vue.http.get( 'experiments/' + to.params.id );
        let state = await resp.json();

        next( vm => vm.running = state.running );
      } catch ( err ) {
        console.log( err );

        Vue.toast.open({
          message: 'Getting the ' + to.params.id + ' experiment failed.',
          type: 'is-danger',
          duration: 4000
        });

        next();
      }
    },

    /*  
    This computed value is based on the routing parameter 
    determined by the user clicking into an experiment from the 
    experiments table. The result is to pass the State of Health Dashboard
    if the experiment is running.
    */
    computed: {
      component: function () {
        if ( this.running == null ) {
          return
        }

        if ( this.running == true ) {
          return 'running';
        }

        return 'notrunning';
      }
    },
    
    data () {
      return {
        running: null
      }
    }
  }
</script>
