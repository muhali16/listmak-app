<template>
  <div class="profile-container">
    <!-- Header -->
    <header class="page-header">
      <h1 class="page-title">Profil</h1>
      <p class="page-subtitle">Informasi akun Anda</p>
    </header>

    <!-- Profile Card -->
    <section class="profile-section">
      <div class="profile-card">
        <div class="profile-avatar">
          <img 
            v-if="user && user.avatar" 
            :src="user.avatar" 
            :alt="user.name"
            class="avatar-image"
            referrerpolicy="no-referrer"
          />
          <div v-else class="avatar-placeholder">
            {{ userInitials }}
          </div>
        </div>
        
        <h2 class="profile-name">{{ userName }}</h2>
        <p class="profile-email">{{ userEmail }}</p>
        
        <div class="profile-badge">
          <i class="pi pi-verified"></i>
          <span>{{ userRole }}</span>
        </div>
      </div>
    </section>

    <!-- Info Cards -->
    <section class="info-section">
      <div class="info-card">
        <div class="info-icon">
          <i class="pi pi-envelope"></i>
        </div>
        <div class="info-content">
          <span class="info-label">Email</span>
          <span class="info-value">{{ userEmail }}</span>
        </div>
      </div>

      <div class="info-card">
        <div class="info-icon">
          <i class="pi pi-calendar"></i>
        </div>
        <div class="info-content">
          <span class="info-label">Bergabung</span>
          <span class="info-value">{{ joinedDate }}</span>
        </div>
      </div>

      <div class="info-card">
        <div class="info-icon">
          <i class="pi pi-shield"></i>
        </div>
        <div class="info-content">
          <span class="info-label">Role</span>
          <span class="info-value">{{ userRole }}</span>
        </div>
      </div>
    </section>

    <!-- Stats Section -->
    <section class="stats-section">
      <h3 class="section-title">Statistik</h3>
      <div class="stats-grid">
        <div class="stat-card">
          <span class="stat-value">{{ totalOrders }}</span>
          <span class="stat-label">Total Order</span>
        </div>
        <div class="stat-card">
          <span class="stat-value">{{ totalContacts }}</span>
          <span class="stat-label">Kontak</span>
        </div>
      </div>
    </section>

    <!-- Actions -->
    <section class="actions-section">
      <Button 
        label="Logout" 
        icon="pi pi-sign-out"
        severity="danger"
        outlined
        @click="handleLogout"
        class="logout-btn"
      />
    </section>

    <!-- App Info -->
    <section class="app-info">
      <p class="app-version">ListMak v{{ appVersion }}</p>
      <p class="app-copyright">&copy; {{ new Date().getFullYear() }} ListMak. All rights reserved.</p>
    </section>
  </div>
</template>

<script>
import Button from 'primevue/button'
import { getCurrentUser, logout } from '../api/auth'
import changelog from '../data/changelog.json'

export default {
  name: 'ProfileView',
  components: {
    Button
  },
  data() {
    return {
      user: null,
      totalOrders: 0,
      totalContacts: 0,
      appVersion: changelog[0]?.version ?? '-'
    }
  },
  computed: {
    userName() {
      return this.user?.name || 'User'
    },
    userEmail() {
      return this.user?.email || 'user@email.com'
    },
    userRole() {
      return this.user?.role || 'User'
    },
    userInitials() {
      if (!this.user?.name) return 'U'
      return this.user.name
        .split(' ')
        .map(w => w[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    },
    joinedDate() {
      if (!this.user?.created_at) return '-'
      return new Date(this.user.created_at).toLocaleDateString('id-ID', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
      })
    }
  },
  mounted() {
    this.user = getCurrentUser()
    this.loadStats()
  },
  methods: {
    loadStats() {
      // Count orders from all days
      let orderCount = 0
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i)
        if (key.startsWith('listmak_2')) {
          const data = JSON.parse(localStorage.getItem(key))
          orderCount += data.length
        }
      }
      this.totalOrders = orderCount
      
      // Count contacts
      const contacts = localStorage.getItem('listmak_contacts')
      this.totalContacts = contacts ? JSON.parse(contacts).length : 0
    },
    async handleLogout() {
      const result = await logout()
      
      this.$toast.add({
        severity: result.success ? 'success' : 'error',
        summary: result.success ? 'Logout Berhasil' : 'Error',
        detail: result.message,
        life: 3000
      })
      
      if (result.success) {
        this.$router.push('/')
      }
    }
  }
}
</script>

<style scoped>
.profile-container {
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

/* Profile Card */
.profile-section {
  margin-bottom: 1.5rem;
}

.profile-card {
  background: linear-gradient(145deg, #1e293b, #0f172a);
  border-radius: 1rem;
  padding: 1.5rem;
  text-align: center;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.profile-avatar {
  width: 80px;
  height: 80px;
  margin: 0 auto 1rem;
  border-radius: 50%;
  overflow: hidden;
  border: 3px solid #3b82f6;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
  color: white;
}

.profile-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 0.25rem;
}

.profile-email {
  font-size: 0.875rem;
  color: #64748b;
  margin-bottom: 0.75rem;
}

.profile-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background: rgba(59, 130, 246, 0.15);
  border-radius: 1rem;
  color: #3b82f6;
  font-size: 0.75rem;
  font-weight: 500;
}

/* Info Cards */
.info-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.info-card {
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.75rem;
  padding: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.info-icon {
  width: 40px;
  height: 40px;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #3b82f6;
  flex-shrink: 0;
}

.info-content {
  display: flex;
  flex-direction: column;
}

.info-label {
  font-size: 0.6875rem;
  color: #64748b;
}

.info-value {
  font-size: 0.875rem;
  color: #f1f5f9;
  font-weight: 500;
}

/* Stats */
.section-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.75rem;
}

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
  border-radius: 0.75rem;
  padding: 1rem;
  text-align: center;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.stat-value {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  color: #3b82f6;
  margin-bottom: 0.25rem;
}

.stat-label {
  font-size: 0.75rem;
  color: #64748b;
}

/* Actions */
.actions-section {
  margin-bottom: 2rem;
}

.logout-btn {
  width: 100%;
}

/* App Info */
.app-info {
  text-align: center;
}

.app-version {
  font-size: 0.75rem;
  color: #64748b;
  margin-bottom: 0.25rem;
}

.app-copyright {
  font-size: 0.6875rem;
  color: #475569;
}

@media (min-width: 768px) {
  .profile-container {
    padding: 1.5rem 2rem;
  }
}
</style>
