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
        <template v-if="item.path === '/profile'">
          <div class="nav-avatar-mini" :class="{ 'nav-avatar-mini--admin': isAdmin }">
            <img v-if="currentUser.avatar" :src="currentUser.avatar" :alt="currentUser.name" referrerpolicy="no-referrer" />
            <span v-else>{{ userInitials }}</span>
          </div>
        </template>
        <i v-else :class="item.icon"></i>
        <span>{{ item.label }}</span>
      </router-link>

      <!-- Admin submenu -->
      <template v-if="isAdmin">
        <button
          class="nav-item nav-item--group"
          :class="{ active: isAdminRoute }"
          @click="adminOpen = !adminOpen"
        >
          <i class="pi pi-shield"></i>
          <span>Admin</span>
          <i class="pi nav-chevron" :class="adminOpen ? 'pi-chevron-up' : 'pi-chevron-down'"></i>
        </button>
        <div v-if="adminOpen" class="nav-subitems">
          <router-link to="/admin/listmaks" class="nav-subitem" :class="{ active: $route.path === '/admin/listmaks' }">
            <i class="pi pi-list"></i>
            <span>Kelola Listmak</span>
          </router-link>
          <router-link to="/admin/price-catalog" class="nav-subitem" :class="{ active: $route.path === '/admin/price-catalog' }">
            <i class="pi pi-tag"></i>
            <span>Price Catalog</span>
          </router-link>
          <router-link to="/admin/ai-logs" class="nav-subitem" :class="{ active: $route.path === '/admin/ai-logs' }">
            <i class="pi pi-microchip-ai"></i>
            <span>AI Logs</span>
          </router-link>
          <router-link to="/admin/system-logs" class="nav-subitem" :class="{ active: $route.path === '/admin/system-logs' }">
            <i class="pi pi-server"></i>
            <span>System Logs</span>
          </router-link>
        </div>
      </template>
    </nav>

    <div class="sidebar-footer">
      <router-link to="/changelog" class="changelog-link" :class="{ active: $route.path === '/changelog' }">
        <i class="pi pi-megaphone"></i>
        <span>Pembaruan</span>
        <span v-if="hasUpdate" class="update-badge"></span>
      </router-link>
      <button @click="handleLogout" class="logout-btn">
        <i class="pi pi-sign-out"></i>
        <span>Logout</span>
      </button>
      <div class="version-info">v{{ currentVersion }}</div>
    </div>
  </aside>
</template>

<script>
import { logout } from '../api/auth'
import changelog from '../data/changelog.json'

export default {
  name: 'Sidebar',
  data() {
    return {
      adminOpen: false
    }
  },
  computed: {
    currentUser() {
      return JSON.parse(localStorage.getItem('user') || '{}')
    },
    isAdmin() {
      return this.currentUser?.role === 'admin'
    },
    userInitials() {
      const name = this.currentUser?.name || ''
      return name.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2) || 'U'
    },
    isAdminRoute() {
      return this.$route.path.startsWith('/admin')
    },
    hasUpdate() {
      return localStorage.getItem('lastSeenVersion') !== changelog[0]?.version
    },
    currentVersion() {
      return changelog[0]?.version ?? '-'
    },
    navItems() {
      return [
        { path: '/today', icon: 'pi pi-home', label: 'Hari Ini' },
        { path: '/listmak/daily', icon: 'pi pi-calendar', label: 'Riwayat' },
        { path: '/profile', icon: 'pi pi-user', label: 'Profil' },
      ]
    }
  },
  watch: {
    isAdminRoute: {
      immediate: true,
      handler(val) {
        if (val) this.adminOpen = true
      }
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
  width: 100%;
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #f1f5f9;
}

.nav-item.active {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.nav-chevron {
  margin-left: auto;
  font-size: 0.75rem;
  opacity: 0.6;
}

.nav-subitems {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  padding-left: 1rem;
}

.nav-subitem {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.625rem 0.875rem;
  color: #64748b;
  text-decoration: none;
  border-radius: 0.5rem;
  font-size: 0.8125rem;
  transition: all 0.15s;
  border-left: 2px solid rgba(255, 255, 255, 0.06);
}

.nav-subitem:hover {
  color: #94a3b8;
  background: rgba(255, 255, 255, 0.04);
}

.nav-subitem.active {
  color: #818cf8;
  border-left-color: #818cf8;
  background: rgba(99, 102, 241, 0.08);
}

.nav-subitem i {
  font-size: 0.875rem;
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.changelog-link {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  color: #94a3b8;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  text-decoration: none;
  transition: all 0.2s;
  margin-bottom: 0.25rem;
}

.changelog-link:hover,
.changelog-link.active {
  background: rgba(255, 255, 255, 0.05);
  color: #e2e8f0;
}

.update-badge {
  margin-left: auto;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ef4444;
  flex-shrink: 0;
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

.version-info {
  margin-top: 0.5rem;
  padding: 0 1rem;
  font-size: 0.7rem;
  color: #475569;
  text-align: center;
}

.nav-avatar-mini {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid #475569;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.625rem;
  font-weight: 700;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
}

.nav-avatar-mini img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.nav-avatar-mini--admin {
  border-color: #f59e0b;
  background: linear-gradient(135deg, #f59e0b, #b45309);
}

.nav-item.active .nav-avatar-mini {
  border-color: #3b82f6;
}

.nav-item.active .nav-avatar-mini--admin {
  border-color: #f59e0b;
}
</style>
