<template>
  <div class="shared-container">
    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat...</p>
    </div>

    <!-- Expired State -->
    <div v-else-if="isExpired" class="expired-state">
      <div class="expired-icon">
        <i class="pi pi-clock"></i>
      </div>
      <h2>Waktu Habis</h2>
      <p class="expired-message">
        Waktu untuk input pesanan telah berakhir pada 
        <strong>{{ formatDateTime(listmakData.expiresAt) }}</strong>
      </p>
      <div class="expired-info">
        <i class="pi pi-info-circle"></i>
        <span>Silakan hubungi OB (Office Boy) untuk informasi ketersediaan ListMak tambahan.</span>
      </div>
    </div>

    <!-- Not Found State -->
    <div v-else-if="notFound" class="expired-state">
      <div class="expired-icon not-found-icon">
        <i class="pi pi-search"></i>
      </div>
      <h2>ListMak Tidak Ditemukan</h2>
      <p class="expired-message">
        Link yang Anda akses tidak valid atau sudah dihapus.
      </p>
      <div class="expired-info">
        <i class="pi pi-info-circle"></i>
        <span>Silakan minta link baru dari OB (Office Boy).</span>
      </div>
    </div>

    <!-- Active Input State -->
    <div v-else class="input-state">
      <!-- Header -->
      <header class="shared-header">
        <h1 class="shared-title">Input Pesanan</h1>
        <p class="shared-subtitle">{{ listmakData.title || 'ListMak Hari Ini' }}</p>
      </header>

      <!-- Countdown Timer -->
      <div class="countdown-card">
        <div class="countdown-label">
          <i class="pi pi-clock"></i>
          <span>Waktu Tersisa</span>
        </div>
        <div class="countdown-timer">
          <div class="countdown-unit">
            <span class="countdown-value">{{ countdown.hours }}</span>
            <span class="countdown-text">Jam</span>
          </div>
          <span class="countdown-separator">:</span>
          <div class="countdown-unit">
            <span class="countdown-value">{{ countdown.minutes }}</span>
            <span class="countdown-text">Menit</span>
          </div>
          <span class="countdown-separator">:</span>
          <div class="countdown-unit">
            <span class="countdown-value">{{ countdown.seconds }}</span>
            <span class="countdown-text">Detik</span>
          </div>
        </div>
        <p class="countdown-deadline">
          Batas waktu: <strong>{{ formatDateTime(listmakData.expiresAt) }}</strong>
        </p>
      </div>

      <!-- Two-column layout: Form | Orders -->
      <div class="content-grid">
        <!-- Left column: form -->
        <div class="form-column">
          <!-- Input Mode Tabs -->
          <div class="input-tabs">
            <button
              class="input-tab"
              :class="{ active: inputMode === 'single' }"
              @click="inputMode = 'single'"
            >
              <i class="pi pi-user"></i>
              <span>Satu Pesanan</span>
            </button>
            <button
              class="input-tab"
              :class="{ active: inputMode === 'bulk' }"
              @click="inputMode = 'bulk'"
            >
              <i class="pi pi-users"></i>
              <span>Banyak Pesanan</span>
            </button>
          </div>

          <!-- Single Input Form -->
          <div v-if="inputMode === 'single'" class="order-form">
            <div class="form-group">
              <label>Nama Anda</label>
              <input
                type="text"
                v-model="orderForm.name"
                placeholder="Masukkan nama Anda"
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Pesanan</label>
              <textarea
                v-model="orderForm.order"
                placeholder="Contoh: Nasi Ayam Madura PAHA"
                rows="3"
                class="form-textarea"
              ></textarea>
            </div>
            <Button
              @click="submitOrder"
              label="Kirim Pesanan"
              icon="pi pi-send"
              class="submit-btn"
              :loading="submitting"
              :disabled="!orderForm.name.trim() || !orderForm.order.trim()"
            />
          </div>

          <!-- Bulk Input Form -->
          <div v-else class="order-form">
            <div class="form-group">
              <label>Daftar Pesanan</label>
              <p class="form-hint">Format: <strong>Nama: Pesanan</strong> (satu per baris)</p>
              <textarea
                v-model="bulkInput"
                placeholder="Budi: Nasi Ayam Madura PAHA
