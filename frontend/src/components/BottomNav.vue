<template>
  <nav class="bottom-nav">
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
</template>

<script>
export default {
  name: 'BottomNav',
  data() {
    return {
      navItems: [
        { path: '/today', icon: 'pi pi-home', label: 'Hari Ini' },
        { path: '/listmak/input', icon: 'pi pi-plus-circle', label: 'Input' },
        { path: '/listmak/daily', icon: 'pi pi-calendar', label: 'Riwayat' },
        { path: '/contacts', icon: 'pi pi-users', label: 'Kontak' },
        { path: '/profile', icon: 'pi pi-user', label: 'Profil' }
      ]
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

@media (min-width: 768px) {
  .bottom-nav {
    display: none;
  }
}
</style>
