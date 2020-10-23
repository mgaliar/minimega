import Vue              from 'vue'
import Router           from 'vue-router'
import Experiments      from './components/Experiments.vue'
import Experiment       from './components/Experiment.vue'
import Hosts            from './components/Hosts.vue'
import SignIn           from './components/SignIn.vue'
import Users            from './components/Users.vue'
import Log              from './components/Log.vue'
import VMtiles          from './components/VMtiles.vue'
import StateOfHealth    from './components/StateOfHealth.vue'
import SoH              from './components/SoH.vue'

import NetworkExample from './components/NetworkExample.vue'

import store            from './store'

Vue.use(Router)

const router = new Router({
  mode: 'history',
  base: process.env.BASE_URL,
  routes: [
    { path: '/',                  name: 'home',             redirect: '/experiments' },
    { path: '/experiments',       name: 'experiments',      component: Experiments },
    { path: '/experiment/:id',    name: 'experiment',       component: Experiment },
    { path: '/vmtiles',           name: 'vmtiles',          component: VMtiles },
    { path: '/hosts',             name: 'hosts',            component: Hosts },
    { path: '/users',             name: 'users',            component: Users },
    { path: '/log',               name: 'log',              component: Log },
    { path: '/signin',            name: 'signin',           component: SignIn },
    // { path: '/stateofhealth',     name: 'stateofhealth',    component: StateOfHealth },
    { path: '/stateofhealth/:id', name: 'soh',              component: SoH },
    { path: '/example',           name: 'networkExample',   component: NetworkExample },
    { path: '*',                  redirect: '/signin' }
  ]
})

router.beforeEach( ( to, from, next ) => {
  if ( process.env.VUE_APP_AUTH === 'disabled' ) {
    if ( !store.getters.auth ) {
      store.commit( 'LOGIN', { 'user': { 'role': 'Global Admin' }, 'remember': false } )
    }

    next()
    return
  }

  let pub = ['/signup', '/signin']

  if ( pub.includes( to.path ) ) {
    next()
    return
  }

  if ( store.state.auth ) {
    next()
  } else {
    next( '/signin' )
  }
})

export default router