Ani: Nasi Ayam Madura DADA
Citra: Nasi Goreng Spesial"
                rows="8"
                class="form-textarea bulk-textarea"
              ></textarea>
            </div>
            <div class="bulk-preview" v-if="parsedBulkOrders.length > 0">
              <p class="preview-label">Preview: {{ parsedBulkOrders.length }} pesanan</p>
              <ul class="preview-list">
                <li v-for="(order, index) in parsedBulkOrders.slice(0, 5)" :key="index">
                  <strong>{{ order.name }}</strong>: {{ order.order }}
                </li>
                <li v-if="parsedBulkOrders.length > 5" class="preview-more">
                  ... dan {{ parsedBulkOrders.length - 5 }} pesanan lainnya
                </li>
              </ul>
            </div>
            <Button
              @click="submitBulkOrders"
              :label="`Kirim ${parsedBulkOrders.length} Pesanan`"
              icon="pi pi-send"
              class="submit-btn"
              :loading="submitting"
              :disabled="parsedBulkOrders.length === 0"
            />
          </div>

          <!-- Success Message -->
          <div v-if="orderSubmitted" class="success-card">
            <i class="pi pi-check-circle"></i>
            <div class="success-content">
              <h3>Pesanan Terkirim!</h3>
              <p>{{ successMessage }}</p>
            </div>
          </div>
        </div>

        <!-- Right column: orders list -->
        <div class="orders-column">
          <div class="orders-section">
            <div class="orders-header">
              <h3 class="orders-title">
                <i class="pi pi-list"></i>
                Daftar Pesanan
                <span v-if="orders.length > 0" class="orders-count">{{ orders.length }}</span>
              </h3>
              <button class="refresh-btn" @click="loadOrders" :disabled="loadingOrders">
                <i class="pi" :class="loadingOrders ? 'pi-spin pi-spinner' : 'pi-refresh'"></i>
              </button>
            </div>

            <div v-if="loadingOrders && orders.length === 0" class="orders-loading">
              <i class="pi pi-spin pi-spinner"></i>
              <span>Memuat pesanan...</span>
            </div>

            <div v-else-if="orders.length === 0" class="orders-empty">
              <i class="pi pi-inbox"></i>
              <p>Belum ada pesanan masuk</p>
            </div>

            <ul v-else class="orders-list">
              <li v-for="(order, index) in orders" :key="order.id" class="order-item">
                <span class="order-number">{{ index + 1 }}</span>
                <div class="order-info">
                  <span class="order-name">{{ order.name }}</span>
                  <span class="order-detail">{{ order.order_detail }}</span>
                </div>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Button from 'primevue/button'
import { share } from '../api'

