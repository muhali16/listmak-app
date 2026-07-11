<template>
  <div class="testing-banner" role="alert">
    <i class="pi pi-exclamation-triangle"></i>
    <span>Aplikasi dalam tahap <strong>TESTING</strong> — data dapat berubah atau dihapus sewaktu-waktu.</span>
  </div>
</template>

<script>
export default {
  name: 'TestingBanner'
}
</script>

<style scoped>
.testing-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 3000;
  height: var(--testing-banner-h, 2.25rem);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0 0.75rem;
  background: #facc15;
  color: #422006;
  font-size: 0.8125rem;
  font-weight: 600;
  text-align: center;
  line-height: 1.2;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.25);
}

.testing-banner i {
  font-size: 0.9375rem;
  flex-shrink: 0;
}

.testing-banner strong {
  font-weight: 800;
  letter-spacing: 0.03em;
}
</style>
