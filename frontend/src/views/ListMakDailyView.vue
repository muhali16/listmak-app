<template>
  <div class="daily-container">
    <!-- Header -->
    <header class="page-header">
      <h1 class="page-title">Riwayat Harian</h1>
      <p class="page-subtitle">Lihat pesanan berdasarkan tanggal</p>
    </header>

    <!-- Date Picker -->
    <section class="date-section">
      <div class="date-picker-wrapper">
        <Button 
          icon="pi pi-chevron-left" 
          text 
          rounded
          @click="previousDay"
          class="date-nav-btn"
        />
        <div class="date-display" @click="showDatePicker = !showDatePicker">
          <i class="pi pi-calendar"></i>
          <span>{{ formattedDate }}</span>
        </div>
        <Button 
          icon="pi pi-chevron-right" 
          text 
          rounded
          @click="nextDay"
          class="date-nav-btn"
          :disabled="isToday"
        />
      </div>
      
      <input 
        v-if="showDatePicker"
        type="date" 
        v-model="selectedDate"
        @change="loadOrdersByDate"
        class="date-input"
      />
    </section>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat data...</p>
    </div>

    <!-- Data Content -->
    <div v-else-if="orders.length > 0" class="data-content">
      <!-- Summary -->
      <section class="summary-section">
        <div class="summary-grid">
          <div class="summary-card">
            <i class="pi pi-list"></i>
            <div class="summary-info">
              <span class="summary-value">{{ orders.length }}</span>
              <span class="summary-label">Total</span>
            </div>
          </div>
          <div class="summary-card summary-green">
            <i class="pi pi-check-circle"></i>
            <div class="summary-info">
              <span class="summary-value">{{ paidCount }}</span>
              <span class="summary-label">Bayar</span>
            </div>
          </div>
          <div class="summary-card summary-yellow">
            <i class="pi pi-clock"></i>
            <div class="summary-info">
              <span class="summary-value">{{ unpaidCount }}</span>
              <span class="summary-label">Belum</span>
            </div>
          </div>
          <div class="summary-card summary-blue">
            <i class="pi pi-wallet"></i>
            <div class="summary-info">
              <span class="summary-value">{{ formatShortCurrency(totalAmount) }}</span>
              <span class="summary-label">Total</span>
            </div>
          </div>
        </div>
      </section>
  
      <!-- Orders List -->
      <section class="orders-section">
        <div class="orders-list">
          <div 
            v-for="(order, index) in orders" 
            :key="index" 
            class="order-card"
            :class="{ 'order-paid': order.paid }"
          >
            <div class="order-status">
              <i :class="order.paid ? 'pi pi-check-circle' : 'pi pi-circle'"></i>
            </div>
            <div class="order-content">
              <div class="order-header">
                <span class="order-name">{{ order.name }}</span>
                <span class="order-total">{{ formatCurrency(order.price * order.qty) }}</span>
              </div>
              <p class="order-detail">{{ order.order }}</p>
              <div class="order-meta">
                <span>{{ formatCurrency(order.price) }} x {{ order.qty }}</span>
                <div class="meta-right">
                  <span v-if="order.listmakTitle" class="listmak-badge">{{ order.listmakTitle }}</span>
                  <span :class="order.paid ? 'status-paid' : 'status-unpaid'">
                    {{ order.paid ? 'Sudah Bayar' : 'Belum Bayar' }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Empty State -->
    <section v-else class="empty-state">
      <div class="empty-icon">
        <i class="pi pi-calendar-times"></i>
      </div>
      <h3>Tidak Ada Data</h3>
      <p>Belum ada pesanan untuk tanggal ini</p>
      <router-link :to="{ path: '/listmak/input', query: { date: selectedDate } }" class="empty-action">
        <Button label="Input Pesanan" icon="pi pi-plus" />
      </router-link>
    </section>
  </div>
</template>

<script>
import Button from 'primevue/button'
import { listmak } from '../api'

export default {
  name: 'ListMakDailyView',
  components: {
    Button
  },
  data() {
    return {
      selectedDate: new Date().toISOString().split('T')[0],
      orders: [],
      showDatePicker: false,
      loading: false
    }
  },
  computed: {
    formattedDate() {
      const date = new Date(this.selectedDate)
      const options = { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' }
      return date.toLocaleDateString('id-ID', options)
    },
    isToday() {
      return this.selectedDate === new Date().toISOString().split('T')[0]
    },
    paidCount() {
      return this.orders.filter(o => o.paid).length
    },
    unpaidCount() {
      return this.orders.filter(o => !o.paid).length
    },
    totalAmount() {
      return this.orders.reduce((sum, o) => sum + (o.price * o.qty), 0)
    }
  },
  mounted() {
    this.loadOrdersByDate()
  },
  methods: {
    async loadOrdersByDate() {
      this.showDatePicker = false
      this.loading = true
      this.orders = []
      
      try {
        const response = await listmak.getListMakByDate(this.selectedDate)
        if (response.success && response.data) {
          // Flatten orders from all listmaks of the day
          const listmaks = Array.isArray(response.data) ? response.data : [response.data]
          
          this.orders = listmaks.flatMap(lm => (lm.orders || []).map(o => ({
             id: o.id,
             name: o.name,
             order: o.order_detail || o.order,
             price: o.price,
             qty: o.qty,
             paid: o.is_paid,
             listmakTitle: lm.title
          })))
        }
      } catch (error) {
        console.error('Failed to load daily orders:', error)
      } finally {
        this.loading = false
      }
    },
    previousDay() {
      const date = new Date(this.selectedDate)
      date.setDate(date.getDate() - 1)
      this.selectedDate = date.toISOString().split('T')[0]
      this.loadOrdersByDate()
    },
    nextDay() {
      const date = new Date(this.selectedDate)
      date.setDate(date.getDate() + 1)
      this.selectedDate = date.toISOString().split('T')[0]
      this.loadOrdersByDate()
    },
    formatCurrency(value) {
      return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0
      }).format(value)
    },
    formatShortCurrency(value) {
      if (value >= 1000000) {
        return `${(value / 1000000).toFixed(1)}jt`
      }
      if (value >= 1000) {
        return `${(value / 1000).toFixed(0)}rb`
      }
      return value.toString()
    }
  }
}
</script>