export default {
  name: 'SharedListMakView',
  components: {
    Button
  },
  data() {
    return {
      loading: true,
      notFound: false,
      listmakData: {
        id: null,
        title: '',
        expiresAt: null,
        createdBy: ''
      },
      orderForm: {
        name: '',
        order: ''
      },
      countdown: {
        hours: '00',
        minutes: '00',
        seconds: '00'
      },
      countdownInterval: null,
      submitting: false,
      orderSubmitted: false,
      inputMode: 'single',
      bulkInput: '',
      successMessage: '',
      expiredNow: false, // Tracks if expired during user session
      orders: [],
      loadingOrders: false
    }
  },
  computed: {
    isExpired() {
      // Check both initial expiry and runtime expiry
      if (this.expiredNow) return true
      if (!this.listmakData.expiresAt) return false
      return new Date() > new Date(this.listmakData.expiresAt)
    },
    shareId() {
      return this.$route.params.shareId
    },
    parsedBulkOrders() {
      if (!this.bulkInput.trim()) return []
      
      const lines = this.bulkInput.split('\n').filter(l => l.trim())
      const orders = []
      
      for (const line of lines) {
        // Match pattern "Nama: Pesanan" or "Nama - Pesanan"
        const match = line.match(/^(.+?)[:\-]\s*(.+)$/)
        if (match) {
          orders.push({
            name: match[1].trim(),
            order_detail: match[2].trim() // Updated to match API
          })
        }
      }
      
      return orders
    }
  },
  mounted() {
    this.loadListMakData()
  },
  beforeUnmount() {
    if (this.countdownInterval) {
      clearInterval(this.countdownInterval)
    }
  },
  methods: {
    async loadListMakData() {
      try {
        const response = await share.getShareLink(this.shareId)
        
        if (response.success && response.data) {
          this.listmakData = {
            title: response.data.title,
            expiresAt: response.data.expires_at,
            isExpired: response.data.is_expired
          }
          
          // Handle expired state from backend
          if (response.data.is_expired) {
            this.expiredNow = true
          } else {
             this.startCountdown()
             this.loadOrders()
          }
        } else {
           this.notFound = true
        }
      } catch (error) {
        console.error('Failed to load share link:', error)
        if (error.code === 'EXPIRED') {
             this.listmakData.expiresAt = error.data?.expires_at
             this.expiredNow = true
             this.loading = false
             return
        }
        if (error.code === 'NOT_FOUND' || error.message.includes('404')) {
             this.notFound = true
        } else {
            // Check for 410 Gone manually if not handled by error code
            if (error.message.includes('410') || error.message.includes('EXPIRED')) {
                 this.expiredNow = true
            } else {
                 this.notFound = true
            }
        }
      } finally {
        this.loading = false
      }
    },
    async loadOrders() {
      this.loadingOrders = true
      try {
        const response = await share.getShareOrders(this.shareId)
        if (response.success) {
          this.orders = response.data || []
        }
      } catch (error) {
        // Silently fail — orders list is a secondary display
        console.error('Failed to load orders:', error)
      } finally {
        this.loadingOrders = false
      }
    },
    startCountdown() {
      this.updateCountdown()
      this.countdownInterval = setInterval(() => {
        this.updateCountdown()
        
        // Check if expired during countdown
        if (this.isExpired) {
          clearInterval(this.countdownInterval)
        }
      }, 1000)
    },
    updateCountdown() {
      const now = new Date()
      const expires = new Date(this.listmakData.expiresAt)
      const diff = expires - now

      if (diff <= 0) {
        this.countdown = { hours: '00', minutes: '00', seconds: '00' }
        // Set expiredNow to trigger reactive update and show expired state
        this.expiredNow = true
        return
      }

      const hours = Math.floor(diff / (1000 * 60 * 60))
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
      const seconds = Math.floor((diff % (1000 * 60)) / 1000)

      this.countdown = {
        hours: hours.toString().padStart(2, '0'),
        minutes: minutes.toString().padStart(2, '0'),
        seconds: seconds.toString().padStart(2, '0')
      }
    },
    formatDateTime(dateString) {
      if (!dateString) return '-'
      const date = new Date(dateString)
      return date.toLocaleString('id-ID', {
        weekday: 'long',
        day: 'numeric',
        month: 'long',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    },
    async submitOrder() {
      if (!this.orderForm.name.trim() || !this.orderForm.order.trim()) {
        this.$toast.add({
          severity: 'warn',
          summary: 'Perhatian',
          detail: 'Mohon isi nama dan pesanan',
          life: 3000
        })
        return
      }

      this.submitting = true

      try {
        const payload = {
           name: this.orderForm.name.trim(),
           order_detail: this.orderForm.order.trim()
        }

        const response = await share.submitShareOrder(this.shareId, payload)

        if (response.success) {
            this.orderSubmitted = true
            this.successMessage = 'Pesanan Anda telah berhasil ditambahkan ke ListMak.'
            this.orderForm = { name: '', order: '' }
            this.loadOrders()

            this.$toast.add({
                severity: 'success',
                summary: 'Berhasil',
                detail: 'Pesanan Anda telah ditambahkan',
                life: 3000
            })

            setTimeout(() => {
                this.orderSubmitted = false
            }, 5000)
        }
      } catch (error) {
        console.error('Failed to submit order:', error)
         this.$toast.add({
            severity: 'error',
            summary: 'Gagal',
            detail: error.message || 'Gagal mengirim pesanan',
            life: 3000
        })
      } finally {
         this.submitting = false
      }
    },
    async submitBulkOrders() {
      if (this.parsedBulkOrders.length === 0) {
        this.$toast.add({
          severity: 'warn',
          summary: 'Perhatian',
          detail: 'Tidak ada pesanan yang valid. Gunakan format "Nama: Pesanan"',
          life: 3000
        })
        return
      }

      this.submitting = true

      try {
        const payload = {
            orders: this.parsedBulkOrders
        }

        const response = await share.submitShareOrder(this.shareId, payload)
        
        if (response.success) {
            const addedCount = this.parsedBulkOrders.length // Or response.data.added_count
            this.orderSubmitted = true
            this.successMessage = `${addedCount} pesanan telah berhasil ditambahkan ke ListMak.`
            this.bulkInput = ''
            this.loadOrders()

            this.$toast.add({
                severity: 'success',
                summary: 'Berhasil',
                detail: `${addedCount} pesanan telah ditambahkan`,
                life: 3000
            })

            setTimeout(() => {
                this.orderSubmitted = false
            }, 5000)
        }
      } catch (error) {
         console.error('Failed to submit bulk orders:', error)
         this.$toast.add({
            severity: 'error',
            summary: 'Gagal',
            detail: error.message || 'Gagal mengirim pesanan',
            life: 3000
        })
      } finally {
         this.submitting = false
      }
    }
  }
}
</script>

<style scoped>
.shared-container {
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

/* Expired State */
.expired-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  height: 100dvh;
  text-align: center;
  padding: 1.5rem;
}

.expired-icon {
  width: 100px;
  height: 100px;
  background: rgba(239, 68, 68, 0.15);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.expired-icon i {
  font-size: 3rem;
  color: #ef4444;
}

.not-found-icon {
  background: rgba(234, 179, 8, 0.15);
}

.not-found-icon i {
  color: #eab308;
}

.expired-state h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 0.5rem;
}

.expired-message {
  font-size: 0.9375rem;
  color: #94a3b8;
  margin-bottom: 1.5rem;
  max-width: 300px;
}

.expired-info {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 0.75rem;
  padding: 1rem;
  text-align: left;
  max-width: 320px;
}

.expired-info i {
  color: #3b82f6;
  font-size: 1.25rem;
  flex-shrink: 0;
}

.expired-info span {
  font-size: 0.8125rem;
  color: #94a3b8;
  line-height: 1.5;
}

/* Input State */
.input-state {
  max-width: 960px;
  margin: 0 auto;
}

/* Two-column grid */
.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  align-items: start;
}

.orders-column .orders-section {
  max-height: 70vh;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.1) transparent;
}

