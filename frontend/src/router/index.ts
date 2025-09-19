import { createRouter, createWebHistory } from 'vue-router'
import Login from '@/pages/Login.vue'
import Register from '@/pages/Register.vue'
import Home from '@/pages/Home.vue'
import FindFriends from '@/pages/FindFriends.vue'
import FriendRequests from '@/pages/FriendRequests.vue'

const routes = [
  { path: '/', redirect: '/login' },
  { path: '/login', name: 'Login', component: Login },
  { path: '/register', name: 'Register', component: Register },

  { path: '/home', name: 'Home', component: Home, meta: { requiresAuth: true } },

  { path: '/:pathMatch(.*)*', redirect: '/login' },
  {
    path: '/find-friends', name: 'FindFriends', component: FindFriends, meta: { requiresAuth: true}},

  {path: '/friend-requests', name: 'FriendRequests', component: FriendRequests, meta: { requiresAuth: true }},
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('auth_token')

  if (to.meta.requiresAuth && !token) {
    next({ path: '/login', replace: true })
    return
  }
  if ((to.path === '/login' || to.path === '/register') && token) {
    next({ path: '/home', replace: true })
    return
  }
  next()
})