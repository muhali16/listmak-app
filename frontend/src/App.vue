<template>
  <div class="dark-mode" :class="{ 'testing-on': testing }">
    <TestingBanner v-if="testing" />
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
import TestingBanner from './components/TestingBanner.vue'

export default {
  name: 'App',
  components: {
    Toast,
    OfflineModal,
    BottomNav,
    Sidebar,
    TestingBanner
  },
  data() {
    return {
      testing: import.meta.env.VITE_TESTING_MODE === 'true'
    }
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

/* Testing mode: reserve space for the fixed banner. The CSS var inherits into
   child components (Sidebar reads it to offset its fixed top). */
.dark-mode.testing-on {
  --testing-banner-h: 2.25rem;
  padding-top: 2.25rem;
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
