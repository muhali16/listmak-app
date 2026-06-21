<template>
  <div class="admin-container">
    <div class="admin-header">
      <h1 class="admin-title">AI Request Logs</h1>
      <span class="admin-badge">Admin</span>
    </div>

    <!-- Filter panel -->
    <div class="filter-panel">
      <div class="filter-search-wrap">
        <i class="pi pi-search filter-search-icon"></i>
        <input
          v-model="filters.search"
          class="filter-search"
          type="search"
          placeholder="Cari prompt..."
          @keyup.enter="applyFilters"
        />
      </div>
      <div class="filter-row">
        <span class="filter-section-label">Status:</span>
        <button
          class="filter-chip"
          :class="{ 'filter-chip--active': filters.status === '' }"
          @click="setStatus('')"
        >Semua</button>
        <button
          class="filter-chip status-ok-chip"
          :class="{ 'filter-chip--active': filters.status === 'success' }"
          @click="setStatus('success')"
        >Success</button>
        <button
          class="filter-chip status-err-chip"
          :class="{ 'filter-chip--active': filters.status === 'failed' }"
          @click="setStatus('failed')"
        >Failed</button>
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
              <th>Prompt</th>
              <th>Provider</th>
              <th>Latency</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td class="td-time">{{ formatTime(log.created_at) }}</td>
              <td class="td-reqid">
                <span v-if="log.request_id" class="reqid-badge" :title="log.request_id">{{ log.request_id.slice(0, 8) }}</span>
                <span v-else class="td-empty">—</span>
              </td>
              <td class="td-input" :title="log.input">{{ truncate(log.input, 70) }}</td>
              <td class="td-provider">{{ log.provider }}</td>
              <td class="td-latency">{{ log.latency_ms }}ms</td>
              <td>
                <span class="status-badge" :class="log.status === 'success' ? 'status-ok' : 'status-err'">
                  {{ log.status }}
                </span>
              </td>
              <td>
                <button class="detail-btn" @click="openDetail(log)">
                  <i class="pi pi-eye"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="state-block">
        <p>Belum ada AI logs.</p>
      </div>

      <div v-if="logs.length > 0" class="pagination">
        <button class="page-btn" :disabled="page <= 1" @click="changePage(page - 1)">
          <i class="pi pi-chevron-left"></i> Sebelumnya
        </button>
        <button class="page-btn" :disabled="page * 50 >= total" @click="changePage(page + 1)">
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
            <span class="detail-value mono">{{ detailLog.request_id || '—' }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Order ID</span>
            <span class="detail-value">{{ detailLog.order_id ?? '—' }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Status</span>
            <span class="status-badge" :class="detailLog.status === 'success' ? 'status-ok' : 'status-err'">
              {{ detailLog.status }}
            </span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Provider</span>
            <span class="detail-value">{{ detailLog.provider }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Model</span>
            <span class="detail-value mono">{{ detailLog.model }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Latency</span>
            <span class="detail-value">{{ detailLog.latency_ms }}ms</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Created At</span>
            <span class="detail-value">{{ formatTime(detailLog.created_at) }}</span>
          </div>

          <div class="detail-field">
            <span class="detail-label">Prompt</span>
            <pre v-if="isJson(detailLog.input)" class="detail-code"><code>{{ prettyJson(detailLog.input) }}</code></pre>
            <p v-else class="detail-text">{{ detailLog.input || '—' }}</p>
          </div>

          <div class="detail-field">
            <span class="detail-label">Output</span>
            <pre v-if="isJson(detailLog.output)" class="detail-code"><code>{{ prettyJson(detailLog.output) }}</code></pre>
            <p v-else class="detail-text">{{ detailLog.output || '—' }}</p>
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

export default {
  name: 'AdminAILogsView',

  data() {
    return {
      logs: [],
      total: 0,
      page: 1,
      loading: false,
      error: '',
      detailLog: null,
      filters: {
        status: '',
        search: ''
      }
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
    setStatus(val) {
      this.filters.status = val
      this.applyFilters()
    },

    applyFilters() {
      this.page = 1
      this.loadLogs()
    },

    resetFilters() {
      this.filters = { status: '', search: '' }
      this.page = 1
      this.loadLogs()
    },

    async loadLogs() {
      this.loading = true
      this.error = ''
      try {
        const res = await admin.getAILogs(this.page, this.filters.status, this.filters.search)
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

    truncate(str, max) {
      if (!str) return '—'
      return str.length > max ? str.slice(0, max) + '…' : str
    },

    openDetail(log) {
      this.detailLog = log
    },

    isJson(str) {
      if (!str) return false
      const s = str.trimStart()
      return s.startsWith('{') || s.startsWith('[')
    },

    prettyJson(str) {
      try {
        return JSON.stringify(JSON.parse(str), null, 2)
      } catch {
        return str
      }
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

.filter-chip:hover { background: rgba(255, 255, 255, 0.07); }

.filter-chip--active {
  background: rgba(99, 179, 237, 0.15);
  border-color: rgba(99, 179, 237, 0.4);
  color: #63b3ed;
  font-weight: 600;
}

.status-ok-chip.filter-chip--active {
  background: rgba(34, 197, 94, 0.12);
  border-color: rgba(34, 197, 94, 0.35);
  color: #22c55e;
}

.status-err-chip.filter-chip--active {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.35);
  color: #ef4444;
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

.apply-btn:hover { background: rgba(99, 102, 241, 0.25); }

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
  padding: 0.625rem 0.875rem;
  text-align: left;
  font-size: 0.7rem;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: rgba(15, 23, 42, 0.6);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.logs-table td {
  padding: 0.5rem 0.875rem;
  color: #cbd5e1;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  vertical-align: middle;
}

.logs-table tbody tr:last-child td { border-bottom: none; }
.logs-table tbody tr:hover td { background: rgba(255, 255, 255, 0.02); }

.td-time {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.7rem;
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

.td-input {
  max-width: 320px;
  color: #94a3b8;
  cursor: default;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.td-provider {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.7rem;
}

.td-latency {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.7rem;
}

.status-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 600;
}

.status-ok {
  background: rgba(34, 197, 94, 0.12);
  color: #22c55e;
}

.status-err {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
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

.detail-btn:hover { color: #94a3b8; background: rgba(255, 255, 255, 0.08); }

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
  max-width: 680px;
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
  flex-shrink: 0;
  min-width: 80px;
}

.detail-row .detail-label { min-width: 80px; }

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
  max-height: 300px;
  overflow-y: auto;
}

.detail-error { color: #f87171; }

.detail-code {
  margin: 0;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 0.5rem;
  padding: 0.75rem;
  overflow: auto;
  font-size: 0.75rem;
  line-height: 1.6;
  color: #a5b4fc;
  font-family: 'Fira Code', 'Cascadia Code', monospace;
  white-space: pre;
  max-height: 300px;
}

.detail-code code { font-family: inherit; }
</style>
