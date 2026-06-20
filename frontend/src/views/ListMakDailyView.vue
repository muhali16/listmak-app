<template>
  <div class="history-container">
    <header class="page-header">
      <h1 class="page-title">Riwayat</h1>
    </header>

    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat riwayat...</p>
    </div>

    <div v-else-if="error" class="state-block state-error">
      <i class="pi pi-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <button class="retry-btn" @click="loadHistory">Coba lagi</button>
    </div>

    <div v-else-if="dateGroups.length === 0" class="state-block empty">
      <div class="empty-icon"><i class="pi pi-inbox"></i></div>
      <h3>Belum ada riwayat</h3>
      <p>Listmak yang dibuat akan muncul di sini.</p>
    </div>

    <section v-else class="date-list">
      <div v-for="group in dateGroups" :key="group.date" class="date-group">
        <!-- Date row — tap to expand/collapse -->
        <button class="date-row" @click="toggleGroup(group)">
          <div class="date-info">
            <span class="date-label">{{ formatDate(group.date) }}</span>
            <span class="date-count">{{ group.listmaks.length }} listmak</span>
          </div>
          <i class="pi" :class="group.expanded ? 'pi-chevron-up' : 'pi-chevron-down'"></i>
        </button>

        <!-- Expanded: listmak cards + create button -->
        <div v-if="group.expanded" class="date-content">
          <button class="create-btn" @click="goCreate(group.date)">
            <i class="pi pi-plus"></i>
            <span>Buat listmak baru</span>
          </button>

          <article v-for="lm in group.listmaks" :key="lm.id" class="listmak-card">
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
        </div>
      </div>
    </section>
  </div>
</template>

<script>
import { listmak } from '../api'

export default {
  name: 'ListMakDailyView',

  data() {
    return {
      dateGroups: [],
      loading: false,
      error: ''
    }
  },

  mounted() {
    this.loadHistory()
  },

  methods: {
    async loadHistory() {
      this.loading = true
      this.error = ''
      try {
        const res = await listmak.getAllListMaks()
        if (res.success && res.data) {
          const all = Array.isArray(res.data) ? res.data : [res.data]
          const grouped = {}
          for (const lm of all) {
            const date = lm.date
            if (!grouped[date]) grouped[date] = []
            grouped[date].push(lm)
          }
          this.dateGroups = Object.keys(grouped)
            .sort((a, b) => b.localeCompare(a))
            .map((date, i) => ({
              date,
              listmaks: grouped[date].map(this.summarize),
              expanded: i === 0
            }))
        }
      } catch {
        this.error = 'Gagal memuat riwayat. Periksa koneksi lalu coba lagi.'
      } finally {
        this.loading = false
      }
    },

    summarize(lm) {
      const orders = lm.orders || []
      const totalOrders = lm.total_orders ?? orders.length
      const paidCount = orders.filter(o => o.is_paid).length
      const unpaidCount = Math.max(totalOrders - paidCount, 0)
      return { id: lm.id, title: lm.title, totalOrders, paidCount, unpaidCount }
    },

    formatDate(dateStr) {
      const d = new Date(dateStr + 'T00:00:00')
      const today = new Date().toISOString().split('T')[0]
      const yesterday = new Date(Date.now() - 86400000).toISOString().split('T')[0]
      if (dateStr === today) return 'Hari Ini — ' + d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
      if (dateStr === yesterday) return 'Kemarin — ' + d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
      return d.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
    },

    toggleGroup(group) {
      group.expanded = !group.expanded
    },

    goCreate(date) {
      this.$router.push({ path: '/listmak/input', query: { date } })
    },

    openListmak(id) {
      this.$router.push({ path: `/listmak/${id}` })
    }
  }
}
</script>

<style scoped>
.history-container {
  padding: 1rem;
  padding-bottom: 2rem;
}

.page-header {
  margin-bottom: 1.25rem;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #f1f5f9;
}

/* Date list */
.date-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.date-group {
  border-radius: 0.875rem;
  overflow: hidden;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.date-row {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1rem;
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
  gap: 0.75rem;
}

.date-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.date-info {
  flex: 1;
  min-width: 0;
}

.date-label {
  display: block;
  font-size: 0.9375rem;
  font-weight: 600;
  color: #f1f5f9;
  overflow-wrap: anywhere;
}

.date-count {
  display: block;
  font-size: 0.75rem;
  color: #64748b;
  margin-top: 0.125rem;
}

.date-row > i {
  flex-shrink: 0;
  color: #475569;
  font-size: 0.875rem;
}

/* Expanded content */
.date-content {
  padding: 0 0.75rem 0.75rem;
  border-top: 1px solid rgba(255, 255, 255, 0.04);
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  padding-top: 0.75rem;
}

/* Create button — same as TodayView */
.create-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border: none;
  border-radius: 0.75rem;
  color: white;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.create-btn:hover {
  opacity: 0.9;
}

/* Listmak cards — identical to TodayView */
.listmak-card {
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.75rem;
  padding: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.card-main {
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.2rem;
  overflow-wrap: anywhere;
}

.card-count {
  font-size: 0.75rem;
  color: #94a3b8;
  margin-bottom: 0.5rem;
}

.status-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.2rem 0.5rem;
  border-radius: 0.4rem;
  font-size: 0.6875rem;
  font-weight: 600;
}

.status-chip i {
  font-size: 0.7rem;
}

.status-paid {
  background: rgba(34, 197, 94, 0.12);
  color: #22c55e;
}

.status-unpaid {
  background: rgba(234, 179, 8, 0.12);
  color: #eab308;
}

.open-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.75rem;
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 0.5rem;
  color: #3b82f6;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.open-btn:hover {
  background: rgba(59, 130, 246, 0.25);
}

.open-btn i {
  font-size: 0.75rem;
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

.retry-btn {
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
}

@media (min-width: 768px) {
  .history-container {
    padding: 1.5rem 2rem;
    max-width: 720px;
  }
}
</style>
