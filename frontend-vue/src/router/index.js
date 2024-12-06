import { createRouter, createWebHistory } from 'vue-router'
import IndexPage from '../pages/IndexPage.vue'
import PresenterPage from '../pages/PresenterPage.vue'
import DisplayPage from '../pages/DisplayPage.vue'
import 'animate.css'

const routes = [
  { path: '/', component: IndexPage },
  { path: '/presenter', component: PresenterPage },
  { path: '/display', component: DisplayPage }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
