import './assets/main.css'

import { createApp } from 'vue'
import router from './router'
import App from './App.vue'

import VueDiff from 'vue-diff'
import 'vue-diff/dist/index.css'

const app = createApp(App)
app.use(router)
app.use(VueDiff)
app.mount('#app')
