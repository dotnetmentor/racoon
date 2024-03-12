import './assets/main.css'

import { createApp } from 'vue'
import router from './router'
import App from './App.vue'

import VueDiff from 'vue-diff'
import 'vue-diff/dist/index.css'

import { OhVueIcon, addIcons } from 'oh-vue-icons'
import {
  FaShieldAlt,
  FaUnlockAlt,
  FaHome,
  FaCogs,
  FaArrowsAltH,
  FaGithub
} from 'oh-vue-icons/icons/fa'

addIcons(FaShieldAlt, FaUnlockAlt, FaHome, FaCogs, FaArrowsAltH, FaGithub)

const app = createApp(App)
app.component('v-icon', OhVueIcon)
app.use(router)
app.use(VueDiff)
app.mount('#app')