@media (max-width: 640px) {
  .content-grid {
    grid-template-columns: 1fr;
  }

  .orders-column .orders-section {
    max-height: none;
  }
}

.shared-header {
  text-align: center;
  padding: 1.5rem 0;
}

.shared-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 0.25rem;
}

.shared-subtitle {
  font-size: 0.875rem;
  color: #64748b;
}

/* Countdown Card */
.countdown-card {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15), rgba(139, 92, 246, 0.15));
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 1rem;
  padding: 1.25rem;
  text-align: center;
  margin-bottom: 1.5rem;
}

.countdown-label {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin-bottom: 0.75rem;
}

.countdown-label i {
  color: #3b82f6;
}

.countdown-timer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.countdown-unit {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.countdown-value {
  font-size: 2rem;
  font-weight: 700;
  color: #f1f5f9;
  background: rgba(15, 23, 42, 0.5);
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  min-width: 60px;
}

.countdown-text {
  font-size: 0.625rem;
  color: #64748b;
  text-transform: uppercase;
  margin-top: 0.25rem;
}

.countdown-separator {
  font-size: 2rem;
  font-weight: 700;
  color: #64748b;
}

.countdown-deadline {
  font-size: 0.75rem;
  color: #64748b;
}

.countdown-deadline strong {
  color: #94a3b8;
}

/* Order Form */
.order-form {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 1rem;
  padding: 1.25rem;
  margin-bottom: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #94a3b8;
  margin-bottom: 0.5rem;
}

.form-input,
.form-textarea {
  width: 100%;
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.9375rem;
  padding: 0.75rem;
  font-family: inherit;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-input::placeholder,
.form-textarea::placeholder {
  color: #64748b;
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.submit-btn {
  width: 100%;
  background: linear-gradient(135deg, #22c55e, #16a34a) !important;
  border: none !important;
  font-size: 1rem !important;
  padding: 0.875rem !important;
}

.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Success Card */
.success-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.2);
  border-radius: 0.75rem;
  padding: 1rem;
}

.success-card i {
  font-size: 2rem;
  color: #22c55e;
}

.success-content h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.25rem;
}

.success-content p {
  font-size: 0.8125rem;
  color: #94a3b8;
}

/* Input Tabs */
.input-tabs {
  display: flex;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.75rem;
  padding: 0.25rem;
  gap: 0.25rem;
  margin-bottom: 1rem;
}

.input-tab {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  padding: 0.75rem 0.5rem;
  background: transparent;
  border: none;
  color: #64748b;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.75rem;
}

.input-tab i {
  font-size: 1.25rem;
}

.input-tab.active {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.input-tab:hover:not(.active) {
  background: rgba(255, 255, 255, 0.05);
}

/* Form Hint */
.form-hint {
  font-size: 0.75rem;
  color: #64748b;
  margin-bottom: 0.5rem;
}

.form-hint strong {
  color: #94a3b8;
}

/* Bulk Textarea */
.bulk-textarea {
  min-height: 160px;
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 0.8125rem;
  line-height: 1.6;
}

/* Bulk Preview */
.bulk-preview {
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.5rem;
  padding: 0.75rem;
  margin-bottom: 1rem;
}

.preview-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #22c55e;
  margin-bottom: 0.5rem;
}

.preview-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.preview-list li {
  font-size: 0.75rem;
  color: #94a3b8;
  padding: 0.25rem 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
}

.preview-list li:last-child {
  border-bottom: none;
}

.preview-list li strong {
  color: #f1f5f9;
}

.preview-more {
  font-style: italic;
  color: #64748b !important;
}

/* Orders Section */
.orders-section {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 1rem;
  padding: 1.25rem;
  margin-bottom: 1.5rem;
}

.orders-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.orders-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9375rem;
  font-weight: 600;
  color: #f1f5f9;
  margin: 0;
}

.orders-title i {
  color: #3b82f6;
  font-size: 1rem;
}

.orders-count {
  background: rgba(59, 130, 246, 0.2);
  color: #3b82f6;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
}

.refresh-btn {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #64748b;
  padding: 0.375rem 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
  line-height: 1;
}

.refresh-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.05);
  color: #94a3b8;
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.orders-loading,
.orders-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1.5rem 0;
  color: #64748b;
  font-size: 0.875rem;
}

.orders-empty i {
  font-size: 2rem;
}

.orders-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.order-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  background: rgba(15, 23, 42, 0.4);
  border-radius: 0.5rem;
  padding: 0.625rem 0.75rem;
}

.order-number {
  font-size: 0.6875rem;
  font-weight: 700;
  color: #64748b;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 0.25rem;
  padding: 0.125rem 0.375rem;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.order-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.order-name {
  font-size: 0.8125rem;
  font-weight: 600;
  color: #f1f5f9;
}

.order-detail {
  font-size: 0.75rem;
  color: #94a3b8;
  word-break: break-word;
}
</style>
