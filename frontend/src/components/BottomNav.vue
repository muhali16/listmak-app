<template>
  <nav class="bottom-nav">
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

    <!-- Admin trigger -->
    <button
      v-if="isAdmin"
      class="nav-item"
      :class="{ active: isAdminRoute }"
      @click.stop="adminMenuOpen = !adminMenuOpen"
    >
      <i class="pi pi-shield"></i>
      <span>Admin</span>
    </button>

    <!-- Admin dropdown popup -->
    <transition name="admin-popup">
      <div v-if="adminMenuOpen" class="admin-popup">
        <router-link
          to="/admin/listmaks"
          class="admin-popup-item"
          :class="{ active: $route.path === '/admin/listmaks' }"
          @click="adminMenuOpen = false"
        >
          <i class="pi pi-list"></i>
          <span>Kelola Listmak</span>
        </router-link>
        <div class="admin-popup-divider"></div>
        <router-link
          to="/admin/price-catalog"
          class="admin-popup-item"
          :class="{ active: $route.path === '/admin/price-catalog' }"
          @click="adminMenuOpen = false"
        >
          <i class="pi pi-tag"></i>
          <span>Price Catalog</span>
        </router-link>
        <div class="admin-popup-divider"></div>
        <router-link
          to="/admin/ai-logs"
          class="admin-popup-item"
          :class="{ active: $route.path === '/admin/ai-logs' }"
          @click="adminMenuOpen = false"
        >
          <i class="pi pi-microchip-ai"></i>
          <span>AI Logs</span>
        </router-link>
        <div class="admin-popup-divider"></div>
        <router-link
          to="/admin/system-logs"
          class="admin-popup-item"
          :class="{ active: $route.path === '/admin/system-logs' }"
          @click="adminMenuOpen = false"
        >
          <i class="pi pi-server"></i>
          <span>System Logs</span>
        </router-link>
      </div>
    </transition>

    <!-- backdrop to close -->
    <div v-if="adminMenuOpen" class="admin-popup-backdrop" @click="adminMenuOpen = false"></div>
  </nav>
</template>

<script>
export default {
  name: 'BottomNav',
  data() {
    return {
      adminMenuOpen: false
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
    navItems() {
      return [
        { path: '/today', icon: 'pi pi-home', label: 'Hari Ini' },
        { path: '/listmak/daily', icon: 'pi pi-calendar', label: 'Riwayat' },
        { path: '/profile', icon: 'pi pi-user', label: 'Profil' },
      ]
    },
    isAdminRoute() {
      return this.$route.path.startsWith('/admin')
    }
  },
  watch: {
    $route() {
      this.adminMenuOpen = false
    }
  },
  methods: {
    isActive(path) {
      return this.$route.path === path
    }
  }
}
</script>

<style scoped>
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 64px;
  background: rgba(15, 23, 42, 0.98);
  backdrop-filter: blur(10px);
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding: 0 0.5rem;
  z-index: 100;
  padding-bottom: env(safe-area-inset-bottom);
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 0.5rem 0.75rem;
  color: #64748b;
  text-decoration: none;
  border-radius: 0.75rem;
  transition: all 0.2s;
  min-width: 56px;
  background: transparent;
  border: none;
  cursor: pointer;
  font-family: inherit;
}

.nav-item i {
  font-size: 1.25rem;
}

.nav-item span {
  font-size: 0.625rem;
  font-weight: 500;
}

.nav-item.active {
  color: #3b82f6;
}

.nav-item.active i {
  transform: scale(1.1);
}

/* Admin popup */
.admin-popup-backdrop {
  position: fixed;
  inset: 0;
  z-index: 98;
}

.admin-popup {
  position: absolute;
  bottom: calc(100% + 0.5rem);
  right: 0.75rem;
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.875rem;
  overflow: hidden;
  z-index: 99;
  min-width: 180px;
  box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.4);
}

.admin-popup-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1.125rem;
  color: #94a3b8;
  text-decoration: none;
  font-size: 0.9rem;
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
}

.admin-popup-item i {
  font-size: 1rem;
  width: 1.25rem;
  text-align: center;
}

.admin-popup-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #f1f5f9;
}

.admin-popup-item.active {
  color: #818cf8;
  background: rgba(99, 102, 241, 0.1);
}

.admin-popup-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.06);
  margin: 0;
}

/* Transition */
.admin-popup-enter-active,
.admin-popup-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.admin-popup-enter-from,
.admin-popup-leave-to {
  opacity: 0;
  transform: translateY(6px);
}

@media (min-width: 768px) {
  .bottom-nav {
    display: none;
  }
}

.nav-avatar-mini {
  width: 28px;
  height: 28px;
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
