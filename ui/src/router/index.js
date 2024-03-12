import { createRouter, createWebHashHistory } from 'vue-router'

import Start from '@/views/Start.vue'
import Configs from '@/views/Configs.vue'
import Compare from '@/views/Compare.vue'

const routes = [
  { path: '/', name: 'Start', component: Start },
  { path: '/configs', name: 'Configurations', component: Configs },
  { path: '/compare', name: 'Compare', component: Compare }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
