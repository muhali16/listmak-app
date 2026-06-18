<template>
  <teleport to="body">
    <div v-if="isOffline" class="offline-overlay">
      <div class="offline-modal">
        <div class="offline-icon">
          <i class="pi pi-wifi" style="font-size: 3rem; color: #ef4444;"></i>
        </div>
        <h2 class="offline-title">Tidak Ada Koneksi</h2>
        <p class="offline-text">
          Periksa koneksi internet Anda dan coba lagi
        </p>
        <div class="offline-spinner">
          <i class="pi pi-spin pi-spinner" style="font-size: 1.5rem;"></i>
        </div>
        <p class="offline-hint">Mencari koneksi...</p>
      </div>
    </div>
  </teleport>
</template>

<script>
export default {
  name: 'OfflineModal',
  data() {
    return {
      isOffline: !navigator.onLine
    }
  },
  mounted() {
    window.addEventListener('online', this.handleOnline)
    window.addEventListener('offline', this.handleOffline)
  },
  beforeUnmount() {
    window.removeEventListener('online', this.handleOnline)
    window.removeEventListener('offline', this.handleOffline)
  },
  methods: {
    handleOnline() {
      this.isOffline = false
    },
    handleOffline() {
      this.isOffline = true
    }
  }
}
</script>

<style scoped>
.offline-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 1rem;
}

.offline-modal {
  background: linear-gradient(145deg, #1e293b, #0f172a);
  border-radius: 1rem;
  padding: 2rem;
  text-align: center;
  max-width: 320px;
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
}

.offline-icon {
  margin-bottom: 1rem;
}

.offline-title {
  color: #f1f5f9;
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.offline-text {
  color: #94a3b8;
  font-size: 0.875rem;
  margin-bottom: 1.5rem;
}

.offline-spinner {
  color: #60a5fa;
  margin-bottom: 0.5rem;
}

.offline-hint {
  color: #64748b;
  font-size: 0.75rem;
}
</style>
