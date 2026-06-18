<template>
  <div class="view-container">
    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat...</p>
    </div>

    <!-- Not Found State -->
    <div v-else-if="notFound" class="not-found-state">
      <div class="not-found-icon">
        <i class="pi pi-search"></i>
      </div>
      <h2>ListMak Tidak Ditemukan</h2>
      <p class="not-found-message">
        Link yang Anda akses tidak valid atau ListMak sudah dihapus.
      </p>
    </div>

    <!-- View State -->
    <div v-else class="view-state">
      <!-- Header -->
      <header class="view-header">
        <h1 class="view-title">{{ listmakData.title }}</h1>
        <p class="view-date">{{ formatDate(listmakData.createdAt) }}</p>
      </header>

      <!-- Summary Cards -->
      <div class="summary-section">
        <div class="summary-grid">
          <div class="summary-card">
            <span class="summary-label">Total</span>
            <span class="summary-value">{{ listmakData.orders.length }}</span>
          </div>
          <div class="summary-card summary-green">
            <span class="summary-label">Bayar</span>
            <span class="summary-value">{{ paidCount }}</span>
          </div>
          <div class="summary-card summary-blue">
            <span class="summary-label">Total Biaya</span>
            <span class="summary-value">{{ formatShortCurrency(totalAmount) }}</span>
          </div>
          <div class="summary-card summary-purple">
            <span class="summary-label">Terbayar</span>
            <span class="summary-value">{{ formatShortCurrency(paidAmount) }}</span>
          </div>
        </div>
      </div>

      <!-- Orders List -->
      <div class="orders-section">
        <h3 class="section-title">Daftar Pesanan</h3>
        <div class="orders-list">
          <div 
            v-for="(order, index) in listmakData.orders" 
            :key="index" 
            class="order-item"
            :class="{ 'order-paid': order.paid }"
          >
            <div class="order-status">
              <i v-if="order.paid" class="pi pi-check-circle status-paid"></i>
              <i v-else class="pi pi-circle status-unpaid"></i>
            </div>
            <div class="order-details">
              <div class="order-name">{{ order.name }}</div>
              <div class="order-content">{{ order.order }}</div>
            </div>
            <div class="order-price">
              <div class="price-amount" v-if="order.price">
                {{ formatShortCurrency(order.price * order.qty) }}
              </div>
              <div class="price-qty" v-if="order.qty > 1">x{{ order.qty }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Legend -->
      <div class="legend-section">
        <div class="legend-item">
          <i class="pi pi-check-circle status-paid"></i>
          <span>Sudah Bayar</span>
        </div>
        <div class="legend-item">
          <i class="pi pi-circle status-unpaid"></i>
          <span>Belum Bayar</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { share } from '../api'

export default {
  name: 'ViewListMakView',
  data() {
    return {
      loading: true,
      notFound: false,
      listmakData: {
        title: '',
        orders: [],
        createdAt: null
      }
    }
  },
  computed: {
    viewId() {
      return this.$route.params.viewId
    },
    paidCount() {
      return this.listmakData.orders.filter(o => o.paid).length
    },
    totalAmount() {
      return this.listmakData.orders.reduce((sum, o) => sum + (o.price * o.qty), 0)
    },
    paidAmount() {
      return this.listmakData.orders.filter(o => o.paid).reduce((sum, o) => sum + (o.price * o.qty), 0)
    }
  },
  mounted() {
    this.loadListMakData()
  },
  methods: {
    async loadListMakData() {
      try {
        const response = await share.getViewShare(this.viewId)
        
        if (response.success && response.data) {
           const snapshot = response.data.snapshot || {}
           
           this.listmakData = {
              title: response.data.title || snapshot.title,
              createdAt: response.data.created_at,
              orders: (snapshot.orders || []).map(o => ({
                 name: o.name,
                 order: o.order_detail || o.order,
                 price: o.price,
                 qty: o.qty,
                 paid: o.is_paid
              }))
           }
        } else {
           this.notFound = true
        }
      } catch (error) {
        console.error('Failed to load view link:', error)
        this.notFound = true
      } finally {
        this.loading = false
      }
    },
    formatDate(dateString) {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toLocaleDateString('id-ID', {
        weekday: 'long',
        day: 'numeric',
        month: 'long',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    },
    formatShortCurrency(value) {
      if (!value) return '-'
      if (value >= 1000000) {
        return `${(value / 1000000).toFixed(1)}jt`
      }
      if (value >= 1000) {
        return `${Math.round(value / 1000)}rb`
      }
      return value.toString()
    }
  }
}
</script>

<style scoped>
.view-container {
  min-height: 100vh;
  min-height: 100dvh;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  padding: 1rem;
}

/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  height: 100dvh;
  gap: 1rem;
  color: #64748b;
}

.loading-state i {
  font-size: 2.5rem;
  color: #3b82f6;
}

/* Not Found State */
.not-found-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  height: 100dvh;
  text-align: center;
  padding: 1.5rem;
}

.not-found-icon {
  width: 100px;
  height: 100px;
  background: rgba(234, 179, 8, 0.15);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.not-found-icon i {
  font-size: 3rem;
  color: #eab308;
}

.not-found-state h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 0.5rem;
}

.not-found-message {
  font-size: 0.9375rem;
  color: #94a3b8;
  max-width: 300px;
}

/* View State */
.view-state {
  max-width: 600px;
  margin: 0 auto;
}

.view-header {
  text-align: center;
  padding: 1rem 0 1.5rem;
}

.view-title {
  font-size: 1.375rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 0.25rem;
}

.view-date {
  font-size: 0.75rem;
  color: #64748b;
}

/* Summary Section */
.summary-section {
  margin-bottom: 1.5rem;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.375rem;
}

.summary-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.5rem;
  padding: 0.5rem;
  text-align: center;
}

