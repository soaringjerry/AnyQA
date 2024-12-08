import { createRouter, createWebHashHistory } from 'vue-router'
import IndexPage from '../pages/IndexPage.vue'
import PresenterPage from '../pages/PresenterPage.vue'
import DisplayPage from '../pages/DisplayPage.vue'
import SessionSetup from '../pages/SessionSetup.vue'
import 'animate.css'

const routes = [
  { path: '/', component: IndexPage },
  { path: '/presenter', component: PresenterPage },
  { path: '/display', component: DisplayPage },
  { path: '/setup', component: SessionSetup }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
