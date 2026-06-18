<template>
  <!--
    PLACEHOLDER ONLY — not the real Order list view (BRIEF step 5b).
    The full grouped-by-name view with the per-person "Lunas" bulk toggle is
    gated until task #1 (bulk-update is_paid) is verified end-to-end.
    This stub just prevents the Today view "Buka" button from dead-linking.
  -->
  <div class="placeholder-container">
    <button class="back-btn" @click="goBack">
      <i class="pi pi-arrow-left"></i>
      <span>Kembali</span>
    </button>

    <div class="placeholder-body">
      <div class="placeholder-icon">
        <i class="pi pi-clock"></i>
      </div>
      <h2>Halaman daftar pesanan sedang disiapkan</h2>
      <p>Listmak #{{ id }} — tampilan daftar pesanan akan tersedia di tahap berikutnya.</p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'OrderListView',
  computed: {
    id() {
      return this.$route.params.id
    }
  },
  methods: {
    goBack() {
      this.$router.push('/today')
    }
  }
}
</script>

<style scoped>
.placeholder-container {
  padding: 1rem;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.75rem;
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
}

.placeholder-body {
  text-align: center;
  padding: 3rem 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.placeholder-icon {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: rgba(30, 41, 59, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
}

.placeholder-icon i {
  font-size: 1.75rem;
  color: #475569;
}

.placeholder-body h2 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #f1f5f9;
}

.placeholder-body p {
  font-size: 0.875rem;
  color: #64748b;
}
</style>
