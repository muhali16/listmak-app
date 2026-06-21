<template>
  <div class="admin-container">
    <div class="admin-header">
      <h1 class="admin-title">System Logs</h1>
      <span class="admin-badge">Admin</span>
    </div>

    <!-- Filters -->
    <div class="filter-panel">
      <!-- Request ID search -->
      <div class="filter-search-wrap">
        <i class="pi pi-search filter-search-icon"></i>
        <input
          v-model="filters.requestId"
          class="filter-search"
          type="search"
          placeholder="Cari request ID..."
          @keyup.enter="applyFilters"
        />
      </div>

      <!-- Period chips -->
      <div class="filter-row">
        <button
          v-for="p in periodOptions"
          :key="p.value"
          class="filter-chip"
          :class="{ 'filter-chip--active': filters.period === p.value }"
          @click="setPeriod(p.value)"
        >{{ p.label }}</button>
        <button
          class="filter-chip"
          :class="{ 'filter-chip--active': filters.period === 'custom' }"
          @click="filters.period = 'custom'"
        >Kustom</button>
      </div>

      <!-- Custom date range -->
      <div v-if="filters.period === 'custom'" class="filter-row filter-dates">
        <div class="filter-date-group">
          <label class="filter-label">Dari</label>
          <input v-model="filters.from" type="datetime-local" class="filter-date-input" />
        </div>
        <div class="filter-date-group">
          <label class="filter-label">Sampai</label>
          <input v-model="filters.to" type="datetime-local" class="filter-date-input" />
        </div>
      </div>

      <!-- Method chips -->
      <div class="filter-row">
        <span class="filter-section-label">Method:</span>
        <button
          v-for="m in methodOptions"
          :key="m"
          class="filter-chip method-chip"
          :class="[`method-${m.toLowerCase()}`, filters.method === m ? 'filter-chip--active' : '']"
          @click="toggleMethod(m)"
        >{{ m }}</button>
      </div>

      <!-- Status chips -->
      <div class="filter-row">
        <span class="filter-section-label">Status:</span>
        <button
          v-for="s in statusOptions"
          :key="s.value"
          class="filter-chip"
          :class="{ 'filter-chip--active': filters.status === s.value }"
          @click="toggleStatus(s.value)"
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
      <p>Memuat logs...</p>
    </div>

    <div v-else-if="error" class="state-block state-error">
      <i class="pi pi-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <button class="retry-btn" @click="loadLogs">Coba lagi</button>
    </div>

    <template v-else>
      <div class="stats-row">
        <span class="stats-label">Total: {{ total }} logs</span>
        <span class="stats-label">Halaman {{ page }}</span>
      </div>

      <div v-if="logs.length > 0" class="logs-table-wrap">
        <table class="logs-table">
          <thead>
            <tr>
              <th>Waktu</th>
              <th>Req ID</th>
              <th>Method</th>
              <th>Path</th>
              <th>Status</th>
              <th>Latency</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td class="td-time">{{ formatTimeCompact(log.created_at) }}</td>
              <td class="td-reqid">
                <span v-if="log.request_id" class="reqid-badge" :title="log.request_id">{{ log.request_id.slice(0, 8) }}</span>
                <span v-else class="td-empty">—</span>
              </td>
              <td><span class="method-badge" :class="`method-${log.method?.toLowerCase()}`">{{ log.method }}</span></td>
              <td class="td-path" :title="log.path">{{ log.path }}</td>
              <td><span class="status-badge" :class="statusClass(log.status_code)">{{ log.status_code }}</span></td>
              <td class="td-latency">{{ log.latency }}</td>
              <td>
                <button class="detail-btn" @click="detailLog = log">
                  <i class="pi pi-eye"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="state-block">
        <p>Tidak ada logs untuk filter ini.</p>
      </div>

      <div v-if="logs.length > 0" class="pagination">
        <button class="page-btn" :disabled="page <= 1" @click="changePage(page - 1)">
          <i class="pi pi-chevron-left"></i> Sebelumnya
        </button>
        <button class="page-btn" :disabled="page * 100 >= total" @click="changePage(page + 1)">
          Berikutnya <i class="pi pi-chevron-right"></i>
        </button>
      </div>
    </template>
  </div>

  <!-- Detail modal -->
  <teleport to="body">
    <div v-if="detailLog" class="detail-backdrop" @click.self="detailLog = null">
      <div class="detail-modal">
        <div class="detail-header">
          <h2 class="detail-title">Log #{{ detailLog.id }}</h2>
          <button class="detail-close" @click="detailLog = null">
            <i class="pi pi-times"></i>
          </button>
        </div>
        <div class="detail-body">
          <div class="detail-row">
            <span class="detail-label">ID</span>
            <span class="detail-value">{{ detailLog.id }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Request ID</span>
            <span class="detail-value mono">{{ detailLog.request_id }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Method</span>
            <span class="method-badge" :class="`method-${detailLog.method?.toLowerCase()}`">{{ detailLog.method }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Path</span>
            <span class="detail-value mono">{{ detailLog.path }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Status</span>
            <span class="status-badge" :class="statusClass(detailLog.status_code)">{{ detailLog.status_code }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Latency</span>
            <span class="detail-value">{{ detailLog.latency }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Client IP</span>
            <span class="detail-value mono">{{ detailLog.client_ip }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Created At</span>
            <span class="detail-value">{{ formatTime(detailLog.created_at) }}</span>
          </div>
          <div v-if="detailLog.error_msg" class="detail-field">
            <span class="detail-label">Error</span>
            <p class="detail-text detail-error">{{ detailLog.error_msg }}</p>
          </div>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script>
import admin from '../api/admin'

function defaultFrom() {
  const d = new Date()
  d.setDate(d.getDate() - 7)
  return d.toISOString().slice(0, 16)
}

export default {
  name: 'AdminSystemLogsView',

  data() {
    return {
      logs: [],
      total: 0,
      page: 1,
      loading: false,
      error: '',
      detailLog: null,
      filters: {
        requestId: '',
        method: '',
        status: 0,
        period: '7d',
        from: defaultFrom(),
        to: ''
      },
      periodOptions: [
        { value: '1d', label: '1 hari' },
        { value: '7d', label: '7 hari' },
        { value: '30d', label: '30 hari' },
      ],
      methodOptions: ['GET', 'POST', 'PATCH', 'PUT', 'DELETE'],
      statusOptions: [
        { value: 200, label: '2xx' },
        { value: 300, label: '3xx' },
        { value: 400, label: '4xx' },
        { value: 500, label: '5xx' },
      ]
    }
  },

  created() {
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    if (user.role !== 'admin') {
      this.$router.replace('/today')
      return
    }
    this.loadLogs()
  },

  methods: {
    setPeriod(val) {
      this.filters.period = val
      const days = { '1d': 1, '7d': 7, '30d': 30 }[val] || 7
      const d = new Date()
      d.setDate(d.getDate() - days)
      this.filters.from = d.toISOString().slice(0, 16)
      this.filters.to = ''
    },

    toggleMethod(m) {
      this.filters.method = this.filters.method === m ? '' : m
    },

    toggleStatus(s) {
      this.filters.status = this.filters.status === s ? 0 : s
    },

    applyFilters() {
      this.page = 1
      this.loadLogs()
    },

    resetFilters() {
      this.filters = {
        requestId: '',
        method: '',
        status: 0,
        period: '7d',
        from: defaultFrom(),
        to: ''
      }
      this.page = 1
      this.loadLogs()
    },

    buildQuery() {
      const params = new URLSearchParams()
      params.set('page', this.page)
      if (this.filters.requestId) params.set('request_id', this.filters.requestId)
      if (this.filters.method) params.set('method', this.filters.method)
      if (this.filters.status) params.set('status', this.filters.status)
      if (this.filters.from) params.set('from', new Date(this.filters.from).toISOString())
      if (this.filters.to) params.set('to', new Date(this.filters.to).toISOString())
      return params.toString()
    },

    async loadLogs() {
      this.loading = true
      this.error = ''
      try {
        const res = await admin.getSystemLogs(this.buildQuery())
        if (!res.success) throw new Error(res.message || 'Gagal memuat logs.')
        this.logs = res.data?.logs || []
        this.total = res.data?.total || 0
      } catch (err) {
        this.error = err.message || 'Gagal memuat logs.'
      } finally {
        this.loading = false
      }
    },

    changePage(newPage) {
      this.page = newPage
      this.loadLogs()
    },

    formatTime(iso) {
      if (!iso) return '—'
      return new Date(iso).toLocaleString('id-ID', {
        day: '2-digit', month: 'short', year: 'numeric',
        hour: '2-digit', minute: '2-digit', second: '2-digit'
      })
    },

    formatTimeCompact(iso) {
      if (!iso) return '—'
      const d = new Date(iso)
      return d.toLocaleString('id-ID', {
        day: '2-digit', month: 'short',
        hour: '2-digit', minute: '2-digit'
      })
    },

    statusClass(code) {
      if (!code) return ''
      if (code < 300) return 'status-ok'
      if (code < 400) return 'status-redirect'
      if (code < 500) return 'status-client-err'
      return 'status-server-err'
    }
  }
}
</script>

<style scoped>
.admin-container {
  padding: 1.5rem;
  max-width: 1100px;
}

.admin-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}

.admin-title {
  font-size: 1.375rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}

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

.filter-search-wrap {
  position: relative;
}

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
  transition: border-color 0.15s;
}

.filter-search:focus {
  outline: none;
  border-color: rgba(99, 102, 241, 0.4);
}

.filter-search::placeholder { color: #334155; }

.filter-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.375rem;
}

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

.filter-chip:hover {
  background: rgba(255, 255, 255, 0.07);
}

.filter-chip--active {
  background: rgba(99, 179, 237, 0.15);
  border-color: rgba(99, 179, 237, 0.4);
  color: #63b3ed;
  font-weight: 600;
}

/* Method chips keep color when active */
.method-chip.filter-chip--active.method-get    { background: rgba(34, 197, 94, 0.15);  border-color: rgba(34, 197, 94, 0.4);  color: #22c55e; }
.method-chip.filter-chip--active.method-post   { background: rgba(59, 130, 246, 0.15); border-color: rgba(59, 130, 246, 0.4); color: #60a5fa; }
.method-chip.filter-chip--active.method-patch  { background: rgba(234, 179, 8, 0.15);  border-color: rgba(234, 179, 8, 0.4);  color: #eab308; }
.method-chip.filter-chip--active.method-put    { background: rgba(234, 179, 8, 0.15);  border-color: rgba(234, 179, 8, 0.4);  color: #eab308; }
.method-chip.filter-chip--active.method-delete { background: rgba(239, 68, 68, 0.15);  border-color: rgba(239, 68, 68, 0.4);  color: #ef4444; }

.filter-dates {
  gap: 0.75rem;
}

.filter-date-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
  min-width: 0;
}

.filter-label {
  font-size: 0.6875rem;
  font-weight: 600;
  color: #475569;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.filter-date-input {
  padding: 0.375rem 0.625rem;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.8125rem;
  width: 100%;
  box-sizing: border-box;
}

.filter-date-input:focus {
  outline: none;
  border-color: rgba(99, 102, 241, 0.4);
}

.filter-actions {
  display: flex;
  gap: 0.5rem;
}

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
  transition: background 0.15s;
}

.apply-btn:hover {
  background: rgba(99, 102, 241, 0.25);
}

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
  transition: color 0.15s;
}

.reset-btn:hover { color: #94a3b8; }

.stats-row {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1rem;
}

.stats-label {
  font-size: 0.8125rem;
  color: #64748b;
}

.logs-table-wrap {
  overflow-x: auto;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.75rem;
}

.logs-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.logs-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: rgba(15, 23, 42, 0.6);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.logs-table td {
  padding: 0.625rem 1rem;
  color: #cbd5e1;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  vertical-align: top;
}

.logs-table tbody tr:last-child td { border-bottom: none; }
.logs-table tbody tr:hover td { background: rgba(255, 255, 255, 0.02); }

.td-time {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.75rem;
}

.td-path {
  max-width: 240px;
  color: #94a3b8;
  font-family: monospace;
  font-size: 0.75rem;
  word-break: break-all;
}

.td-reqid { white-space: nowrap; }

.reqid-badge {
  display: inline-block;
  padding: 0.1rem 0.4rem;
  background: rgba(99, 102, 241, 0.08);
  border: 1px solid rgba(99, 102, 241, 0.15);
  border-radius: 0.25rem;
  font-family: monospace;
  font-size: 0.7rem;
  color: #818cf8;
  letter-spacing: 0.02em;
  cursor: default;
}

.td-empty { color: #334155; }

.td-latency {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.75rem;
}

.method-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.method-get    { background: rgba(34, 197, 94, 0.12);  color: #22c55e; }
.method-post   { background: rgba(59, 130, 246, 0.12); color: #60a5fa; }
.method-patch  { background: rgba(234, 179, 8, 0.12);  color: #eab308; }
.method-put    { background: rgba(234, 179, 8, 0.12);  color: #eab308; }
.method-delete { background: rgba(239, 68, 68, 0.12);  color: #ef4444; }

.status-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-ok         { background: rgba(34, 197, 94, 0.12);  color: #22c55e; }
.status-redirect   { background: rgba(99, 102, 241, 0.12); color: #818cf8; }
.status-client-err { background: rgba(234, 179, 8, 0.12);  color: #eab308; }
.status-server-err { background: rgba(239, 68, 68, 0.12);  color: #ef4444; }

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
  transition: color 0.15s, background 0.15s;
}

.detail-btn:hover {
  color: #94a3b8;
  background: rgba(255, 255, 255, 0.08);
}

.pagination {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.25rem;
  justify-content: flex-end;
}

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
  transition: background 0.15s;
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
  flex-shrink: 0;
}

.detail-title {
  font-size: 1rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}

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
  font-size: 0.875rem;
}

.detail-body {
  padding: 1.25rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.detail-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.detail-field {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.detail-label {
  font-size: 0.6875rem;
  font-weight: 700;
  color: #475569;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  min-width: 80px;
  flex-shrink: 0;
}

.detail-value {
  font-size: 0.8125rem;
  color: #cbd5e1;
}

.detail-value.mono {
  font-family: monospace;
  font-size: 0.75rem;
  color: #94a3b8;
  word-break: break-all;
}

.detail-text {
  font-size: 0.8125rem;
  color: #94a3b8;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  margin: 0;
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.5rem;
  padding: 0.625rem 0.75rem;
}

.detail-error { color: #f87171; }
</style>
