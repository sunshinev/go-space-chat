import Vue from 'vue'
import App from './App.vue'
import ViewUI from 'view-design';
import 'view-design/dist/styles/iview.css';

Vue.use(ViewUI);

import axios from 'axios'
import VueAxios from 'vue-axios'

Vue.use(VueAxios,axios);

Vue.config.productionTip = false

new Vue({
  render: h => h(App),
}).$mount('#app')
