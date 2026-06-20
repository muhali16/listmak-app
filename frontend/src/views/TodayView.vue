<template>
  <div class="today-container">
    <!-- Header -->
    <header class="page-header">
      <h1 class="page-title">Hari Ini</h1>
      <p class="page-date">{{ formattedToday }}</p>
    </header>

    <!-- Create new listmak — always visible, explicit text label (not just "+") -->
    <button class="create-btn" @click="showCreateModal = true">
      <i class="pi pi-plus"></i>
      <span>Buat listmak baru</span>
    </button>

    <!-- Loading -->
    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat listmak hari ini...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="state-block state-error">
      <i class="pi pi-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <button class="retry-btn" @click="loadToday">Coba lagi</button>
    </div>

    <!-- Listmak cards (vertical list) -->
    <section v-else-if="listmaks.length > 0" class="cards">
      <article v-for="lm in listmaks" :key="lm.id" class="listmak-card">
        <div class="card-main">
          <h2 class="card-title">{{ lm.title || 'Tanpa judul' }}</h2>
          <p class="card-count">{{ lm.totalOrders }} pesanan</p>

          <div class="status-row">
            <span class="status-chip status-paid">
              <i class="pi pi-check-circle"></i>
              {{ lm.paidCount }} lunas
            </span>
            <span class="status-chip status-unpaid">
              <i class="pi pi-clock"></i>
              {{ lm.unpaidCount }} belum bayar
            </span>
          </div>
        </div>

        <button class="open-btn" @click="openListmak(lm.id)">
          <span>Buka</span>
          <i class="pi pi-arrow-right"></i>
        </button>
      </article>
    </section>

    <!-- Empty -->
    <section v-else class="state-block empty">
      <div class="empty-icon">
        <i class="pi pi-inbox"></i>
      </div>
      <h3>Belum ada listmak hari ini</h3>
      <p>Buat listmak baru untuk mulai mencatat pesanan.</p>
    </section>

    <!-- Create Listmak Modal -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="closeCreateModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Buat ListMak Baru</h3>
          <button @click="closeCreateModal" class="modal-close">
            <i class="pi pi-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Judul ListMak</label>
            <input
              ref="titleInput"
              type="text"
              v-model="newTitle"
              placeholder="Contoh: Makan Siang Ops"
              class="modal-input"
              @keyup.enter="createListmak"
            />
          </div>
          <p v-if="createError" class="create-error">{{ createError }}</p>
        </div>
        <div class="modal-footer">
          <button
            class="submit-btn"
            @click="createListmak"
            :disabled="!newTitle.trim() || creating"
          >
            <i v-if="creating" class="pi pi-spin pi-spinner"></i>
            <i v-else class="pi pi-check"></i>
            {{ creating ? 'Membuat...' : 'Buat ListMak' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { listmak } from '../api'

export default {
  name: 'TodayView',
  data() {
    return {
      listmaks: [],
      loading: false,
      error: '',
      showCreateModal: false,
      newTitle: '',
      creating: false,
      createError: ''
    }
  },
  computed: {
    today() {
      return new Date().toISOString().split('T')[0]
    },
    formattedToday() {
      const options = { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }
      return new Date().toLocaleDateString('id-ID', options)
    }
  },
  watch: {
    showCreateModal(val) {
      if (val) this.$nextTick(() => this.$refs.titleInput?.focus())
    }
  },
  mounted() {
    this.loadToday()
  },
  methods: {
    async loadToday() {
      this.loading = true
      this.error = ''
      try {
        const response = await listmak.getListMakByDate(this.today)
        if (response.success && response.data) {
          const raw = Array.isArray(response.data) ? response.data : [response.data]
          this.listmaks = raw.map(this.summarize)
        } else {
          this.listmaks = []
        }
      } catch (err) {
        console.error('Failed to load today listmaks:', err)
        this.error = 'Gagal memuat data. Periksa koneksi lalu coba lagi.'
      } finally {
        this.loading = false
      }
    },
    // Derive per-card summary counts from the preloaded orders.
    summarize(lm) {
      const orders = lm.orders || []
      const totalOrders = lm.total_orders ?? orders.length
      const paidCount = orders.filter(o => o.is_paid).length
      const unpaidCount = Math.max(totalOrders - paidCount, 0)
      return {
        id: lm.id,
        title: lm.title,
        totalOrders,
        paidCount,
        unpaidCount
      }
    },
    closeCreateModal() {
      if (this.creating) return
      this.showCreateModal = false
      this.newTitle = ''
      this.createError = ''
    },
    async createListmak() {
      if (!this.newTitle.trim() || this.creating) return
      this.creating = true
      this.createError = ''
      try {
        const response = await listmak.createListMak({
          title: this.newTitle.trim(),
          date: new Date(this.today).toISOString()
        })
        if (response.success && response.data) {
          this.showCreateModal = false
          this.newTitle = ''
          this.$router.push({ path: `/listmak/${response.data.id}` })
        } else {
          this.createError = 'Gagal membuat ListMak.'
        }
      } catch (err) {
        this.createError = err.message || 'Gagal membuat ListMak. Coba lagi.'
      } finally {
        this.creating = false
      }
    },
    openListmak(id) {
      this.$router.push({ path: `/listmak/${id}` })
    }
  }
}
</script>

<style scoped>
.today-container {
  padding: 1rem;
  padding-bottom: 2rem;
}

.page-header {
  margin-bottom: 1rem;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 0.25rem;
}

.page-date {
  font-size: 0.8125rem;
  color: #64748b;
}

/* Create button — full-width, labelled */
.create-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.875rem 1rem;
  margin-bottom: 1.25rem;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border: none;
  border-radius: 0.875rem;
  color: white;
  font-size: 0.9375rem;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.15s, opacity 0.15s;
}

.create-btn:hover {
  opacity: 0.95;
  transform: translateY(-1px);
}

.create-btn i {
  font-size: 1rem;
}

/* Cards */
.cards {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.listmak-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.875rem;
  padding: 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.card-main {
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 1rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.2rem;
  overflow-wrap: anywhere;
}

.card-count {
  font-size: 0.8125rem;
  color: #94a3b8;
  margin-bottom: 0.625rem;
}

.status-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.55rem;
  border-radius: 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-chip i {
  font-size: 0.8rem;
}

.status-paid {
  background: rgba(34, 197, 94, 0.12);
  color: #22c55e;
}

.status-unpaid {
  background: rgba(234, 179, 8, 0.12);
  color: #eab308;
}

/* Open button — label + icon, not just an arrow */
.open-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.625rem 0.9rem;
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 0.625rem;
  color: #3b82f6;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.open-btn:hover {
  background: rgba(59, 130, 246, 0.25);
}

.open-btn i {
  font-size: 0.8rem;
}

/* State blocks */
.state-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 2.5rem 1rem;
  color: #64748b;
  gap: 0.75rem;
}

