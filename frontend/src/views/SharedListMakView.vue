<template>
  <div class="shared-container">
    <!-- Loading -->
    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat...</p>
    </div>

    <!-- Expired -->
    <div v-else-if="isExpired" class="state-block">
      <div class="state-icon state-icon--red">
        <i class="pi pi-clock"></i>
      </div>
      <h2 class="state-title">Waktu Habis</h2>
      <p class="state-desc">
        Batas waktu berakhir pada <strong>{{ formatDateTime(listmakData.expiresAt) }}</strong>
      </p>
      <div class="info-box">
        <i class="pi pi-info-circle"></i>
        <span>Hubungi OB untuk link baru.</span>
      </div>
    </div>

    <!-- Not Found -->
    <div v-else-if="notFound" class="state-block">
      <div class="state-icon state-icon--yellow">
        <i class="pi pi-search"></i>
      </div>
      <h2 class="state-title">Link Tidak Valid</h2>
      <p class="state-desc">Link tidak ditemukan atau sudah dihapus.</p>
      <div class="info-box">
        <i class="pi pi-info-circle"></i>
        <span>Minta link baru dari OB.</span>
      </div>
    </div>

    <!-- Active -->
    <div v-else class="active-state">
      <!-- Header -->
      <div class="page-header">
        <span class="page-badge">Input Pesanan</span>
        <h1 class="page-title">{{ listmakData.title }}</h1>
      </div>

      <!-- Countdown -->
      <div class="countdown-card" :class="{ 'countdown-card--urgent': isUrgent }">
        <span class="countdown-label">Sisa waktu</span>
        <div class="countdown-display">
          <div class="countdown-unit">
            <span class="countdown-val">{{ countdown.hours }}</span>
            <span class="countdown-unit-label">jam</span>
          </div>
          <span class="countdown-sep">:</span>
          <div class="countdown-unit">
            <span class="countdown-val">{{ countdown.minutes }}</span>
            <span class="countdown-unit-label">menit</span>
          </div>
          <span class="countdown-sep">:</span>
          <div class="countdown-unit">
            <span class="countdown-val">{{ countdown.seconds }}</span>
            <span class="countdown-unit-label">detik</span>
          </div>
        </div>
        <p class="countdown-deadline">Batas: {{ formatDateTime(listmakData.expiresAt) }}</p>
      </div>

      <!-- Mode toggle -->
      <div class="mode-row">
        <button
          class="mode-btn"
          :class="{ 'mode-btn--active': inputMode === 'single' }"
          @click="inputMode = 'single'"
        >
          <i class="pi pi-user"></i>
          <span>Satu pesanan</span>
        </button>
        <button
          class="mode-btn"
          :class="{ 'mode-btn--active': inputMode === 'bulk' }"
          @click="inputMode = 'bulk'"
        >
          <i class="pi pi-list"></i>
          <span>Banyak sekaligus</span>
        </button>
      </div>

      <!-- Single form -->
      <div v-if="inputMode === 'single'" class="form-card">
        <div class="form-group">
          <label class="form-label">Nama</label>
          <input
            v-model="singleForm.name"
            class="form-input"
            type="text"
            placeholder="Nama kamu"
            :disabled="submitting"
          />
        </div>
        <div class="form-group" style="position: relative;">
          <label class="form-label">Pesanan</label>
          <textarea
            v-model="singleForm.order_detail"
            class="form-textarea"
            rows="3"
            placeholder="Contoh: Nasi Ayam Madura Paha - Warung Bu Sari"
            :disabled="submitting"
            @input="onOrderDetailInput"
            @focus="onOrderDetailFocus"
            @blur="onOrderDetailBlur"
          ></textarea>

          <ul v-if="showSuggestions && suggestions.length" class="suggestions-dropdown">
            <li
              v-for="item in suggestions"
              :key="item"
              class="suggestion-item"
              @mousedown.prevent="selectSuggestion(item)"
            >
              {{ item }}
            </li>
          </ul>

          <p v-if="orderDetailNudge" class="nudge-hint">
            <i class="pi pi-info-circle"></i>
            Tulis juga belinya dimana, contoh: Ayam Goreng - Warung Bu Sari
          </p>
        </div>

        <div v-if="submitSuccess" class="success-banner">
          <i class="pi pi-check-circle"></i>
          <span>Pesanan terkirim!</span>
        </div>
        <p v-if="submitError" class="form-error">{{ submitError }}</p>

        <button
          class="submit-btn"
          :disabled="submitting || !singleForm.name.trim() || !singleForm.order_detail.trim()"
          @click="submitSingle"
        >
          <i v-if="submitting" class="pi pi-spin pi-spinner"></i>
          <i v-else class="pi pi-sparkles"></i>
          <span>{{ submitting ? 'AI lagi parsing...' : 'Proses Pesanan' }}</span>
        </button>
      </div>

      <!-- Bulk form -->
      <div v-else class="form-card">
        <p class="bulk-hint">
          Satu baris = satu pesanan.<br>
          Format: <code>Nama: Pesanan</code> atau <code>1. Nama: Pesanan</code>
        </p>
        <textarea
          v-model="bulkInput"
          class="form-textarea bulk-textarea"
          rows="8"
          placeholder="Budi: Nasi Ayam Madura Paha&#10;Ani: Nasi Ayam Madura Dada&#10;Citra: Nasi Goreng Spesial"
          :disabled="submitting"
        ></textarea>

        <div v-if="parsedBulkOrders.length > 0" class="bulk-preview">
          <p class="bulk-preview-label">{{ parsedBulkOrders.length }} pesanan terdeteksi:</p>
          <ul class="bulk-preview-list">
            <li v-for="(o, i) in bulkOrdersWithWarnings" :key="i" class="bulk-preview-item">
              <span class="preview-name">{{ o.name }}</span>
              <span class="preview-detail">{{ o.order_detail }}</span>
              <span v-if="!o.hasLocation" class="location-warning">⚠️ Belum tulis lokasi</span>
            </li>
          </ul>
        </div>

        <div v-if="submitSuccess" class="success-banner">
          <i class="pi pi-check-circle"></i>
          <span>{{ submitSuccessMsg }}</span>
        </div>
        <p v-if="submitError" class="form-error">{{ submitError }}</p>

        <button
          class="submit-btn"
          :disabled="submitting || parsedBulkOrders.length === 0"
          @click="submitBulk"
        >
          <i v-if="submitting" class="pi pi-spin pi-spinner"></i>
          <i v-else class="pi pi-sparkles"></i>
          <span>{{ submitting ? 'AI lagi parsing...' : `Proses ${parsedBulkOrders.length} pesanan` }}</span>
        </button>
      </div>

      <!-- AI Parse confirm modal -->
      <div v-if="showConfirmModal" class="confirm-overlay">
        <div class="confirm-sheet">
          <div class="sheet-handle"></div>
          <div class="confirm-header">
            <h2 class="confirm-title">Konfirmasi Pesanan</h2>
            <p class="confirm-subtitle">AI sudah parsing pesananmu. Cek & edit harga sebelum kirim.</p>
          </div>

          <div v-if="confirmItems.length === 0" class="confirm-empty">
            Tidak ada pesanan terdeteksi.
          </div>

          <div v-else class="confirm-list">
            <div
              v-for="(item, i) in confirmItems"
              :key="i"
              class="confirm-item"
            >
              <div class="confirm-item-info">
                <span class="confirm-item-name">{{ item.name }}</span>
                <span class="confirm-item-detail">{{ item.order_detail }}</span>
                <span v-if="item.vendor_name" class="confirm-item-vendor">
                  <i class="pi pi-map-marker"></i> {{ item.vendor_name }}
                </span>
              </div>
              <div class="confirm-item-price">
                <span v-if="item.qty > 1" class="confirm-item-qty">{{ item.qty }}x</span>
                <span class="confirm-rp">Rp</span>
                <input
                  v-model.number="item.editPrice"
                  type="number"
                  class="confirm-price-input"
                  placeholder="0"
                  min="0"
                  :disabled="confirmingSubmit"
                />
                <button
                  class="confirm-delete-btn"
                  :disabled="confirmingSubmit"
                  @click="removeConfirmItem(i)"
                  title="Hapus item ini"
                >
                  <i class="pi pi-times"></i>
                </button>
              </div>
            </div>
          </div>

          <p v-if="confirmError" class="form-error">{{ confirmError }}</p>

          <button
            class="submit-btn"
            style="margin-top: 1rem"
            :disabled="confirmingSubmit || confirmItems.length === 0"
            @click="confirmSubmit"
          >
            <i v-if="confirmingSubmit" class="pi pi-spin pi-spinner"></i>
            <i v-else class="pi pi-check"></i>
            <span>{{ confirmingSubmit ? 'Mengirim...' : `Kirim ${confirmItems.length} pesanan` }}</span>
          </button>

          <button
            class="cancel-confirm-btn"
            :disabled="confirmingSubmit"
            @click="cancelConfirm"
          >
            Batalkan
          </button>
        </div>
      </div>

      <!-- Orders list -->
      <div class="orders-card">
        <div class="orders-header">
          <h3 class="orders-title">
            Pesanan masuk
            <span v-if="orders.length > 0" class="orders-badge">{{ orders.length }}</span>
          </h3>
          <button class="refresh-btn" @click="loadOrders" :disabled="loadingOrders">
            <i class="pi" :class="loadingOrders ? 'pi-spin pi-spinner' : 'pi-refresh'"></i>
          </button>
        </div>

        <div v-if="loadingOrders && orders.length === 0" class="orders-loading">
          <i class="pi pi-spin pi-spinner"></i>
        </div>

        <div v-else-if="orders.length === 0" class="orders-empty">
          <i class="pi pi-inbox"></i>
          <p>Belum ada pesanan</p>
        </div>

        <ul v-else class="orders-list">
          <li v-for="(order, i) in orders" :key="order.id" class="order-item">
            <span class="order-num">{{ i + 1 }}</span>
            <div class="order-info">
              <span class="order-name">{{ order.name }}</span>
              <span class="order-detail">{{ order.order_detail }}</span>
            </div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script>