<style scoped>
.daily-container {
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
  margin-bottom: 0.25rem;
}

.page-subtitle {
  font-size: 0.8125rem;
  color: #64748b;
}

/* Date Section */
.date-section {
  margin-bottom: 1.25rem;
}

.date-picker-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.75rem;
  padding: 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.date-nav-btn {
  color: #94a3b8 !important;
}

.date-nav-btn:hover {
  color: #f1f5f9 !important;
  background: rgba(255, 255, 255, 0.1) !important;
}

.date-display {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 0.5rem;
  cursor: pointer;
  color: #3b82f6;
  font-weight: 500;
  font-size: 0.875rem;
}

.date-input {
  width: 100%;
  margin-top: 0.5rem;
  padding: 0.75rem;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.875rem;
}

/* Summary */
.summary-section {
  margin-bottom: 1.25rem;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.5rem;
}

.summary-card {
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.75rem;
  padding: 0.75rem 0.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.summary-card i {
  font-size: 1.25rem;
  color: #64748b;
}

.summary-card.summary-green i { color: #22c55e; }
.summary-card.summary-yellow i { color: #eab308; }
.summary-card.summary-blue i { color: #3b82f6; }

.summary-info {
  text-align: center;
}

.summary-value {
  display: block;
  font-size: 1rem;
  font-weight: 700;
  color: #f1f5f9;
}

.summary-label {
  font-size: 0.625rem;
  color: #64748b;
}

/* Orders List */
.orders-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.order-card {
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.75rem;
  padding: 0.875rem;
  display: flex;
  gap: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.order-card.order-paid {
  background: rgba(34, 197, 94, 0.05);
  border-color: rgba(34, 197, 94, 0.1);
}

.order-status {
  flex-shrink: 0;
}

.order-status i {
  font-size: 1.25rem;
  color: #475569;
}

.order-paid .order-status i {
  color: #22c55e;
}

.order-content {
  flex: 1;
  min-width: 0;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.25rem;
}

.order-name {
  font-weight: 600;
  color: #f1f5f9;
  font-size: 0.9375rem;
}

.order-total {
  font-weight: 700;
  color: #3b82f6;
  font-size: 0.875rem;
  flex-shrink: 0;
}

.order-detail {
  font-size: 0.8125rem;
  color: #94a3b8;
  margin-bottom: 0.5rem;
  line-height: 1.4;
}

.order-meta {
  display: flex;
  justify-content: space-between;
  font-size: 0.6875rem;
  color: #64748b;
}

.status-paid {
  color: #22c55e;
}

.status-unpaid {
  color: #eab308;
}


/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  color: #64748b;
  gap: 1rem;
}

.loading-state i {
  font-size: 2rem;
  color: #3b82f6;
}

/* Listmak Badge */
.meta-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.listmak-badge {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  font-size: 0.6rem;
  font-weight: 500;
  text-transform: uppercase;
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 3rem 1rem;
}

.empty-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 1rem;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-icon i {
  font-size: 2rem;
  color: #475569;
}

.empty-state h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.5rem;
}

.empty-state p {
  font-size: 0.875rem;
  color: #64748b;
  margin-bottom: 1.5rem;
}

.empty-action {
  text-decoration: none;
}

@media (min-width: 768px) {
  .daily-container {
    padding: 1.5rem 2rem;
  }
}
</style>
