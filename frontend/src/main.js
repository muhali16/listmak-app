import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { VueQueryPlugin, QueryClient } from '@tanstack/vue-query'
import PrimeVue from 'primevue/config'
import Aura from '@primevue/themes/aura'
import ToastService from 'primevue/toastservice'
import 'primeicons/primeicons.css'
import './style.css'

import App from './App.vue'
import { verifyAuth, isAuthenticated } from './api/auth'

// Views
import LoginView from './views/LoginView.vue'
import DashboardView from './views/DashboardView.vue'
import TodayView from './views/TodayView.vue'
import OrderListView from './views/OrderListView.vue'
import ListMakInputView from './views/ListMakInputView.vue'
import ListMakDailyView from './views/ListMakDailyView.vue'
import ContactsView from './views/ContactsView.vue'
import ProfileView from './views/ProfileView.vue'
import SharedListMakView from './views/SharedListMakView.vue'
import ViewListMakView from './views/ViewListMakView.vue'
import AdminAILogsView from './views/AdminAILogsView.vue'

// Router configuration
const routes = [
  { 
    path: '/', 
    name: 'Login', 
    component: LoginView, 
    meta: { hideNav: true, guest: true } 
  },
  {
    path: '/today',
    name: 'Today',
    component: TodayView,
    meta: { requiresAuth: true }
  },
  {
    path: '/listmak/:id(\\d+)',
    name: 'OrderList',
    component: OrderListView,
    meta: { requiresAuth: true }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: DashboardView,
    meta: { requiresAuth: true }
  },
  { 
    path: '/listmak/input', 
    name: 'ListMakInput', 
    component: ListMakInputView,
    meta: { requiresAuth: true }
  },
  { 
    path: '/listmak/daily', 
    name: 'ListMakDaily', 
    component: ListMakDailyView,
    meta: { requiresAuth: true }
  },
  { 
    path: '/contacts', 
    name: 'Contacts', 
    component: ContactsView,
    meta: { requiresAuth: true }
  },
  { 
    path: '/profile', 
    name: 'Profile', 
    component: ProfileView,
    meta: { requiresAuth: true }
  },
  { 
    path: '/listmak/order/:shareId', 
    name: 'SharedListMak', 
    component: SharedListMakView,
    meta: { hideNav: true, public: true }
  },
  {
    path: '/listmak/view/:viewId',
    name: 'ViewListMak',
    component: ViewListMakView,
    meta: { hideNav: true, public: true }
  },
  {
    path: '/admin/ai-logs',
    name: 'AdminAILogs',
    component: AdminAILogsView,
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Track if initial auth check is done
let authChecked = false

// Navigation guards
router.beforeEach(async (to, from, next) => {
  // On first navigation, verify auth with API
  if (!authChecked) {
    authChecked = true
    await verifyAuth()
  }
  
  // Allow public routes without authentication
  if (to.meta.public) {
    next()
    return
  }
  
  const authenticated = isAuthenticated()
  
  // If route requires auth and user is not authenticated
  if (to.meta.requiresAuth && !authenticated) {
    next({ name: 'Login' })
  }
  // If route is for guests only (login) and user is authenticated
  else if (to.meta.guest && authenticated) {
    next({ name: 'Today' })
  }
  // Otherwise proceed
  else {
    next()
  }
})

// TanStack Query client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      retry: 1
    }
  }
})

// Create app
const app = createApp(App)

// Use plugins
app.use(router)
app.use(VueQueryPlugin, { queryClient })
app.use(ToastService)
app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      darkModeSelector: '.dark-mode',
      cssLayer: false
    }
  }
})

app.mount('#app')