import { share } from '../api'

export default {
  name: 'SharedListMakView',

  data() {
    return {
      loading: true,
      notFound: false,
      expiredNow: false,
      listmakData: { title: '', expiresAt: null },
      inputMode: 'single',
      singleForm: { name: '', order_detail: '' },
      bulkInput: '',
      submitting: false,
      submitSuccess: false,
      submitSuccessMsg: '',
      submitError: '',
      orders: [],
      loadingOrders: false,
      countdown: { hours: '00', minutes: '00', seconds: '00' },
      countdownInterval: null,
      suggestions: [],
      showSuggestions: false,
      suggestionTimer: null,
      orderDetailNudge: false,
      // AI confirm flow
      showConfirmModal: false,
      confirmItems: [],
      confirmError: '',
      confirmingSubmit: false,
    }
  },

  computed: {
    shareId() {
      return this.$route.params.shareId
    },
    isExpired() {
      if (this.expiredNow) return true
      if (!this.listmakData.expiresAt) return false
      return new Date() > new Date(this.listmakData.expiresAt)
    },
    isUrgent() {
      if (!this.listmakData.expiresAt) return false
      const diff = new Date(this.listmakData.expiresAt) - new Date()
      return diff > 0 && diff < 30 * 60 * 1000
    },
    parsedBulkOrders() {
      const INVISIBLE = /[​‌‍⁠﻿]/g
      return this.bulkInput
        .split('\n')
        .map(line => line.replace(INVISIBLE, '').trim())
        .filter(Boolean)
        .map(line => {
          const stripped = line.replace(/^\d+[\.\)]\s+/, '')
          const m = stripped.match(/^(.+?)[:\-]\s+(.+)$/)
          if (!m) return null
          return { name: m[1].trim(), order_detail: m[2].trim() }
        })
        .filter(Boolean)
    },
    bulkOrdersWithWarnings() {
      return this.parsedBulkOrders.map(o => ({
        ...o,
        hasLocation: this.hasLocationHint(o.order_detail)
      }))
    }
  },

  mounted() {
    this.loadShareData()
  },

  beforeUnmount() {
    if (this.countdownInterval) clearInterval(this.countdownInterval)
  },

  methods: {
    hasLocationHint(detail) {
      const keywords = /\b(pak|bu|mas|mba|mbak|warung|resto|kedai|depot|kantin|gerobak)\b/i
      const dashSeparator = /[-–—]/
      return keywords.test(detail) || dashSeparator.test(detail)
    },

    async loadSuggestions(query) {
      try {
        const res = await share.getFoodSuggestions(this.shareId, query)
        this.suggestions = res.data || []
        this.showSuggestions = this.suggestions.length > 0
      } catch {
        this.showSuggestions = false
      }
    },

    onOrderDetailInput() {
      clearTimeout(this.suggestionTimer)
      const val = this.singleForm.order_detail
      this.orderDetailNudge = false
      if (val.length < 2) {
        this.showSuggestions = false
        return
      }
      this.suggestionTimer = setTimeout(() => this.loadSuggestions(val), 300)
    },

    onOrderDetailFocus() {
      if (this.singleForm.order_detail.length < 2) {
        this.loadSuggestions('')
      }
    },

    onOrderDetailBlur() {
      setTimeout(() => { this.showSuggestions = false }, 200)
      const val = this.singleForm.order_detail.trim()
      this.orderDetailNudge = val.length > 0 && !this.hasLocationHint(val)
    },

    selectSuggestion(item) {
      this.singleForm.order_detail = item
      this.showSuggestions = false
      this.orderDetailNudge = !this.hasLocationHint(item)
    },

    async loadShareData() {
      try {
        const res = await share.getShareLink(this.shareId)
        if (res.success && res.data) {
          this.listmakData = { title: res.data.title, expiresAt: res.data.expires_at }
          if (res.data.is_expired) {
            this.expiredNow = true
          } else {
            this.startCountdown()
            this.loadOrders()
          }
        } else {
          this.notFound = true
        }
      } catch (err) {
        if (err.code === 'EXPIRED') {
          this.listmakData.expiresAt = err.data?.expires_at
          this.expiredNow = true
        } else {
          this.notFound = true
        }
      } finally {
        this.loading = false
      }
    },

    async loadOrders() {
      this.loadingOrders = true
      try {
        const res = await share.getShareOrders(this.shareId)
        if (res.success) this.orders = res.data || []
      } catch {
        // secondary display — fail silently
      } finally {
        this.loadingOrders = false
      }
    },

    startCountdown() {
      this.updateCountdown()
      this.countdownInterval = setInterval(() => {
        this.updateCountdown()
        if (this.isExpired) clearInterval(this.countdownInterval)
      }, 1000)
    },

    updateCountdown() {
      const diff = new Date(this.listmakData.expiresAt) - new Date()
      if (diff <= 0) {
        this.countdown = { hours: '00', minutes: '00', seconds: '00' }
        this.expiredNow = true
        return
      }
      this.countdown = {
        hours: String(Math.floor(diff / 3600000)).padStart(2, '0'),
        minutes: String(Math.floor((diff % 3600000) / 60000)).padStart(2, '0'),
        seconds: String(Math.floor((diff % 60000) / 1000)).padStart(2, '0')
      }
    },

    formatDateTime(dateString) {
      if (!dateString) return '-'
      return new Date(dateString).toLocaleString('id-ID', {
        weekday: 'short', day: 'numeric', month: 'short',
        year: 'numeric', hour: '2-digit', minute: '2-digit'
      })
    },

    async submitSingle() {
      const name = this.singleForm.name.trim()
      const order_detail = this.singleForm.order_detail.trim()
      if (!name || !order_detail) return
      this.submitting = true
      this.submitError = ''
      this.submitSuccess = false
      try {
        const res = await share.parseOrders(this.shareId, [{ name, order_detail }])
        const items = (res.data || []).map(item => ({ ...item, editPrice: item.estimated_price || 0 }))
        if (items.length === 0) {
          this.submitError = 'AI tidak bisa memproses pesanan ini. Coba tulis lebih jelas.'
          return
        }
        this.confirmItems = items
        this.confirmError = ''
        this.showConfirmModal = true
      } catch (err) {
        this.submitError = err.message || 'Gagal proses pesanan. Coba lagi.'
      } finally {
        this.submitting = false
      }
    },

    async submitBulk() {
      if (!this.parsedBulkOrders.length) return
      this.submitting = true
      this.submitError = ''
      this.submitSuccess = false
      try {
        const res = await share.parseOrders(this.shareId, this.parsedBulkOrders)
        const items = (res.data || []).map(item => ({ ...item, editPrice: item.estimated_price || 0 }))
        if (items.length === 0) {
          this.submitError = 'AI tidak bisa memproses pesanan. Coba tulis ulang.'
          return
        }
        this.confirmItems = items
        this.confirmError = ''
        this.showConfirmModal = true
      } catch (err) {
        this.submitError = err.message || 'Gagal proses pesanan. Coba lagi.'
      } finally {
        this.submitting = false
      }
    },

    removeConfirmItem(index) {
      this.confirmItems.splice(index, 1)
    },

    cancelConfirm() {
      this.showConfirmModal = false
      this.confirmItems = []
      this.confirmError = ''
    },

    async confirmSubmit() {
      this.confirmingSubmit = true
      this.confirmError = ''
      try {
        const orders = this.confirmItems.map(item => ({
          name: item.name,
          order_detail: item.order_detail,
          vendor_name: item.vendor_name || '',
          price: item.editPrice || 0,
          qty: item.qty || 1,
        }))
        const res = await share.submitShareOrder(this.shareId, { orders })
        const count = res.data?.added_count || orders.length
        this.showConfirmModal = false
        this.confirmItems = []
        this.singleForm = { name: '', order_detail: '' }
        this.bulkInput = ''
        this.submitSuccessMsg = `${count} pesanan berhasil dikirim!`
        this.submitSuccess = true
        await this.loadOrders()
        setTimeout(() => { this.submitSuccess = false }, 4000)
      } catch (err) {
        this.confirmError = err.message || 'Gagal kirim. Coba lagi.'
      } finally {
        this.confirmingSubmit = false
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

.state-icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.5rem;
}

.state-icon i {
  font-size: 2.25rem;
}

.state-icon--red {
  background: rgba(239, 68, 68, 0.15);
}
.state-icon--red i { color: #ef4444; }

.state-icon--yellow {
  background: rgba(234, 179, 8, 0.15);
}
.state-icon--yellow i { color: #eab308; }

.state-title {
  font-size: 1.375rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}

.state-desc {
  font-size: 0.9375rem;
  color: #94a3b8;
  max-width: 300px;
  line-height: 1.5;
}

.state-desc strong { color: #cbd5e1; }

.info-box {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  background: rgba(59, 130, 246, 0.08);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 0.75rem;
  padding: 0.875rem 1rem;
  text-align: left;
  max-width: 320px;
  margin-top: 0.5rem;
}

.info-box i {
  color: #3b82f6;
  font-size: 1rem;
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.info-box span {
  font-size: 0.8125rem;
  color: #94a3b8;
  line-height: 1.5;
}

/* Active state */
.active-state {
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
  color: #818cf8;
  background: rgba(99, 102, 241, 0.12);
  border: 1px solid rgba(99, 102, 241, 0.25);
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

/* Countdown */
.countdown-card {
  background: rgba(30, 41, 59, 0.7);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 0.875rem;
  padding: 1rem 1.25rem;
  margin-bottom: 1rem;
  text-align: center;
}

.countdown-card--urgent {
  border-color: rgba(239, 68, 68, 0.35);
  background: rgba(239, 68, 68, 0.06);
}

.countdown-label {
  display: block;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #64748b;
  margin-bottom: 0.625rem;
}

.countdown-display {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  margin-bottom: 0.5rem;
}

.countdown-unit {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.125rem;
}

.countdown-val {
  font-size: 1.75rem;
  font-weight: 700;
  color: #f1f5f9;
  background: rgba(15, 23, 42, 0.5);
  border-radius: 0.5rem;
  padding: 0.25rem 0.5rem;
  min-width: 52px;
  line-height: 1;
}

.countdown-card--urgent .countdown-val {
  color: #ef4444;
}

.countdown-unit-label {
  font-size: 0.5625rem;
  color: #64748b;
  text-transform: uppercase;
}

.countdown-sep {
  font-size: 1.5rem;
  font-weight: 700;
  color: #475569;
  margin-bottom: 1rem;
}

.countdown-deadline {
  font-size: 0.75rem;
  color: #64748b;
  margin: 0;
}

/* Mode toggle */
.mode-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.mode-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.75rem 0.5rem;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 0.75rem;
  color: #64748b;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}

.mode-btn--active {
  background: rgba(99, 102, 241, 0.15);
  border-color: rgba(99, 102, 241, 0.35);
  color: #818cf8;
}

/* Form card */
.form-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.875rem;
  padding: 1.125rem;
  margin-bottom: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #94a3b8;
  margin-bottom: 0.375rem;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 0.625rem 0.75rem;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.625rem;
  color: #f1f5f9;
  font-size: 0.9375rem;
  font-family: inherit;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: rgba(99, 102, 241, 0.5);
}

.form-input::placeholder,
.form-textarea::placeholder {
  color: #334155;
}

.form-input:disabled,
.form-textarea:disabled {
  opacity: 0.5;
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

/* Bulk */
.bulk-hint {
  font-size: 0.8125rem;
  color: #64748b;
  margin-bottom: 0.75rem;
  line-height: 1.6;
}

.bulk-hint code {
  background: rgba(255, 255, 255, 0.07);
  padding: 0.125rem 0.35rem;
  border-radius: 0.3rem;
  color: #94a3b8;
  font-size: 0.8125rem;
}

.bulk-textarea {
  min-height: 160px;
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 0.8125rem;
  line-height: 1.6;
}

.bulk-preview {
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.625rem;
  padding: 0.75rem;
  margin: 0.75rem 0;
}

.bulk-preview-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #22c55e;
  margin-bottom: 0.5rem;
}

.bulk-preview-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.bulk-preview-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
}

.preview-name {
  font-weight: 600;
  color: #e2e8f0;
  min-width: 5rem;
  flex-shrink: 0;
}

.preview-detail {
  color: #94a3b8;
  overflow-wrap: anywhere;
}

/* Success / error */
.success-banner {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.2);
  border-radius: 0.625rem;
  padding: 0.75rem 0.875rem;
  margin-bottom: 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: #22c55e;
}

.form-error {
  font-size: 0.8125rem;
  color: #ef4444;
  margin-bottom: 0.75rem;
}

/* Submit button */
.submit-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.875rem;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border: none;
  border-radius: 0.875rem;
  color: white;
  font-size: 0.9375rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.submit-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Orders card */
.orders-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.875rem;
  padding: 1.125rem;
}

.orders-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.875rem;
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

.orders-badge {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
  font-size: 0.6875rem;
  font-weight: 700;
  padding: 0.125rem 0.45rem;
  border-radius: 999px;
}

.refresh-btn {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 0.5rem;
  color: #64748b;
  cursor: pointer;
  transition: color 0.15s;
}

.refresh-btn:hover:not(:disabled) { color: #94a3b8; }
.refresh-btn:disabled { opacity: 0.5; cursor: not-allowed; }

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

.orders-empty i { font-size: 2rem; }

.orders-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.order-item {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  background: rgba(15, 23, 42, 0.4);
  border-radius: 0.5rem;
  padding: 0.5rem 0.625rem;
}

.order-num {
  font-size: 0.6875rem;
  font-weight: 700;
  color: #64748b;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 0.25rem;
  padding: 0.1rem 0.35rem;
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.order-info {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
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
  overflow-wrap: anywhere;
}

.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #1e293b;
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 0.625rem;
  list-style: none;
  margin: 0.25rem 0 0;
  padding: 0.25rem 0;
  z-index: 100;
  max-height: 220px;
  overflow-y: auto;
  box-shadow: 0 8px 24px rgba(0,0,0,0.4);
}

.suggestion-item {
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  color: #e2e8f0;
  cursor: pointer;
  transition: background 0.1s;
}

.suggestion-item:hover {
  background: rgba(99, 102, 241, 0.15);
}

.nudge-hint {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: #94a3b8;
  margin-top: 0.375rem;
}

.nudge-hint i {
  color: #60a5fa;
  font-size: 0.75rem;
  flex-shrink: 0;
}

.location-warning {
  font-size: 0.6875rem;
  color: #f59e0b;
  margin-left: auto;
  flex-shrink: 0;
}

/* Confirm modal */
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  z-index: 200;
  display: flex;
  align-items: flex-end;
}

.confirm-sheet {
  width: 100%;
  max-width: 560px;
  margin: 0 auto;
  background: #0f172a;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 1.25rem 1.25rem 0 0;
  padding: 0 1.125rem 2rem;
  max-height: 90dvh;
  overflow-y: auto;
}

.sheet-handle {
  width: 2.5rem;
  height: 0.25rem;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 999px;
  margin: 0.75rem auto 1rem;
}

.confirm-header {
  margin-bottom: 1rem;
}

.confirm-title {
  font-size: 1.0625rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0 0 0.25rem;
}

.confirm-subtitle {
  font-size: 0.8125rem;
  color: #64748b;
  margin: 0;
}

.confirm-empty {
  color: #64748b;
  font-size: 0.875rem;
  text-align: center;
  padding: 1rem 0;
}

.confirm-list {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  max-height: 50dvh;
  overflow-y: auto;
}

.confirm-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.75rem;
  padding: 0.625rem 0.75rem;
}

.confirm-item-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
  flex: 1;
}

.confirm-item-name {
  font-size: 0.8125rem;
  font-weight: 700;
  color: #f1f5f9;
}

.confirm-item-detail {
  font-size: 0.8125rem;
  color: #cbd5e1;
  overflow-wrap: anywhere;
}

.confirm-item-vendor {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.6875rem;
  color: #64748b;
}

.confirm-item-vendor i {
  font-size: 0.625rem;
}

.confirm-item-price {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  flex-shrink: 0;
}

.confirm-item-qty {
  font-size: 0.75rem;
  font-weight: 600;
  color: #94a3b8;
  background: rgba(148, 163, 184, 0.1);
  border-radius: 0.25rem;
  padding: 0.1rem 0.3rem;
  flex-shrink: 0;
}

.confirm-rp {
  font-size: 0.75rem;
  color: #64748b;
  flex-shrink: 0;
}

.confirm-price-input {
  width: 5rem;
  padding: 0.3rem 0.4rem;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.375rem;
  color: #f1f5f9;
  font-size: 0.8125rem;
  text-align: right;
}

.confirm-price-input:focus {
  outline: none;
  border-color: rgba(99, 102, 241, 0.5);
}

.cancel-confirm-btn {
  width: 100%;
  margin-top: 0.5rem;
  padding: 0.75rem;
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.875rem;
  color: #64748b;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
}

.cancel-confirm-btn:disabled { opacity: 0.5; }

.confirm-delete-btn {
  width: 1.75rem;
  height: 1.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 0.375rem;
  color: #ef4444;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s;
}

.confirm-delete-btn:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.2);
}

.confirm-delete-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.confirm-delete-btn i {
  font-size: 0.6875rem;
}
</style>
