<template>
  <div class="order-list-container">
    <!-- Header -->
    <div class="page-header">
      <button class="back-btn" @click="goBack">
        <i class="pi pi-arrow-left"></i>
        <span>Kembali</span>
      </button>
      <div class="header-title-row">
        <h1 class="page-title">{{ listmakTitle }}</h1>
        <button class="share-btn" @click="openShare">
          <i class="pi pi-share-alt"></i>
          <span>Bagikan</span>
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat pesanan...</p>
    </div>

    <!-- Error -->
    <div v-else-if="loadError" class="state-block state-error">
      <i class="pi pi-exclamation-triangle"></i>
      <p>{{ loadError }}</p>
      <button class="retry-btn" @click="loadData">Coba lagi</button>
    </div>

    <template v-else>
      <!-- Summary 2-column -->
      <div class="summary-card">
        <div class="summary-item">
          <span class="summary-value">{{ totalOrders }}</span>
          <span class="summary-label">Total pesanan</span>
        </div>
        <div class="summary-divider"></div>
        <div class="summary-item">
          <span class="summary-value" :class="{ 'value-unpaid': unpaidGroupCount > 0 }">{{ unpaidGroupCount }}</span>
          <span class="summary-label">Belum bayar</span>
        </div>
      </div>

      <!-- Add order — full-width, explicit label -->
      <button class="add-btn" @click="openAddOrder">
        <i class="pi pi-plus"></i>
        <span>+ Tambah pesanan</span>
      </button>

      <!-- Empty -->
      <div v-if="groups.length === 0" class="state-block empty">
        <div class="empty-icon"><i class="pi pi-inbox"></i></div>
        <h3>Belum ada pesanan</h3>
        <p>Tekan "+ Tambah pesanan" untuk mulai.</p>
      </div>

      <!-- Grouped list -->
      <section v-else class="groups">
        <article v-for="group in visibleGroups" :key="group.key" class="group-card">
          <!-- Group header: name, total, paid toggle -->
          <div class="group-header">
            <div class="group-meta">
              <span class="group-name">{{ group.name }}</span>
              <span class="group-total">Rp {{ formatRupiah(group.total) }}</span>
            </div>
            <button
              class="paid-btn"
              :class="{
                'paid-btn--paid': group.allPaid && !group.hasUnpriced,
                'paid-btn--waiting': group.hasUnpriced,
                'paid-btn--loading': group.loading
              }"
              :disabled="group.hasUnpriced || group.loading"
              @click="togglePaid(group)"
            >
              <i v-if="group.loading" class="pi pi-spin pi-spinner"></i>
              <i v-else-if="group.hasUnpriced" class="pi pi-clock"></i>
              <i v-else-if="group.allPaid" class="pi pi-check-circle"></i>
              <i v-else class="pi pi-circle"></i>
              <span>{{ paidLabel(group) }}</span>
            </button>
          </div>

          <!-- Items — all shown, never collapsed -->
          <ul class="item-list">
            <li v-for="order in group.orders" :key="order.id" class="item-row">
              <div class="item-info">
                <span class="item-name">{{ order.order_detail }}</span>
                <span v-if="!order.price || order.price === 0" class="item-no-price">
                  <i class="pi pi-exclamation-circle"></i>
                  harga belum diisi
                </span>
                <span v-else class="item-price">
                  Rp {{ formatRupiah((order.price || 0) * (order.qty || 1)) }}
                  <span v-if="order.qty > 1" class="item-qty">{{ order.qty }}x</span>
                </span>
              </div>
              <button class="edit-btn" @click="openEditOrder(order)" title="Edit pesanan">
                <i class="pi pi-pencil"></i>
              </button>
            </li>
          </ul>
        </article>

        <!-- Explicit load-more — no infinite scroll -->
        <button v-if="hasMore" class="load-more-btn" @click="loadMore">
          Muat lebih banyak ({{ groups.length - visibleCount }} lagi)
        </button>
      </section>
    </template>

    <Toast position="top-center" />
  </div>
</template>

<script>
import Toast from 'primevue/toast'
import { listmak } from '../api'

