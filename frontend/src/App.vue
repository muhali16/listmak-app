<template>
  <div class="dark-mode">
    <OfflineModal />
    <Toast position="top-center" />
    
    <template v-if="showNavigation">
      <Sidebar />
      <main class="main-content">
        <router-view />
      </main>
      <BottomNav />
    </template>
    <template v-else>
      <router-view />
    </template>
  </div>
</template>

<script>
import Toast from 'primevue/toast'
import OfflineModal from './components/OfflineModal.vue'
import BottomNav from './components/BottomNav.vue'
import Sidebar from './components/Sidebar.vue'

export default {
  name: 'App',
  components: {
    Toast,
    OfflineModal,
    BottomNav,
    Sidebar
  },
  computed: {
    showNavigation() {
      return !this.$route.meta.hideNav
    }
  }
}
</script>

<style scoped>
.dark-mode {
  min-height: 100vh;
  min-height: 100dvh;
}

.main-content {
  min-height: 100vh;
  min-height: 100dvh;
  padding-bottom: 80px;
}

@media (min-width: 768px) {
  .main-content {
    margin-left: 240px;
    padding-bottom: 0;
  }
}
</style>