.summary-label {
  display: block;
  font-size: 0.5625rem;
  color: #64748b;
  margin-bottom: 0.125rem;
}

.summary-value {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #f1f5f9;
}

.summary-green .summary-value { color: #22c55e; }
.summary-blue .summary-value { color: #3b82f6; }
.summary-purple .summary-value { color: #a855f7; }

/* Orders Section */
.orders-section {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.75rem;
  padding: 1rem;
  margin-bottom: 1rem;
}

.section-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #94a3b8;
  margin-bottom: 0.75rem;
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.order-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem;
  background: rgba(15, 23, 42, 0.4);
  border-radius: 0.5rem;
  border: 1px solid transparent;
}

.order-item.order-paid {
  background: rgba(34, 197, 94, 0.08);
  border-color: rgba(34, 197, 94, 0.15);
}

.order-status {
  flex-shrink: 0;
  padding-top: 0.125rem;
}

.status-paid {
  color: #22c55e;
  font-size: 1.125rem;
}

.status-unpaid {
  color: #475569;
  font-size: 1.125rem;
}

.order-details {
  flex: 1;
  min-width: 0;
}

.order-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.125rem;
}

.order-paid .order-name {
  text-decoration: line-through;
  color: #64748b;
}

.order-content {
  font-size: 0.75rem;
  color: #94a3b8;
  line-height: 1.4;
}

.order-paid .order-content {
  text-decoration: line-through;
  color: #64748b;
}

.order-price {
  text-align: right;
  flex-shrink: 0;
}

.price-amount {
  font-size: 0.875rem;
  font-weight: 700;
  color: #22c55e;
}

.price-qty {
  font-size: 0.625rem;
  color: #64748b;
}

/* Legend Section */
.legend-section {
  display: flex;
  justify-content: center;
  gap: 1.5rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.6875rem;
  color: #64748b;
}

.legend-item .status-paid {
  font-size: 0.875rem;
}

.legend-item .status-unpaid {
  font-size: 0.875rem;
}
</style>
