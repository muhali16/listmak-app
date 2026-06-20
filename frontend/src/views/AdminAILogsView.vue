<template>
  <div class="admin-container">
    <div class="admin-header">
      <h1 class="admin-title">AI Request Logs</h1>
      <span class="admin-badge">Admin</span>
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

      <div class="logs-table-wrap">
        <table class="logs-table">
          <thead>
            <tr>
              <th>Waktu</th>
              <th>Input</th>
              <th>Output</th>
              <th>Provider</th>
              <th>Latency</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td class="td-time">{{ formatTime(log.created_at) }}</td>
              <td class="td-input" :title="log.input">{{ truncate(log.input, 60) }}</td>
              <td class="td-output">{{ log.output || '—' }}</td>
              <td class="td-provider">{{ log.provider }}</td>
              <td class="td-latency">{{ log.latency_ms }}ms</td>
              <td>
                <span class="status-badge" :class="log.status === 'success' ? 'status-ok' : 'status-err'">
                  {{ log.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="logs.length === 0" class="state-block">
        <p>Belum ada AI logs.</p>
      </div>

      <div class="pagination">
        <button class="page-btn" :disabled="page <= 1" @click="changePage(page - 1)">
          <i class="pi pi-chevron-left"></i> Sebelumnya
        </button>
        <button class="page-btn" :disabled="logs.length < 50" @click="changePage(page + 1)">
          Berikutnya <i class="pi pi-chevron-right"></i>
        </button>
      </div>
    </template>
  </div>
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
      error: ''
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
    async loadLogs() {
      this.loading = true
      this.error = ''
      try {
        const res = await admin.getAILogs(this.page)
        if (res.success && res.data) {
          this.logs = res.data.logs || []
          this.total = res.data.total || 0
        }
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
  margin-bottom: 1.5rem;
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

.logs-table tbody tr:last-child td {
  border-bottom: none;
}

.logs-table tbody tr:hover td {
  background: rgba(255, 255, 255, 0.02);
}

.td-time {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.75rem;
}

.td-input {
  max-width: 280px;
  color: #94a3b8;
  cursor: default;
}

.td-output {
  max-width: 160px;
  color: #e2e8f0;
}

.td-provider {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.75rem;
}

.td-latency {
  white-space: nowrap;
  color: #64748b;
  font-size: 0.75rem;
}

.status-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  font-size: 0.75rem;
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

.page-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.page-btn:not(:disabled):hover {
  background: rgba(30, 41, 59, 0.9);
  color: #cbd5e1;
}

.state-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2.5rem 1rem;
  color: #64748b;
  gap: 0.75rem;
}

.state-block > i {
  font-size: 2rem;
  color: #3b82f6;
}

.state-error > i {
  color: #ef4444;
}

.retry-btn {
  padding: 0.5rem 1rem;
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 0.5rem;
  color: #3b82f6;
  font-size: 0.875rem;
  cursor: pointer;
}
</style>
