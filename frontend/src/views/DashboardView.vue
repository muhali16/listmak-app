<template>
  <div class="dashboard-container">
    <!-- Header -->
    <header class="page-header">
      <h1 class="page-title">Dashboard</h1>
      <p class="page-date">{{ formattedDate }}</p>
    </header>

    <!-- Welcome Section -->
    <section class="welcome-section">
      <div class="welcome-card">
        <div class="welcome-text">
          <h2>Selamat Datang! 👋</h2>
          <p>{{ userName }}</p>
        </div>
        <div class="welcome-icon">
          <i class="pi pi-chart-line"></i>
        </div>
      </div>
    </section>

    <!-- Quick Stats -->
    <section class="stats-section">
      <h3 class="section-title">Ringkasan Hari Ini</h3>
      <div class="stats-grid">
        <div class="stat-card stat-blue">
          <div class="stat-icon">
            <i class="pi pi-list"></i>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ todayStats.totalOrders }}</span>
            <span class="stat-label">Total Order</span>
          </div>
        </div>

        <div class="stat-card stat-green">
          <div class="stat-icon">
            <i class="pi pi-check-circle"></i>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ todayStats.paidOrders }}</span>
            <span class="stat-label">Sudah Bayar</span>
          </div>
        </div>

        <div class="stat-card stat-yellow">
          <div class="stat-icon">
            <i class="pi pi-clock"></i>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ todayStats.unpaidOrders }}</span>
            <span class="stat-label">Belum Bayar</span>
          </div>
        </div>

        <div class="stat-card stat-purple">
          <div class="stat-icon">
            <i class="pi pi-wallet"></i>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ formatCurrency(todayStats.totalAmount) }}</span>
            <span class="stat-label">Total Biaya</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Quick Actions -->
    <section class="actions-section">
      <h3 class="section-title">Aksi Cepat</h3>
      <div class="actions-grid">
        <router-link to="/listmak/input" class="action-card">
          <div class="action-icon action-blue">
            <i class="pi pi-plus"></i>
          </div>
          <span class="action-label">Input ListMak</span>
        </router-link>

        <router-link to="/listmak/daily" class="action-card">
          <div class="action-icon action-green">
            <i class="pi pi-calendar"></i>
          </div>
          <span class="action-label">Lihat Riwayat</span>
        </router-link>

        <router-link to="/contacts" class="action-card">
          <div class="action-icon action-yellow">
            <i class="pi pi-users"></i>
          </div>
          <span class="action-label">Import Kontak</span>
        </router-link>
      </div>
    </section>

    <!-- Info Section -->
    <section class="info-section">
      <div class="info-card">
        <i class="pi pi-info-circle"></i>
        <div class="info-content">
          <h4>Tips</h4>
          <p>Copy-paste daftar pesanan dari WhatsApp ke halaman Input ListMak untuk parsing otomatis!</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script>
import { getCurrentUser } from '../api/auth'

export default {
  name: 'DashboardView',
  data() {
    return {
      user: null,
      todayStats: {
        totalOrders: 0,
        paidOrders: 0,
        unpaidOrders: 0,
        totalAmount: 0
      }
    }
  },
  computed: {
    formattedDate() {
      const options = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' }
      return new Date().toLocaleDateString('id-ID', options)
    },
    userName() {
      return this.user?.name || 'User'
    }
  },
  mounted() {
    this.user = getCurrentUser()
    this.loadTodayStats()
  },
  methods: {
    loadTodayStats() {
      // Load from local storage for now
      const today = new Date().toISOString().split('T')[0]
      const savedData = localStorage.getItem(`listmak_${today}`)
      
      if (savedData) {
        const orders = JSON.parse(savedData)
        this.todayStats.totalOrders = orders.length
        this.todayStats.paidOrders = orders.filter(o => o.paid).length
        this.todayStats.unpaidOrders = orders.filter(o => !o.paid).length
        this.todayStats.totalAmount = orders.reduce((sum, o) => sum + (o.price * o.qty), 0)
      }
    },
    formatCurrency(value) {
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
.dashboard-container {
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

.page-date {
  font-size: 0.8125rem;
  color: #64748b;
}

/* Welcome Section */
.welcome-card {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border-radius: 1rem;
  padding: 1.25rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.welcome-text h2 {
  font-size: 1.125rem;
  font-weight: 600;
  color: white;
  margin-bottom: 0.25rem;
}

.welcome-text p {
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.8);
}

.welcome-icon {
  width: 48px;
  height: 48px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.welcome-icon i {
  font-size: 1.5rem;
  color: white;
}

/* Section */
.section-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.75rem;
}

/* Stats Grid */
.stats-section {
  margin-bottom: 1.5rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.stat-card {
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.875rem;
  padding: 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: 0.625rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
}

.stat-blue .stat-icon {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.stat-green .stat-icon {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.stat-yellow .stat-icon {
  background: rgba(234, 179, 8, 0.15);
  color: #eab308;
}

.stat-purple .stat-icon {
  background: rgba(168, 85, 247, 0.15);
  color: #a855f7;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: #f1f5f9;
  line-height: 1.2;
}

.stat-label {
  font-size: 0.6875rem;
  color: #64748b;
}

/* Actions Grid */
.actions-section {
  margin-bottom: 1.5rem;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
}

.action-card {
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.875rem;
  padding: 1rem 0.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  text-decoration: none;
  border: 1px solid rgba(255, 255, 255, 0.05);
  transition: all 0.2s;
}

.action-card:hover {
  background: rgba(30, 41, 59, 0.8);
  transform: translateY(-2px);
}

.action-icon {
  width: 44px;
  height: 44px;
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.action-blue {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.action-green {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.action-yellow {
  background: rgba(234, 179, 8, 0.15);
  color: #eab308;
}

.action-label {
  font-size: 0.6875rem;
  color: #94a3b8;
  text-align: center;
}

/* Info Section */
.info-card {
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 0.75rem;
  padding: 1rem;
  display: flex;
  gap: 0.75rem;
}

.info-card > i {
  color: #3b82f6;
  font-size: 1.25rem;
  flex-shrink: 0;
}

.info-content h4 {
  font-size: 0.8125rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.25rem;
}

.info-content p {
  font-size: 0.75rem;
  color: #94a3b8;
  line-height: 1.4;
}

@media (min-width: 768px) {
  .dashboard-container {
    padding: 1.5rem 2rem;
  }
  
  .stats-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}
</style>
