<template>
  <div class="admin-container">
    <div class="admin-header">
      <h1 class="admin-title">Kelola Listmak</h1>
      <span class="admin-badge">Admin</span>
    </div>

    <!-- Filters -->
    <div class="filter-panel">
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
      <div class="filter-row">
        <span class="filter-section-label">Tanggal:</span>
        <input v-model="filters.start_date" class="date-input" type="date" />
        <span class="date-sep">–</span>
        <input v-model="filters.end_date" class="date-input" type="date" />
        <button class="reset-btn" @click="resetFilters"><i class="pi pi-refresh"></i></button>
      </div>
    </div>

    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat listmak...</p>
    </div>
    <div v-else-if="error" class="state-block state-error">
      <i class="pi pi-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <button class="retry-btn" @click="load">Coba lagi</button>
    </div>

    <template v-else>
      <div class="stats-row">
        <span class="stats-label">Total: {{ total }} listmak</span>
        <span class="stats-label">Halaman {{ page }}</span>
      </div>

      <div v-if="listmaks.length > 0">
        <div v-for="lm in listmaks" :key="lm.id" class="listmak-card">
          <div class="lm-main">
            <div class="lm-info">
              <div class="lm-date">{{ formatDate(lm.date) }}</div>
              <div class="lm-title">{{ lm.title || 'Tanpa judul' }}</div>
              <div class="lm-meta">
                <span v-if="lm.user" class="lm-user">
                  <i class="pi pi-user"></i> {{ lm.user.name }}
                </span>
                <span class="lm-orders"><i class="pi pi-shopping-cart"></i> {{ lm.total_orders }} order</span>
                <span class="lm-amount">Rp {{ formatNum(lm.total_amount) }}</span>
              </div>
            </div>
            <div class="lm-right">
              <span class="status-badge" :class="`status-${lm.status}`">{{ lm.status }}</span>
            </div>
          </div>

          <div class="lm-actions">
            <CustomSelect
              :model-value="lm.status"
              :options="statusSelectOpts"
              @update:model-value="changeStatus(lm, $event)"
            />
            <button class="action-btn action-btn--manage" @click="openManage(lm)">
              <i class="pi pi-cog"></i> Kelola
            </button>
            <button class="action-btn action-btn--del" @click="confirmDelete(lm)">
              <i class="pi pi-trash"></i>
            </button>
          </div>
        </div>
      </div>
      <div v-else class="state-block">
        <p>Tidak ada listmak.</p>
      </div>

      <div v-if="listmaks.length > 0" class="pagination">
        <button class="page-btn" :disabled="page <= 1" @click="changePage(page - 1)">
          <i class="pi pi-chevron-left"></i> Sebelumnya
        </button>
        <button class="page-btn" :disabled="page * limit >= total" @click="changePage(page + 1)">
          Berikutnya <i class="pi pi-chevron-right"></i>
        </button>
      </div>
    </template>
  </div>

  <!-- Manage Modal -->
  <teleport to="body">
    <div v-if="manageTarget" class="detail-backdrop" @click.self="closeManage">
      <div class="detail-modal">
        <div class="detail-header">
          <div>
            <h2 class="detail-title">{{ manageTarget.title || 'Tanpa judul' }}</h2>
            <p class="detail-subtitle">{{ formatDate(manageTarget.date) }}</p>
          </div>
          <button class="detail-close" @click="closeManage"><i class="pi pi-times"></i></button>
        </div>
        <div class="detail-body">

          <!-- Share Link -->
          <div class="manage-section">
            <div class="manage-section-header">
              <span class="manage-section-title"><i class="pi pi-link"></i> Share Link</span>
            </div>
            <div v-if="manageLoading" class="manage-loading">
              <i class="pi pi-spin pi-spinner"></i>
            </div>
            <template v-else>
              <div v-if="activeShares.share_link" class="manage-item">
                <div class="manage-item-info">
                  <span class="manage-item-id">{{ activeShares.share_link.share_id }}</span>
                  <span class="manage-item-exp">Exp: {{ formatDate(activeShares.share_link.expires_at) }}</span>
                </div>
                <button class="manage-del-btn" :disabled="deletingShareLink" @click="doDeleteShareLink">
                  <i v-if="deletingShareLink" class="pi pi-spin pi-spinner"></i>
                  <i v-else class="pi pi-trash"></i>
                </button>
              </div>
              <div v-else class="manage-empty">Tidak ada share link aktif</div>
            </template>
          </div>

          <!-- View Share -->
          <div class="manage-section">
            <div class="manage-section-header">
              <span class="manage-section-title"><i class="pi pi-eye"></i> View Share</span>
            </div>
            <div v-if="manageLoading" class="manage-loading">
              <i class="pi pi-spin pi-spinner"></i>
            </div>
            <template v-else>
              <div v-if="activeShares.view_share" class="manage-item">
                <div class="manage-item-info">
                  <span class="manage-item-id">{{ activeShares.view_share.view_id }}</span>
                  <span class="manage-item-exp">{{ formatDate(activeShares.view_share.created_at) }}</span>
                </div>
                <button class="manage-del-btn" :disabled="deletingViewShare" @click="doDeleteViewShare">
                  <i v-if="deletingViewShare" class="pi pi-spin pi-spinner"></i>
                  <i v-else class="pi pi-trash"></i>
                </button>
              </div>
              <div v-else class="manage-empty">Tidak ada view share</div>
            </template>
          </div>

          <!-- Summary -->
          <div class="manage-section">
            <div class="manage-section-header">
              <span class="manage-section-title"><i class="pi pi-chart-bar"></i> Summary</span>
            </div>
            <div v-if="summaryLoading" class="manage-loading">
              <i class="pi pi-spin pi-spinner"></i>
            </div>
            <template v-else>
              <div v-if="summary" class="manage-item">
                <div class="manage-item-info">
                  <span class="manage-item-id">Est: Rp {{ formatNum(summary.total_estimated) }}</span>
                  <span class="manage-item-exp">Gen: {{ formatDate(summary.generated_at) }}</span>
                </div>
                <button class="manage-del-btn" :disabled="deletingSummary" @click="doDeleteSummary">
                  <i v-if="deletingSummary" class="pi pi-spin pi-spinner"></i>
                  <i v-else class="pi pi-trash"></i>
                </button>
              </div>
              <div v-else class="manage-empty">Belum ada summary</div>
            </template>
          </div>

        </div>
      </div>
    </div>

    <!-- Delete Listmak confirm -->
    <div v-if="deleteTarget" class="detail-backdrop" @click.self="deleteTarget = null">
      <div class="detail-modal detail-modal--sm">
        <div class="detail-header">
          <h2 class="detail-title">Hapus Listmak?</h2>
          <button class="detail-close" @click="deleteTarget = null"><i class="pi pi-times"></i></button>
        </div>
        <div class="detail-body">
          <p class="confirm-text">
            <strong>{{ deleteTarget.title || 'Tanpa judul' }}</strong><br/>
            <span class="confirm-sub">{{ formatDate(deleteTarget.date) }} · {{ deleteTarget.total_orders }} order</span>
          </p>
          <p class="confirm-warn">Semua order di listmak ini juga akan dihapus.</p>
          <div class="modal-actions">
            <button class="cancel-btn" @click="deleteTarget = null">Batal</button>
            <button class="del-confirm-btn" :disabled="deleting" @click="doDelete">
              <i v-if="deleting" class="pi pi-spin pi-spinner"></i>
              {{ deleting ? 'Menghapus...' : 'Hapus' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script>
import adminApi from '../api/admin'
import CustomSelect from '../components/CustomSelect.vue'

export default {
  name: 'AdminListmaksView',
  components: { CustomSelect },
  data() {
    return {
      listmaks: [],
      total: 0,
      page: 1,
      limit: 20,
      loading: true,
      error: null,
      filters: { status: '', start_date: '', end_date: '' },
      statusOptions: [
        { value: '', label: 'Semua' },
        { value: 'active', label: 'Active' },
        { value: 'completed', label: 'Completed' },
        { value: 'cancelled', label: 'Cancelled' }
      ],
      statusSelectOpts: [
        { value: 'active', label: 'Active', badgeClass: 'badge-active' },
        { value: 'completed', label: 'Completed', badgeClass: 'badge-completed' },
        { value: 'cancelled', label: 'Cancelled', badgeClass: 'badge-cancelled' }
      ],
      // manage modal
      manageTarget: null,
      manageLoading: false,
      summaryLoading: false,
      activeShares: { share_link: null, view_share: null },
      summary: null,
      deletingShareLink: false,
      deletingViewShare: false,
      deletingSummary: false,
      // delete confirm
      deleteTarget: null,
      deleting: false
    }
  },
  mounted() {
    this.load()
  },
  methods: {
    async load() {
      this.loading = true
      this.error = null
      try {
        const params = { page: this.page, limit: this.limit }
        if (this.filters.status) params.status = this.filters.status
        if (this.filters.start_date) params.start_date = this.filters.start_date
        if (this.filters.end_date) params.end_date = this.filters.end_date
        const res = await adminApi.getAllListmaks(params)
        this.listmaks = res.data || []
        this.total = res.pagination?.total || 0
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },
    setStatus(val) {
      this.filters.status = val
      this.page = 1
      this.load()
    },
    resetFilters() {
      this.filters = { status: '', start_date: '', end_date: '' }
      this.page = 1
      this.load()
    },
    changePage(p) {
      this.page = p
      this.load()
    },
    async changeStatus(lm, newStatus) {
      if (newStatus === lm.status) return
      try {
        await adminApi.updateListmakStatus(lm.id, newStatus)
        lm.status = newStatus
        this.$toast.add({ severity: 'success', summary: 'Status diperbarui', life: 2000 })
      } catch (e) {
        this.$toast.add({ severity: 'error', summary: 'Error', detail: e.message, life: 3000 })
      }
    },
    async openManage(lm) {
      this.manageTarget = lm
      this.activeShares = { share_link: null, view_share: null }
      this.summary = null
      this.manageLoading = true
      this.summaryLoading = true
      try {
        const res = await adminApi.getActiveShares(lm.id)
        this.activeShares = res.data || { share_link: null, view_share: null }
      } catch { /* no shares */ }
      finally { this.manageLoading = false }
      try {
        const res = await adminApi.getListmakSummary(lm.id)
        this.summary = res.data || null
      } catch { /* no summary */ }
      finally { this.summaryLoading = false }
    },
    closeManage() {
      this.manageTarget = null
    },
    async doDeleteShareLink() {
      if (!this.activeShares.share_link) return
      this.deletingShareLink = true
      try {
        await adminApi.deleteShareLink(this.activeShares.share_link.id)
        this.activeShares.share_link = null
        this.$toast.add({ severity: 'success', summary: 'Share link dihapus', life: 2000 })
      } catch (e) {
        this.$toast.add({ severity: 'error', summary: 'Error', detail: e.message, life: 3000 })
      } finally { this.deletingShareLink = false }
    },
    async doDeleteViewShare() {
      if (!this.activeShares.view_share) return
      this.deletingViewShare = true
      try {
        await adminApi.deleteViewShare(this.activeShares.view_share.id)
        this.activeShares.view_share = null
        this.$toast.add({ severity: 'success', summary: 'View share dihapus', life: 2000 })
      } catch (e) {
        this.$toast.add({ severity: 'error', summary: 'Error', detail: e.message, life: 3000 })
      } finally { this.deletingViewShare = false }
    },
    async doDeleteSummary() {
      if (!this.summary) return
      this.deletingSummary = true
      try {
        await adminApi.deleteSummary(this.manageTarget.id)
        this.summary = null
        this.$toast.add({ severity: 'success', summary: 'Summary dihapus', life: 2000 })
      } catch (e) {
        this.$toast.add({ severity: 'error', summary: 'Error', detail: e.message, life: 3000 })
      } finally { this.deletingSummary = false }
    },
    confirmDelete(lm) {
      this.deleteTarget = lm
    },
    async doDelete() {
      this.deleting = true
      try {
        await adminApi.deleteListmak(this.deleteTarget.id)
        this.$toast.add({ severity: 'success', summary: 'Listmak dihapus', life: 2500 })
        this.deleteTarget = null
        await this.load()
      } catch (e) {
        this.$toast.add({ severity: 'error', summary: 'Error', detail: e.message, life: 3000 })
      } finally { this.deleting = false }
    },
    formatDate(dt) {
      if (!dt) return '—'
      return new Date(dt).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
    },
    formatNum(n) {
      return new Intl.NumberFormat('id-ID').format(n || 0)
    }
  }
}
</script>

<style scoped>
.admin-container {
  padding: 1rem;
  padding-bottom: 5rem;
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
}

.admin-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  padding: 0.25rem 0.625rem;
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border-radius: 1rem;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.filter-panel {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.75rem;
  padding: 0.875rem;
  margin-bottom: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.filter-section-label {
  font-size: 0.75rem;
  color: #64748b;
  white-space: nowrap;
}

.filter-chip {
  padding: 0.3125rem 0.75rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: transparent;
  color: #94a3b8;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s;
}

.filter-chip:hover { background: rgba(255, 255, 255, 0.05); color: #f1f5f9; }
.filter-chip--active { background: rgba(59, 130, 246, 0.15); color: #3b82f6; border-color: rgba(59, 130, 246, 0.3); }

.date-input {
  padding: 0.3125rem 0.5rem;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.375rem;
  color: #f1f5f9;
  font-size: 0.75rem;
  color-scheme: dark;
}

.date-sep { color: #64748b; font-size: 0.75rem; }

.reset-btn {
  padding: 0.3125rem 0.5rem;
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.375rem;
  color: #64748b;
  cursor: pointer;
  font-size: 0.75rem;
  transition: color 0.15s;
}

.reset-btn:hover { color: #f1f5f9; }

.state-block {
  text-align: center;
  padding: 3rem 1rem;
  color: #64748b;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.state-block i { font-size: 2rem; }
.state-error { color: #f87171; }

.retry-btn {
  padding: 0.5rem 1rem;
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
}

.stats-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.stats-label { font-size: 0.8125rem; color: #64748b; }

/* Listmak Cards */
.listmak-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.75rem;
  padding: 0.875rem 1rem;
  margin-bottom: 0.625rem;
}

.lm-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.lm-date {
  font-size: 0.75rem;
  color: #64748b;
  margin-bottom: 0.25rem;
}

.lm-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.375rem;
}

.lm-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.lm-user, .lm-orders {
  font-size: 0.75rem;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.lm-amount {
  font-size: 0.75rem;
  color: #34d399;
  font-weight: 500;
}

.status-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  padding: 0.25rem 0.625rem;
  border-radius: 1rem;
  white-space: nowrap;
}

.status-active { background: rgba(52, 211, 153, 0.15); color: #34d399; }
.status-completed { background: rgba(59, 130, 246, 0.15); color: #3b82f6; }
.status-cancelled { background: rgba(239, 68, 68, 0.15); color: #f87171; }

/* Badge classes exposed to CustomSelect via :class binding */
:deep(.badge-active) { background: rgba(52, 211, 153, 0.15); color: #34d399; }
:deep(.badge-completed) { background: rgba(59, 130, 246, 0.15); color: #3b82f6; }
:deep(.badge-cancelled) { background: rgba(239, 68, 68, 0.15); color: #f87171; }

.lm-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}


.action-btn {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border-radius: 0.375rem;
  border: none;
  cursor: pointer;
  font-size: 0.8125rem;
  font-weight: 500;
  transition: background 0.15s;
  white-space: nowrap;
}

.action-btn--manage {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.action-btn--manage:hover { background: rgba(59, 130, 246, 0.2); }

.action-btn--del {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  padding: 0.375rem 0.625rem;
}

.action-btn--del:hover { background: rgba(239, 68, 68, 0.2); }

/* Pagination */
.pagination {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  margin-top: 1.5rem;
}

.page-btn {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.5rem;
  color: #94a3b8;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.15s;
}

.page-btn:hover:not(:disabled) { color: #f1f5f9; border-color: rgba(255, 255, 255, 0.15); }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* Modal */
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
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 1rem;
  width: 100%;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
}

.detail-modal--sm { max-width: 360px; }

.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  position: sticky;
  top: 0;
  background: #1e293b;
}

.detail-title {
  font-size: 1rem;
  font-weight: 600;
  color: #f1f5f9;
}

.detail-subtitle {
  font-size: 0.75rem;
  color: #64748b;
  margin-top: 0.125rem;
}

.detail-close {
  width: 28px;
  height: 28px;
  background: rgba(255, 255, 255, 0.05);
  border: none;
  border-radius: 0.375rem;
  color: #94a3b8;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  flex-shrink: 0;
}

.detail-body {
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

/* Manage sections */
.manage-section {
  background: rgba(15, 23, 42, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.625rem;
  overflow: hidden;
}

.manage-section-header {
  padding: 0.625rem 0.875rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  background: rgba(15, 23, 42, 0.3);
}

.manage-section-title {
  font-size: 0.75rem;
  font-weight: 600;
  color: #94a3b8;
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.manage-loading {
  padding: 0.875rem;
  color: #64748b;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.manage-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 0.875rem;
  gap: 0.75rem;
}

.manage-item-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.manage-item-id {
  font-size: 0.8125rem;
  color: #f1f5f9;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.manage-item-exp {
  font-size: 0.6875rem;
  color: #64748b;
}

.manage-empty {
  padding: 0.75rem 0.875rem;
  font-size: 0.8125rem;
  color: #475569;
  font-style: italic;
}

.manage-del-btn {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  background: rgba(239, 68, 68, 0.1);
  border: none;
  border-radius: 0.375rem;
  color: #f87171;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8125rem;
  transition: background 0.15s;
}

.manage-del-btn:hover:not(:disabled) { background: rgba(239, 68, 68, 0.2); }
.manage-del-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* Delete confirm */
.modal-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 0.25rem;
}

.cancel-btn {
  padding: 0.5rem 1rem;
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #94a3b8;
  font-size: 0.875rem;
  cursor: pointer;
}

.del-confirm-btn {
  padding: 0.5rem 1.25rem;
  background: #ef4444;
  border: none;
  border-radius: 0.5rem;
  color: white;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.del-confirm-btn:disabled { opacity: 0.6; cursor: not-allowed; }

.confirm-text {
  font-size: 0.9375rem;
  color: #f1f5f9;
  line-height: 1.6;
}

.confirm-sub { color: #94a3b8; }
.confirm-warn { font-size: 0.8125rem; color: #f87171; }

@media (min-width: 768px) {
  .admin-container { padding: 1.5rem 2rem; padding-bottom: 2rem; }
}
</style>