export default {
  name: 'OrderListView',
  components: { Toast },

  data() {
    return {
      listmakTitle: '',
      orders: [],
      loading: false,
      loadError: '',
      loadingGroups: {},
      visibleCount: 10
    }
  },

  computed: {
    listmakId() {
      return Number(this.$route.params.id)
    },

    groups() {
      const groupMap = new Map()
      for (const order of this.orders) {
        const key = order.name.trim().toLowerCase()
        if (!groupMap.has(key)) {
          groupMap.set(key, { name: order.name.trim(), orders: [] })
        }
        groupMap.get(key).orders.push(order)
      }

      return Array.from(groupMap.values()).map(g => {
        const key = g.name.toLowerCase()
        const hasUnpriced = g.orders.some(o => !o.price || o.price === 0)
        const allPaid = g.orders.every(o => o.is_paid)
        const total = g.orders.reduce((sum, o) => sum + (o.price || 0) * (o.qty || 1), 0)
        return {
          key,
          name: g.name,
          orders: g.orders,
          hasUnpriced,
          allPaid,
          total,
          loading: !!this.loadingGroups[key]
        }
      })
    },

    visibleGroups() {
      return this.groups.slice(0, this.visibleCount)
    },

    hasMore() {
      return this.visibleCount < this.groups.length
    },

    totalOrders() {
      return this.orders.length
    },

    unpaidGroupCount() {
      return this.groups.filter(g => !g.allPaid).length
    }
  },

  mounted() {
    this.loadData()
  },

  methods: {
    async loadData() {
      this.loading = true
      this.loadError = ''
      try {
        const [lmRes, ordersRes] = await Promise.all([
          listmak.getListMakById(this.listmakId),
          listmak.getOrders(this.listmakId)
        ])
        if (lmRes.success && lmRes.data) {
          this.listmakTitle = lmRes.data.title || `Listmak #${this.listmakId}`
        }
        if (ordersRes.success && ordersRes.data) {
          this.orders = Array.isArray(ordersRes.data) ? ordersRes.data : []
        }
      } catch {
        this.loadError = 'Gagal memuat data. Periksa koneksi lalu coba lagi.'
      } finally {
        this.loading = false
      }
    },

    async refreshOrders() {
      try {
        const res = await listmak.getOrders(this.listmakId)
        if (res.success && res.data) {
          this.orders = Array.isArray(res.data) ? res.data : []
        }
      } catch (err) {
        console.error('Failed to refresh orders:', err)
      }
    },

    async togglePaid(group) {
      const key = group.key
      this.loadingGroups = { ...this.loadingGroups, [key]: true }
      try {
        await listmak.updateOrdersPaidByName(this.listmakId, group.name, !group.allPaid)
        await this.refreshOrders()
      } catch (err) {
        this.$toast.add({
          severity: 'error',
          summary: 'Gagal update',
          detail: err.message || 'Gagal mengubah status bayar. Coba lagi.',
          life: 3000
        })
      } finally {
        const { [key]: _, ...rest } = this.loadingGroups
        this.loadingGroups = rest
      }
    },

    paidLabel(group) {
      if (group.loading) return '...'
      if (group.hasUnpriced) return 'Tunggu'
      return group.allPaid ? 'Lunas' : 'Belum'
    },

    formatRupiah(amount) {
      return Number(amount || 0).toLocaleString('id-ID')
    },

    loadMore() {
      this.visibleCount += 10
    },

    goBack() {
      this.$router.push('/today')
    },

    openShare() {
      // TODO 5d: modal Bagikan
    },

    openAddOrder() {
      // TODO 5c: modal Tambah pesanan
    },

    openEditOrder(_order) {
      // TODO 5c: edit order modal
    }
  }
}
</script>

<style scoped>
.order-list-container {
  padding: 1rem;
  padding-bottom: 2rem;
}

/* Header */
.page-header {
  margin-bottom: 1.25rem;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.375rem 0;
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  margin-bottom: 0.5rem;
}

.back-btn:hover {
  color: #cbd5e1;
}

.header-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.page-title {
  font-size: 1.375rem;
  font-weight: 700;
  color: #f1f5f9;
  overflow-wrap: anywhere;
  flex: 1;
  min-width: 0;
}

.share-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.875rem;
  background: rgba(99, 102, 241, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 0.625rem;
  color: #818cf8;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.share-btn:hover {
  background: rgba(99, 102, 241, 0.25);
}

/* Summary card */
.summary-card {
  display: flex;
  align-items: stretch;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.875rem;
  padding: 1rem;
  margin-bottom: 1rem;
}

.summary-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
}

.summary-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: #f1f5f9;
  line-height: 1;
}

.summary-value.value-unpaid {
  color: #eab308;
}

.summary-label {
  font-size: 0.75rem;
  color: #64748b;
  text-align: center;
}

.summary-divider {
  width: 1px;
  background: rgba(255, 255, 255, 0.06);
  margin: 0 0.5rem;
}

/* Add order button — full-width, labelled */
.add-btn {
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

.add-btn:hover {
  opacity: 0.95;
  transform: translateY(-1px);
}

/* Groups */
.groups {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.group-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.875rem;
  overflow: hidden;
}

.group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.group-meta {
  flex: 1;
  min-width: 0;
}

.group-name {
  display: block;
  font-size: 0.9375rem;
  font-weight: 700;
  color: #f1f5f9;
  overflow-wrap: anywhere;
}

.group-total {
  display: block;
  font-size: 0.8125rem;
  color: #94a3b8;
  margin-top: 0.125rem;
}

/* Paid toggle button */
.paid-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, opacity 0.15s;
  border: 1px solid transparent;
  /* default: belum bayar */
  background: rgba(234, 179, 8, 0.12);
  border-color: rgba(234, 179, 8, 0.25);
  color: #eab308;
}

.paid-btn--paid {
  background: rgba(34, 197, 94, 0.12);
  border-color: rgba(34, 197, 94, 0.25);
  color: #22c55e;
}

.paid-btn--waiting {
  background: rgba(100, 116, 139, 0.12);
  border-color: rgba(100, 116, 139, 0.2);
  color: #64748b;
  cursor: not-allowed;
}

.paid-btn--loading {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Item list */
.item-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
}

.item-row:last-child {
  border-bottom: none;
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
  display: block;
  font-size: 0.875rem;
  color: #e2e8f0;
  overflow-wrap: anywhere;
}

.item-price {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8125rem;
  color: #94a3b8;
  margin-top: 0.15rem;
}

.item-qty {
  font-size: 0.75rem;
  color: #64748b;
}

.item-no-price {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.75rem;
  color: #f97316;
  margin-top: 0.15rem;
}

.item-no-price i {
  font-size: 0.75rem;
}

.edit-btn {
  flex-shrink: 0;
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.5rem;
  color: #64748b;
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
}

.edit-btn:hover {
  color: #94a3b8;
  background: rgba(255, 255, 255, 0.08);
}

.edit-btn i {
  font-size: 0.8rem;
}

/* Load more */
.load-more-btn {
  width: 100%;
  padding: 0.875rem;
  background: rgba(30, 41, 59, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.875rem;
  color: #94a3b8;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.load-more-btn:hover {
  background: rgba(30, 41, 59, 0.7);
  color: #cbd5e1;
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
}

@media (min-width: 768px) {
  .order-list-container {
    padding: 1.5rem 2rem;
    max-width: 720px;
  }
}
</style>
