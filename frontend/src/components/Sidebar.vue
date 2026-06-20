<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <div class="sidebar-logo">
        <i class="pi pi-check-square" style="font-size: 1.5rem; color: #3b82f6;"></i>
        <span class="sidebar-brand">ListMak</span>
      </div>
    </div>

    <nav class="sidebar-nav">
      <router-link 
        v-for="item in navItems" 
        :key="item.path"
        :to="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
      >
        <i :class="item.icon"></i>
        <span>{{ item.label }}</span>
      </router-link>
    </nav>

    <div class="sidebar-footer">
      <button @click="handleLogout" class="logout-btn">
        <i class="pi pi-sign-out"></i>
        <span>Logout</span>
      </button>
    </div>
  </aside>
</template>

<script>
import { logout } from '../api/auth'

export default {
  name: 'Sidebar',
  computed: {
    isAdmin() {
      return JSON.parse(localStorage.getItem('user') || '{}')?.role === 'admin'
    },
    navItems() {
      const items = [
        { path: '/today', icon: 'pi pi-home', label: 'Hari Ini' },
        { path: '/listmak/input', icon: 'pi pi-plus-circle', label: 'Input ListMak' },
        { path: '/listmak/daily', icon: 'pi pi-calendar', label: 'Riwayat' },
        { path: '/contacts', icon: 'pi pi-users', label: 'Kontak' },
        { path: '/profile', icon: 'pi pi-user', label: 'Profil' },
      ]
      if (this.isAdmin) {
        items.push({ path: '/admin/ai-logs', icon: 'pi pi-shield', label: 'Admin' })
      }
      return items
    }
  },
  methods: {
    isActive(path) {
      return this.$route.path === path
    },
    async handleLogout() {
      const result = await logout()
      if (result.success) {
        this.$router.push('/')
      }
    }
  }
}
</script>

<style scoped>
.sidebar {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  width: 240px;
  height: 100vh;
  background: #1e293b;
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  flex-direction: column;
  z-index: 100;
}

@media (min-width: 768px) {
  .sidebar {
    display: flex;
  }
}

.sidebar-header {
  height: 64px;
  padding: 0 1.25rem;
  display: flex;
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.sidebar-brand {
  font-size: 1.25rem;
  font-weight: 700;
  color: #f1f5f9;
}

.sidebar-nav {
  flex: 1;
  padding: 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  color: #94a3b8;
  text-decoration: none;
  border-radius: 0.5rem;
  transition: all 0.2s;
  font-size: 0.875rem;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #f1f5f9;
}

.nav-item.active {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.logout-btn {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: transparent;
  border: none;
  color: #94a3b8;
  border-radius: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}
</style>
