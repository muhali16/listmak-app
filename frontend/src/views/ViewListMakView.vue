<template>
  <div class="view-container">
    <!-- Loading -->
    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat...</p>
    </div>

    <!-- Not Found -->
    <div v-else-if="notFound" class="state-block">
      <div class="state-icon">
        <i class="pi pi-search"></i>
      </div>
      <h2 class="state-title">Link Tidak Valid</h2>
      <p class="state-desc">Link tidak ditemukan atau listmak sudah dihapus.</p>
    </div>

    <!-- View -->
    <div v-else class="view-state">
      <div class="page-header">
        <span class="page-badge">Lihat Saja</span>
        <h1 class="page-title">{{ listmakData.title }}</h1>
        <p class="page-date">{{ formatDate(listmakData.createdAt) }}</p>
      </div>

      <!-- Summary -->
      <div class="summary-card">
        <div class="summary-item">
          <span class="summary-value">{{ totalOrders }}</span>
          <span class="summary-label">Total pesanan</span>
        </div>
        <div class="summary-divider"></div>
        <div class="summary-item">
          <span
            class="summary-value"
            :class="{ 'value-unpaid': unpaidGroupCount > 0 }"
          >{{ unpaidGroupCount }}</span>
          <span class="summary-label">Belum bayar</span>
        </div>
      </div>

      <!-- Empty -->
      <div v-if="groups.length === 0" class="state-block empty">
        <i class="pi pi-inbox"></i>
        <p>Belum ada pesanan</p>
      </div>

      <!-- Groups -->
      <section v-else class="groups">
        <article v-for="group in groups" :key="group.key" class="group-card">
          <div class="group-header">
            <div class="group-meta">
              <span class="group-name">{{ group.name }}</span>
              <span v-if="group.total > 0" class="group-total">
                Rp {{ formatRupiah(group.total) }}
              </span>
            </div>
            <span
              class="paid-badge"
              :class="{
                'paid-badge--paid': group.allPaid && !group.hasUnpriced,
                'paid-badge--waiting': group.hasUnpriced,
                'paid-badge--unpaid': !group.allPaid && !group.hasUnpriced
              }"
            >
              <i v-if="group.hasUnpriced" class="pi pi-clock"></i>
              <i v-else-if="group.allPaid" class="pi pi-check-circle"></i>
              <i v-else class="pi pi-circle"></i>
              {{ group.hasUnpriced ? 'Tunggu' : group.allPaid ? 'Lunas' : 'Belum' }}
            </span>
          </div>

          <ul class="item-list">
            <li v-for="order in group.orders" :key="order.id" class="item-row">
              <div class="item-info">
                <span class="item-name">{{ order.order }}</span>
                <span v-if="!order.price" class="item-no-price">
                  <i class="pi pi-exclamation-circle"></i>
                  harga belum diisi
                </span>
                <span v-else class="item-price">
                  Rp {{ formatRupiah(order.price * (order.qty || 1)) }}
                  <span v-if="order.qty > 1" class="item-qty">{{ order.qty }}x</span>
                </span>
              </div>
            </li>
          </ul>
        </article>
      </section>
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
      listmakData: { title: '', createdAt: null, orders: [] }
    }
  },

  computed: {
    viewId() {
      return this.$route.params.viewId
    },

    groups() {
      const groupMap = new Map()
      for (const order of this.listmakData.orders) {
        const key = order.name.trim().toLowerCase()
        if (!groupMap.has(key)) {
          groupMap.set(key, { name: order.name.trim(), orders: [] })
        }
        groupMap.get(key).orders.push(order)
      }
      return Array.from(groupMap.values()).map(g => {
        const hasUnpriced = g.orders.some(o => !o.price)
        const allPaid = g.orders.every(o => o.paid)
        const total = g.orders.reduce((sum, o) => sum + (o.price || 0) * (o.qty || 1), 0)
        return {
          key: g.name.toLowerCase(),
          name: g.name,
          orders: g.orders,
          hasUnpriced,
          allPaid,
          total
        }
      })
    },

    totalOrders() {
      return this.listmakData.orders.length
    },

    unpaidGroupCount() {
      return this.groups.filter(g => !g.allPaid).length
    }
  },

  mounted() {
    this.loadViewData()
  },

  methods: {
    async loadViewData() {
      try {
        const res = await share.getViewShare(this.viewId)
        if (res.success && res.data) {
          const snapshot = res.data.snapshot || {}
          this.listmakData = {
            title: res.data.title || snapshot.title || 'Listmak',
            createdAt: res.data.created_at,
            orders: (snapshot.orders || []).map(o => ({
              id: o.id,
              name: o.name,
              order: o.order_detail || o.order,
              price: o.price || 0,
              qty: o.qty || 1,
              paid: o.is_paid
            }))
          }
        } else {
          this.notFound = true
        }
      } catch {
        this.notFound = true
      } finally {
        this.loading = false
      }
    },

    formatDate(dateString) {
      if (!dateString) return ''
      return new Date(dateString).toLocaleDateString('id-ID', {
        weekday: 'long', day: 'numeric', month: 'long', year: 'numeric'
      })
    },

    formatRupiah(amount) {
      return Number(amount || 0).toLocaleString('id-ID')
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

/* State blocks */
.state-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  min-height: 100dvh;
  text-align: center;
  padding: 2rem 1.5rem;
  gap: 0.75rem;
  color: #64748b;
}

.state-block > i {
  font-size: 2.5rem;
  color: #3b82f6;
}

.state-block.empty {
  min-height: unset;
  padding: 2.5rem 1rem;
}

.state-icon {
  width: 80px;
  height: 80px;
  background: rgba(234, 179, 8, 0.15);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.5rem;
}

.state-icon i {
  font-size: 2.25rem;
  color: #eab308;
}

.state-title {
  font-size: 1.375rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}

.state-desc {
  font-size: 0.9375rem;
  color: #94a3b8;
  max-width: 280px;
  line-height: 1.5;
}

/* View state */
.view-state {
  max-width: 560px;
  margin: 0 auto;
  padding-bottom: 3rem;
}

/* Page header */
.page-header {
  padding: 1.25rem 0 1rem;
}

.page-badge {
  display: inline-block;
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #94a3b8;
  background: rgba(100, 116, 139, 0.12);
  border: 1px solid rgba(100, 116, 139, 0.2);
  border-radius: 999px;
  padding: 0.25rem 0.625rem;
  margin-bottom: 0.625rem;
}

.page-title {
  font-size: 1.375rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0 0 0.25rem;
  overflow-wrap: anywhere;
}

.page-date {
  font-size: 0.8125rem;
  color: #64748b;
  margin: 0;
}

/* Summary card */
.summary-card {
  display: flex;
  align-items: stretch;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.875rem;
  padding: 1rem;
  margin-bottom: 1.25rem;
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

.summary-value.value-unpaid { color: #eab308; }

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

/* Paid badge (readonly) */
.paid-badge {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.375rem 0.625rem;
  border-radius: 0.5rem;
  font-size: 0.8125rem;
  font-weight: 600;
  border: 1px solid transparent;
}

.paid-badge--unpaid {
  background: rgba(234, 179, 8, 0.12);
  border-color: rgba(234, 179, 8, 0.25);
  color: #eab308;
}

.paid-badge--paid {
  background: rgba(34, 197, 94, 0.12);
  border-color: rgba(34, 197, 94, 0.25);
  color: #22c55e;
}

.paid-badge--waiting {
  background: rgba(100, 116, 139, 0.12);
  border-color: rgba(100, 116, 139, 0.2);
  color: #64748b;
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

.item-row:last-child { border-bottom: none; }

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

.item-no-price i { font-size: 0.75rem; }
</style>