.state-block > i {
  font-size: 2rem;
  color: #3b82f6;
}

.state-error > i {
  color: #ef4444;
}

.retry-btn,
.empty-action-btn {
  margin-top: 0.5rem;
  padding: 0.6rem 1.1rem;
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 0.625rem;
  color: #3b82f6;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
}

.empty .empty-icon {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: rgba(30, 41, 59, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty .empty-icon i {
  font-size: 1.75rem;
  color: #475569;
}

.empty h3 {
  font-size: 1.0625rem;
  font-weight: 600;
  color: #f1f5f9;
}

.empty p {
  font-size: 0.875rem;
  color: #64748b;
  max-width: 18rem;
}

@media (min-width: 768px) {
  .today-container {
    padding: 1.5rem 2rem;
    max-width: 720px;
  }
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: #1e293b;
  border-radius: 1rem;
  width: 100%;
  max-width: 400px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.modal-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #f1f5f9;
  margin: 0;
}

.modal-close {
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  font-size: 1rem;
  transition: color 0.15s, background 0.15s;
}

.modal-close:hover {
  color: #f1f5f9;
  background: rgba(255, 255, 255, 0.1);
}

.modal-body {
  padding: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-group label {
  font-size: 0.75rem;
  font-weight: 500;
  color: #94a3b8;
}

.modal-input {
  width: 100%;
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.9375rem;
  padding: 0.75rem;
  font-family: inherit;
  transition: border-color 0.15s;
  box-sizing: border-box;
}

.modal-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.modal-input::placeholder {
  color: #64748b;
}

.create-error {
  margin-top: 0.75rem;
  font-size: 0.8125rem;
  color: #ef4444;
}

.modal-footer {
  padding: 1rem 1.25rem;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.submit-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.875rem;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border: none;
  border-radius: 0.625rem;
  color: white;
  font-size: 0.9375rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.submit-btn:not(:disabled):hover {
  opacity: 0.9;
}
</style>
