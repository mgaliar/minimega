import Vue        from 'vue'
import Vuex       from 'vuex'
import router     from './router'

Vue.use( Vuex )

export default new Vuex.Store({
  state: {
    username: null,
    token:    null,
    role:     null,
    auth:     null
  },

  mutations: {
    'LOGIN' ( state, { user, remember } ) {
      state.username = user.username;
      state.token    = user.token;
      state.role     = user.role;
      state.auth     = true;

      if ( remember ) {
        localStorage.setItem( 'user',  state.username );
        localStorage.setItem( 'token', state.token );
        localStorage.setItem( 'role',  state.role );
        localStorage.setItem( 'auth',  state.auth );
      }

      if ( state.role === "VM Viewer" ) {
        router.replace( '/vmtiles' );
      } else {
        router.replace( '/experiments' );
      }
    },

    'LOGOUT' ( state ) {
      state.username = null;
      state.token    = null;
      state.role     = null;
      state.auth     = false;
      state.logs     = [];

      if ( localStorage.getItem( 'auth' ) ) {
        localStorage.removeItem( 'user' );
        localStorage.removeItem( 'token' );
        localStorage.removeItem( 'role' );
        localStorage.setItem( 'auth', state.auth );
      }
      router.replace( '/signin' );
    }
  },
  
  getters: {
    username: state => {
      return state.username;
    },

    token: state => {
      return state.token;
    },
    
    role: state => {
      return state.role;
    },
    
    auth: state => {
      return state.auth;
    }
  }
});
