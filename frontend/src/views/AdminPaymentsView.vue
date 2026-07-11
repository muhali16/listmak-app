<template>
  <div class="admin-container">
    <div class="admin-header">
      <h1 class="admin-title">Pembayaran</h1>
      <span class="admin-badge">Admin</span>
      <button class="info-toggle" @click="showInfo = !showInfo">
        <i class="pi" :class="showInfo ? 'pi-chevron-up' : 'pi-info-circle'"></i>
        <span>Cara kerja</span>
      </button>
    </div>

    <!-- In-app documentation: status + reconcile mechanism -->
    <div v-if="showInfo" class="info-panel">
      <div class="info-block">
        <h3 class="info-h">Status transaksi</h3>
        <ul class="info-list">
          <li><span class="status-badge status-pending">Pending</span> QR dibuat, menunggu pembayaran. Belum jadi order.</li>
          <li><span class="status-badge status-ok">Lunas</span> Dibayar &amp; diverifikasi Pakasir. Order otomatis masuk ke listmak.</li>
          <li><span class="status-badge status-gray">Dibatalkan</span> QR dibatalkan (oleh guest, admin, atau sistem). Order tidak dibuat.</li>
        </ul>
      </div>

      <div class="info-block">
        <h3 class="info-h"><span class="status-badge status-warn">Belum difulfill</span></h3>
        <p class="info-p">
          Muncul pada transaksi <strong>Lunas</strong>. Artinya uang masuk tapi listmak
          <strong>sudah tutup</strong> saat pembayaran tiba — order sengaja tidak dibuat
          agar tidak ada order hantu. Tindakan: <strong>refund manual</strong> lewat dashboard Pakasir.
        </p>
      </div>

      <div class="info-block">
        <h3 class="info-h">Reconcile otomatis (tiap 2 menit)</h3>
        <p class="info-p">
          Sistem memeriksa transaksi <strong>pending yang lebih tua dari 10 menit</strong>, lalu:
        </p>
        <ul class="info-list">
          <li>Tanya status ke Pakasir (<em>transactiondetail</em> — sumber sah).</li>
          <li>Kalau ternyata <strong>sudah dibayar</strong> → dilunasi otomatis (order dibuat, uang tidak hilang walau webhook telat).</li>
          <li>Kalau <strong>belum dibayar</strong> → dibatalkan, supaya QR (termasuk yang di-screenshot / di-download) <strong>tidak bisa dibayar lagi</strong>.</li>
        </ul>
      </div>

      <div class="info-block">
        <h3 class="info-h">Aksi manual per transaksi</h3>
        <ul class="info-list">
          <li><i class="pi pi-sync info-ic"></i> <strong>Sinkron</strong> — paksa cek status terbaru ke Pakasir saat itu juga. Untuk pending macet / webhook telat.</li>
          <li><i class="pi pi-times info-ic"></i> <strong>Batalkan</strong> — hanya untuk pending. Kalau ternyata sudah dibayar, Pakasir menolak (gunakan refund).</li>
        </ul>
      </div>

      <div class="info-block">
        <h3 class="info-h"><span class="status-badge status-sandbox">Sandbox</span></h3>
        <p class="info-p">
          Transaksi mode tes (ditandai oleh Pakasir). <strong>Tidak dihitung</strong> sebagai
          pendapatan di kartu "Total diterima". Dipakai saat uji coba integrasi.
        </p>
      </div>

      <div class="info-block">
        <h3 class="info-h">Keamanan</h3>
        <p class="info-p">
          Status <strong>Lunas</strong> hanya ditetapkan dari verifikasi Pakasir
          (<em>transactiondetail = completed</em>) + cek nominal. Webhook, screenshot QR,
          atau pembayaran pada transaksi yang sudah dibatalkan <strong>tidak bisa</strong>
          memalsukan pelunasan.
        </p>
      </div>
    </div>

    <!-- Revenue stats -->
    <div class="stat-cards">
      <div class="stat-card stat-card--revenue">
        <span class="stat-card-label">Total diterima</span>
        <span class="stat-card-value">Rp {{ formatRupiah(stats.total_collected) }}</span>
        <span class="stat-card-sub">Fee gateway Rp {{ formatRupiah(stats.total_fee) }} · di luar sandbox</span>
      </div>
      <div class="stat-card">
        <span class="stat-card-label">Lunas</span>
        <span class="stat-card-value stat-green">{{ stats.completed_count || 0 }}</span>
      </div>
      <div class="stat-card">
        <span class="stat-card-label">Pending</span>
        <span class="stat-card-value stat-amber">{{ stats.pending_count || 0 }}</span>
      </div>
      <div class="stat-card">
        <span class="stat-card-label">Dibatalkan</span>
        <span class="stat-card-value stat-gray">{{ stats.cancelled_count || 0 }}</span>
      </div>
    </div>

    <!-- Filter panel -->
    <div class="filter-panel">
      <div class="filter-search-wrap">
        <i class="pi pi-search filter-search-icon"></i>
        <input
          v-model="filters.search"
          class="filter-search"
          type="search"
          placeholder="Cari order id / nama / WhatsApp..."
          @keyup.enter="applyFilters"
        />
      </div>
      <div class="filter-row">
        <span class="filter-section-label">Status:</span>
        <button
          v-for="s in statusOptions"
          :key="s.value"
          class="filter-chip"
          :class="{ 'filter-chip--active': filters.status === s.value }"
          @click="setStatus(s.value)"
        >{{ s.label }}</button>
      </div>
      <div class="filter-actions">
        <button class="apply-btn" @click="applyFilters">
          <i class="pi pi-filter"></i> Terapkan
        </button>
        <button class="reset-btn" @click="resetFilters">
          <i class="pi pi-refresh"></i> Reset
        </button>
      </div>
    </div>

    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat pembayaran...</p>
    </div>

    <div v-else-if="error" class="state-block state-error">
      <i class="pi pi-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <button class="retry-btn" @click="loadPayments">Coba lagi</button>
    </div>

    <template v-else>
      <div class="stats-row">
        <span class="stats-label">{{ total }} transaksi</span>
      </div>

      <div v-if="payments.length === 0" class="state-block">
        <i class="pi pi-inbox"></i>
        <p>Tidak ada pembayaran</p>
      </div>

      <div v-else class="logs-table-wrap">
        <table class="logs-table">
          <thead>
            <tr>
              <th>Waktu</th>
              <th>Order ID</th>
              <th>Pembayar</th>
              <th>Jumlah</th>
              <th>Metode</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in payments" :key="p.order_id">
              <td class="td-time">{{ formatTime(p.created_at) }}</td>
              <td class="td-reqid">
                <span class="reqid-badge">{{ p.order_id }}</span>
              </td>
              <td>
                <div class="pay-guest">
                  <span class="pay-guest-name">{{ p.guest_name }}</span>
                  <span class="pay-guest-wa">{{ p.guest_whatsapp }}</span>
                </div>
              </td>
              <td class="td-amount">Rp {{ formatRupiah(p.total_payment || p.amount) }}</td>
              <td class="td-provider">{{ p.payment_method || '-' }}</td>
              <td>
                <span class="status-badge" :class="statusClass(p.status)">{{ statusLabel(p.status) }}</span>
                <span
                  v-if="p.status === 'completed' && !p.fulfilled"
                  class="status-badge status-warn"
                  title="Lunas tapi listmak sudah tutup — order tidak dibuat. Perlu refund/review."
                >Belum difulfill</span>
                <span v-if="p.is_sandbox" class="status-badge status-sandbox" title="Transaksi mode sandbox (tes) — tidak dihitung sebagai pendapatan">Sandbox</span>
              </td>
              <td class="td-actions">
                <button class="detail-btn" title="Detail" @click="openDetail(p.order_id)">
                  <i class="pi pi-eye"></i>
                </button>
                <button
                  class="detail-btn"
                  title="Sinkron status ke Pakasir"
                  :disabled="busyId === p.order_id"
                  @click="reconcile(p.order_id)"
                >
                  <i class="pi" :class="busyId === p.order_id ? 'pi-spin pi-spinner' : 'pi-sync'"></i>
                </button>
                <button
                  v-if="p.status === 'pending'"
                  class="detail-btn detail-btn--danger"
                  title="Batalkan"
                  :disabled="busyId === p.order_id"
                  @click="cancel(p.order_id)"
                >
                  <i class="pi pi-times"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination">
        <button class="page-btn" :disabled="page <= 1" @click="changePage(-1)">
          <i class="pi pi-chevron-left"></i> Sebelumnya
        </button>
        <button class="page-btn" :disabled="!hasNextPage" @click="changePage(1)">
          Berikutnya <i class="pi pi-chevron-right"></i>
        </button>
      </div>
    </template>

    <!-- Detail modal -->
    <div v-if="showDetail" class="detail-backdrop" @click.self="showDetail = false">
      <div class="detail-modal">
        <div class="detail-header">
          <h2 class="detail-title">Detail Pembayaran</h2>
          <button class="detail-close" @click="showDetail = false">
            <i class="pi pi-times"></i>
          </button>
        </div>

        <div v-if="detailLoading" class="state-block">
          <i class="pi pi-spin pi-spinner"></i>
        </div>

        <div v-else-if="detail" class="detail-body">
          <div class="detail-row">
            <span class="detail-label">Order ID</span>
            <span class="detail-value mono">{{ detail.payment.order_id }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Status</span>
            <span class="status-badge" :class="statusClass(detail.payment.status)">{{ statusLabel(detail.payment.status) }}</span>
            <span
              v-if="detail.payment.status === 'completed' && !detail.payment.fulfilled"
              class="status-badge status-warn"
            >Belum difulfill</span>
            <span v-if="detail.payment.is_sandbox" class="status-badge status-sandbox">Sandbox</span>
          </div>
          <div
            v-if="detail.payment.status === 'completed' && !detail.payment.fulfilled"
            class="detail-warn"
          >
            <i class="pi pi-exclamation-triangle"></i>
            Pembayaran masuk tapi listmak sudah tutup — order tidak dibuat. Perlu refund manual lewat dashboard Pakasir.
          </div>
          <div class="detail-row">
            <span class="detail-label">Pembayar</span>
            <span class="detail-value">{{ detail.payment.guest_name }} · {{ detail.payment.guest_whatsapp }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Jumlah</span>
            <span class="detail-value">
              Rp {{ formatRupiah(detail.payment.amount) }}
              <span class="detail-muted">+ fee Rp {{ formatRupiah(detail.payment.fee) }} = <strong>Rp {{ formatRupiah(detail.payment.total_payment) }}</strong></span>
            </span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Metode</span>
            <span class="detail-value">{{ detail.payment.payment_method || '-' }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Dibuat</span>
            <span class="detail-value">{{ formatTime(detail.payment.created_at) }}</span>
          </div>
          <div v-if="detail.payment.completed_at" class="detail-row">
            <span class="detail-label">Dibayar</span>
            <span class="detail-value">{{ formatTime(detail.payment.completed_at) }}</span>
          </div>
          <div v-if="detail.payment.expires_at" class="detail-row">
            <span class="detail-label">Kedaluwarsa</span>
            <span class="detail-value">{{ formatTime(detail.payment.expires_at) }}</span>
          </div>

          <div class="detail-field">
            <span class="detail-label">Pesanan ({{ (detail.orders || []).length }})</span>
            <ul v-if="(detail.orders || []).length" class="detail-order-list">
              <li v-for="o in detail.orders" :key="o.id" class="detail-order-item">
                <span class="detail-order-name">{{ o.name }}</span>
                <span class="detail-order-detail">{{ o.order_detail }}</span>
                <span class="detail-order-price">Rp {{ formatRupiah((o.price || 0) * (o.qty || 1)) }}</span>
              </li>
            </ul>
            <p v-else class="detail-text">Pesanan belum dibuat (menunggu pembayaran).</p>
          </div>

          <!-- Gateway log -->
          <div class="detail-field">
            <span class="detail-label">Log transaksi ({{ (detail.logs || []).length }})</span>
            <ul v-if="(detail.logs || []).length" class="log-list">
              <li v-for="lg in detail.logs" :key="lg.id" class="log-item">
                <div class="log-head" @click="expandedLog = expandedLog === lg.id ? null : lg.id">
                  <span class="log-action" :class="'log-' + lg.action">{{ logLabel(lg.action) }}</span>
                  <span class="log-code" :class="lg.success ? 'log-ok' : 'log-err'">{{ lg.status_code || '—' }}</span>
                  <span class="log-time">{{ formatTime(lg.created_at) }}</span>
                  <i class="pi log-chev" :class="expandedLog === lg.id ? 'pi-chevron-up' : 'pi-chevron-down'"></i>
                </div>
                <div v-if="expandedLog === lg.id" class="log-body">
                  <p v-if="lg.error" class="log-err-text">{{ lg.error }}</p>
                  <pre v-if="lg.response" class="log-json">{{ prettyJson(lg.response) }}</pre>
                </div>
              </li>
            </ul>
            <p v-else class="detail-text">Belum ada log.</p>
          </div>

          <div class="detail-actions">
            <button class="apply-btn" :disabled="busyId === detail.payment.order_id" @click="reconcile(detail.payment.order_id, true)">
              <i class="pi" :class="busyId === detail.payment.order_id ? 'pi-spin pi-spinner' : 'pi-sync'"></i> Sinkron status
            </button>
            <button
              v-if="detail.payment.status === 'pending'"
              class="reset-btn danger-btn"
              :disabled="busyId === detail.payment.order_id"
              @click="cancel(detail.payment.order_id, true)"
            >
              <i class="pi pi-times"></i> Batalkan
            </button>
          </div>
        </div>
      </div>
    </div>

    <Toast position="top-center" />
  </div>
</template>

<script>
import Toast from 'primevue/toast'
import admin from '../api/admin'

export default {
  name: 'AdminPaymentsView',
  components: { Toast },

  data() {
    return {
      payments: [],
      stats: {},
      total: 0,
      page: 1,
      loading: false,
      error: '',
      filters: { status: '', search: '' },
      statusOptions: [
        { value: '', label: 'Semua' },
        { value: 'pending', label: 'Pending' },
        { value: 'completed', label: 'Lunas' },
        { value: 'cancelled', label: 'Dibatalkan' },
        { value: 'expired', label: 'Kedaluwarsa' },
      ],
      busyId: '',
      showDetail: false,
      detailLoading: false,
      detail: null,
      showInfo: false,
      expandedLog: null,
    }
  },

  computed: {
    hasNextPage() {
      return this.page * 30 < this.total
    },
  },

  mounted() {
    this.loadPayments()
  },

  methods: {
    async loadPayments() {
      this.loading = true
      this.error = ''
      try {
        const res = await admin.getPayments(this.page, this.filters.status, this.filters.search)
        this.payments = res.data?.payments || []
        this.stats = res.data?.stats || {}
        this.total = res.data?.total || 0
      } catch (err) {
        this.error = err.message || 'Gagal memuat pembayaran'
      } finally {
        this.loading = false
      }
    },

    applyFilters() {
      this.page = 1
      this.loadPayments()
    },

    setStatus(status) {
      this.filters.status = status
      this.page = 1
      this.loadPayments()
    },

    resetFilters() {
      this.filters = { status: '', search: '' }
      this.page = 1
      this.loadPayments()
    },

    changePage(delta) {
      this.page += delta
      this.loadPayments()
    },

    async openDetail(orderId) {
      this.showDetail = true
      this.detailLoading = true
      this.detail = null
      this.expandedLog = null
      try {
        const res = await admin.getPaymentDetail(orderId)
        this.detail = res.data
      } catch (err) {
        this.$toast.add({ severity: 'error', summary: 'Gagal', detail: err.message, life: 3000 })
        this.showDetail = false
      } finally {
        this.detailLoading = false
      }
    },

    async reconcile(orderId, inModal = false) {
      this.busyId = orderId
      try {
        const res = await admin.reconcilePayment(orderId)
        const status = res.data?.status
        this.$toast.add({
          severity: status === 'completed' ? 'success' : 'info',
          summary: 'Sinkron selesai',
          detail: `Status sekarang: ${this.statusLabel(status)}`,
          life: 3000,
        })
        await this.loadPayments()
        if (inModal) await this.openDetail(orderId)
      } catch (err) {
        this.$toast.add({ severity: 'error', summary: 'Gagal sinkron', detail: err.message, life: 3000 })
      } finally {
        this.busyId = ''
      }
    },

    async cancel(orderId, inModal = false) {
      if (!confirm('Batalkan pembayaran ini? Kode QR akan dibatalkan di Pakasir.')) return
      this.busyId = orderId
      try {
        await admin.cancelPayment(orderId)
        this.$toast.add({ severity: 'success', summary: 'Dibatalkan', detail: orderId, life: 3000 })
        if (inModal) this.showDetail = false
        await this.loadPayments()
      } catch (err) {
        this.$toast.add({ severity: 'error', summary: 'Gagal batalkan', detail: err.message, life: 3000 })
      } finally {
        this.busyId = ''
      }
    },

    formatRupiah(n) {
      return Number(n || 0).toLocaleString('id-ID')
    },

    formatTime(t) {
      if (!t) return '-'
      return new Date(t).toLocaleString('id-ID', {
        day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit',
      })
    },

    statusLabel(s) {
      return { completed: 'Lunas', pending: 'Pending', cancelled: 'Dibatalkan', expired: 'Kedaluwarsa' }[s] || s
    },

    logLabel(a) {
      return { create: 'Buat', detail: 'Cek status', cancel: 'Batal', webhook: 'Webhook', complete: 'Selesai' }[a] || a
    },

    prettyJson(obj) {
      try {
        return JSON.stringify(obj, null, 2)
      } catch {
        return String(obj)
      }
    },

    statusClass(s) {
      return {
        'status-ok': s === 'completed',
        'status-pending': s === 'pending',
        'status-gray': s === 'cancelled',
        'status-err': s === 'expired',
      }
    },
  },
}
</script>

<style scoped>
.admin-container { padding: 1.5rem; max-width: 1100px; }

.admin-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}

.admin-title { font-size: 1.375rem; font-weight: 700; color: #f1f5f9; margin: 0; }

.admin-badge {
  padding: 0.2rem 0.6rem;
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 700;
  color: #ef4444;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  margin-left: auto;
  padding: 0.375rem 0.75rem;
  background: rgba(99, 102, 241, 0.12);
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 0.625rem;
  color: #818cf8;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.info-toggle:hover { background: rgba(99, 102, 241, 0.2); }

/* Documentation panel */
.info-panel {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 0.875rem;
  padding: 1.25rem;
  margin-bottom: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.info-block { display: flex; flex-direction: column; gap: 0.5rem; }

.info-h {
  font-size: 0.875rem;
  font-weight: 700;
  color: #e2e8f0;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.info-p {
  font-size: 0.8125rem;
  line-height: 1.6;
  color: #94a3b8;
  margin: 0;
}

.info-p strong { color: #cbd5e1; }
.info-p em { color: #a5b4fc; font-style: normal; font-family: monospace; font-size: 0.75rem; }

.info-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.info-list li {
  font-size: 0.8125rem;
  line-height: 1.55;
  color: #94a3b8;
}

.info-list strong { color: #cbd5e1; }
.info-list em { color: #a5b4fc; font-style: normal; font-family: monospace; font-size: 0.75rem; }
.info-ic { color: #64748b; font-size: 0.75rem; }

/* Stat cards */
.stat-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}

.stat-card {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.875rem;
  padding: 0.875rem 1rem;
}

.stat-card--revenue {
  border-color: rgba(34, 197, 94, 0.3);
  background: rgba(34, 197, 94, 0.06);
}

.stat-card-label {
  font-size: 0.7rem;
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-card-value { font-size: 1.25rem; font-weight: 700; color: #f1f5f9; }
.stat-card-sub { font-size: 0.7rem; color: #64748b; }
.stat-green { color: #22c55e; }
.stat-amber { color: #f59e0b; }
.stat-gray { color: #94a3b8; }

/* Filter panel */
.filter-panel {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.875rem;
  padding: 1rem;
  margin-bottom: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.filter-search-wrap { position: relative; }

.filter-search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: #475569;
  font-size: 0.875rem;
  pointer-events: none;
}

.filter-search {
  width: 100%;
  padding: 0.5rem 0.75rem 0.5rem 2.25rem;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.625rem;
  color: #f1f5f9;
  font-size: 0.875rem;
  box-sizing: border-box;
}

.filter-search:focus { outline: none; border-color: rgba(99, 102, 241, 0.4); }
.filter-search::placeholder { color: #334155; }

.filter-row { display: flex; flex-wrap: wrap; align-items: center; gap: 0.375rem; }

.filter-section-label {
  font-size: 0.75rem;
  color: #475569;
  font-weight: 600;
  margin-right: 0.25rem;
}

.filter-chip {
  padding: 0.25rem 0.75rem;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 999px;
  color: #94a3b8;
  font-size: 0.8125rem;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s, color 0.15s;
}

.filter-chip:hover { background: rgba(255, 255, 255, 0.07); }

.filter-chip--active {
  background: rgba(99, 179, 237, 0.15);
  border-color: rgba(99, 179, 237, 0.4);
  color: #63b3ed;
  font-weight: 600;
}

.filter-actions { display: flex; gap: 0.5rem; }

.apply-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  background: rgba(99, 102, 241, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 0.625rem;
  color: #818cf8;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
}

.apply-btn:hover { background: rgba(99, 102, 241, 0.25); }
.apply-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.reset-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.875rem;
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.625rem;
  color: #64748b;
  font-size: 0.875rem;
  cursor: pointer;
}

.reset-btn:hover { color: #94a3b8; }
.danger-btn { border-color: rgba(239, 68, 68, 0.3); color: #f87171; }
.danger-btn:hover { color: #ef4444; }

.stats-row { display: flex; gap: 1.5rem; margin-bottom: 1rem; }
.stats-label { font-size: 0.8125rem; color: #64748b; }

.logs-table-wrap {
  overflow-x: auto;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.75rem;
}

.logs-table { width: 100%; border-collapse: collapse; font-size: 0.8125rem; }

.logs-table th {
  padding: 0.625rem 0.875rem;
  text-align: left;
  font-size: 0.7rem;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: rgba(15, 23, 42, 0.6);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  white-space: nowrap;
}

.logs-table td {
  padding: 0.5rem 0.875rem;
  color: #cbd5e1;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  vertical-align: middle;
}

.logs-table tbody tr:last-child td { border-bottom: none; }
.logs-table tbody tr:hover td { background: rgba(255, 255, 255, 0.02); }

.td-time { white-space: nowrap; color: #64748b; font-size: 0.7rem; }
.td-reqid { white-space: nowrap; }
.td-amount { white-space: nowrap; font-weight: 600; color: #f1f5f9; }
.td-provider { white-space: nowrap; color: #64748b; font-size: 0.7rem; }
.td-actions { white-space: nowrap; display: flex; gap: 0.375rem; }

.reqid-badge {
  display: inline-block;
  padding: 0.1rem 0.4rem;
  background: rgba(99, 102, 241, 0.08);
  border: 1px solid rgba(99, 102, 241, 0.15);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.7rem;
  color: #818cf8;
}

.pay-guest { display: flex; flex-direction: column; }
.pay-guest-name { color: #e2e8f0; font-weight: 600; }
.pay-guest-wa { font-size: 0.7rem; color: #64748b; }

.status-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 600;
  white-space: nowrap;
}

.status-ok { background: rgba(34, 197, 94, 0.12); color: #22c55e; }
.status-pending { background: rgba(245, 158, 11, 0.12); color: #f59e0b; }
.status-gray { background: rgba(148, 163, 184, 0.12); color: #94a3b8; }
.status-err { background: rgba(239, 68, 68, 0.12); color: #ef4444; }
.status-warn { background: rgba(245, 158, 11, 0.14); color: #f59e0b; margin-left: 0.375rem; }
.status-sandbox { background: rgba(168, 85, 247, 0.15); color: #c084fc; margin-left: 0.375rem; }

.detail-warn {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: 0.5rem;
  padding: 0.625rem 0.75rem;
  font-size: 0.8125rem;
  color: #fcd34d;
  line-height: 1.5;
}

.detail-warn i { color: #f59e0b; flex-shrink: 0; margin-top: 0.1rem; }

.pagination { display: flex; gap: 0.75rem; margin-top: 1.25rem; justify-content: flex-end; }

.page-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.875rem;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 0.5rem;
  color: #94a3b8;
  font-size: 0.875rem;
  cursor: pointer;
}

.page-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.page-btn:not(:disabled):hover { background: rgba(30, 41, 59, 0.9); color: #cbd5e1; }

.state-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2.5rem 1rem;
  color: #64748b;
  gap: 0.75rem;
}

.state-block > i { font-size: 2rem; color: #3b82f6; }
.state-error > i { color: #ef4444; }

.retry-btn {
  padding: 0.5rem 1rem;
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 0.5rem;
  color: #3b82f6;
  font-size: 0.875rem;
  cursor: pointer;
}

.detail-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 0.375rem;
  color: #64748b;
  cursor: pointer;
  font-size: 0.8rem;
}

.detail-btn:hover { color: #94a3b8; background: rgba(255, 255, 255, 0.08); }
.detail-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.detail-btn--danger { color: #f87171; border-color: rgba(239, 68, 68, 0.25); }
.detail-btn--danger:hover { color: #ef4444; background: rgba(239, 68, 68, 0.12); }

/* Detail modal */
.detail-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.detail-modal {
  background: #0f172a;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 1rem;
  width: 100%;
  max-width: 560px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.detail-title { font-size: 1rem; font-weight: 700; color: #f1f5f9; margin: 0; }

.detail-close {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.06);
  border: none;
  border-radius: 0.5rem;
  color: #94a3b8;
  cursor: pointer;
}

.detail-body { padding: 1.25rem; overflow-y: auto; display: flex; flex-direction: column; gap: 0.75rem; }
.detail-row { display: flex; align-items: baseline; gap: 0.75rem; }
.detail-field { display: flex; flex-direction: column; gap: 0.5rem; }

.detail-label {
  font-size: 0.6875rem;
  font-weight: 700;
  color: #475569;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  min-width: 90px;
  flex-shrink: 0;
}

.detail-value { font-size: 0.8125rem; color: #cbd5e1; }
.detail-value.mono { font-family: monospace; font-size: 0.75rem; color: #94a3b8; word-break: break-all; }
.detail-muted { color: #64748b; }
.detail-muted strong { color: #cbd5e1; }

.detail-order-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 0.375rem; }

.detail-order-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.5rem;
  padding: 0.5rem 0.625rem;
  font-size: 0.8125rem;
}

.detail-order-name { font-weight: 600; color: #e2e8f0; min-width: 4.5rem; flex-shrink: 0; }
.detail-order-detail { color: #94a3b8; flex: 1; overflow-wrap: anywhere; }
.detail-order-price { color: #cbd5e1; white-space: nowrap; }

.detail-text {
  font-size: 0.8125rem;
  color: #64748b;
  margin: 0;
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.5rem;
  padding: 0.625rem 0.75rem;
}

.detail-actions { display: flex; gap: 0.5rem; margin-top: 0.5rem; }

/* Gateway log */
.log-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 0.375rem; }

.log-item {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.5rem;
  overflow: hidden;
}

.log-head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.625rem;
  cursor: pointer;
  font-size: 0.75rem;
}

.log-head:hover { background: rgba(255, 255, 255, 0.02); }

.log-action {
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 0.25rem;
  background: rgba(99, 102, 241, 0.15);
  color: #a5b4fc;
}

.log-create { background: rgba(59, 130, 246, 0.15); color: #60a5fa; }
.log-detail { background: rgba(148, 163, 184, 0.15); color: #cbd5e1; }
.log-webhook { background: rgba(168, 85, 247, 0.15); color: #c084fc; }
.log-complete { background: rgba(34, 197, 94, 0.15); color: #22c55e; }
.log-cancel { background: rgba(239, 68, 68, 0.15); color: #f87171; }

.log-code { font-family: monospace; font-weight: 700; }
.log-ok { color: #22c55e; }
.log-err { color: #ef4444; }
.log-time { color: #64748b; margin-left: auto; }
.log-chev { color: #475569; font-size: 0.7rem; }

.log-body { padding: 0 0.625rem 0.625rem; }
.log-err-text { color: #f87171; font-size: 0.75rem; margin: 0.25rem 0; }

.log-json {
  margin: 0;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.375rem;
  padding: 0.5rem 0.625rem;
  font-size: 0.7rem;
  line-height: 1.5;
  color: #a5b4fc;
  font-family: 'Fira Code', monospace;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 240px;
  overflow: auto;
}
</style>
