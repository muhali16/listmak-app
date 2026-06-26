<template>
  <div class="admin-container">
    <div class="admin-header">
      <h1 class="admin-title">Price Catalog</h1>
      <span class="admin-badge">Admin</span>
    </div>

    <div class="toolbar">
      <div class="filter-search-wrap">
        <i class="pi pi-search filter-search-icon"></i>
        <input v-model="search" class="filter-search" type="search" placeholder="Cari vendor / item..." />
      </div>
      <button class="add-btn" @click="openAdd">
        <i class="pi pi-plus"></i> Tambah Entry
      </button>
    </div>

    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat katalog...</p>
    </div>
    <div v-else-if="error" class="state-block state-error">
      <i class="pi pi-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <button class="retry-btn" @click="load">Coba lagi</button>
    </div>

    <template v-else>
      <div class="stats-row">
        <span class="stats-label">{{ filtered.length }} dari {{ entries.length }} entries</span>
      </div>

      <div v-if="filtered.length > 0" class="logs-table-wrap">
        <table class="logs-table">
          <thead>
            <tr>
              <th>Vendor</th>
              <th>Item</th>
              <th class="th-price">Harga</th>
              <th>Diperbarui</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="entry in filtered" :key="entry.id">
              <td class="td-vendor">{{ entry.vendor_name }}</td>
              <td>{{ entry.item_name }}</td>
              <td class="td-price">{{ formatPrice(entry.price) }}</td>
              <td class="td-time">{{ formatDate(entry.updated_at) }}</td>
              <td class="td-actions">
                <button class="action-btn action-btn--edit" @click="openEdit(entry)" title="Edit">
                  <i class="pi pi-pencil"></i>
                </button>
                <button class="action-btn action-btn--del" @click="confirmDelete(entry)" title="Hapus">
                  <i class="pi pi-trash"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="state-block">
        <p>Tidak ada entry katalog.</p>
      </div>
    </template>
  </div>

  <!-- Add/Edit Modal -->
  <teleport to="body">
    <div v-if="modal" class="detail-backdrop" @click.self="modal = null">
      <div class="detail-modal">
        <div class="detail-header">
          <h2 class="detail-title">{{ modal.id ? 'Edit Entry' : 'Tambah Entry' }}</h2>
          <button class="detail-close" @click="modal = null"><i class="pi pi-times"></i></button>
        </div>
        <div class="detail-body">
          <div class="form-group">
            <label class="form-label">Vendor</label>
            <input v-model="modal.vendor_name" class="form-input" placeholder="Nama vendor" />
          </div>
          <div class="form-group">
            <label class="form-label">Item</label>
            <input v-model="modal.item_name" class="form-input" placeholder="Nama item/menu" />
          </div>
          <div class="form-group">
            <label class="form-label">Harga (Rp)</label>
            <PriceInput v-model="modal.price" class="form-input" />
          </div>
          <div class="modal-actions">
            <button class="cancel-btn" @click="modal = null">Batal</button>
            <button class="save-btn" :disabled="saving" @click="save">
              <i v-if="saving" class="pi pi-spin pi-spinner"></i>
              {{ saving ? 'Menyimpan...' : 'Simpan' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete confirm -->
    <div v-if="deleteTarget" class="detail-backdrop" @click.self="deleteTarget = null">
      <div class="detail-modal detail-modal--sm">
        <div class="detail-header">
          <h2 class="detail-title">Hapus Entry?</h2>
          <button class="detail-close" @click="deleteTarget = null"><i class="pi pi-times"></i></button>
        </div>
        <div class="detail-body">
          <p class="confirm-text">
            <strong>{{ deleteTarget.vendor_name }}</strong> — {{ deleteTarget.item_name }}<br/>
            <span class="confirm-sub">Rp {{ formatPrice(deleteTarget.price) }}</span>
          </p>
          <p class="confirm-warn">Entry akan dihapus permanen dari katalog harga.</p>
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
import PriceInput from '../components/PriceInput.vue'

export default {
  name: 'AdminPriceCatalogView',
  components: { PriceInput },
  data() {
    return {
      entries: [],
      search: '',
      loading: true,
      error: null,
      modal: null,
      saving: false,
      deleteTarget: null,
      deleting: false
    }
  },
  computed: {
    filtered() {
      if (!this.search.trim()) return this.entries
      const q = this.search.toLowerCase()
      return this.entries.filter(e =>
        e.vendor_name.toLowerCase().includes(q) || e.item_name.toLowerCase().includes(q)
      )
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
        const res = await adminApi.getPriceCatalog()
        this.entries = res.data || []
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },
    openAdd() {
      this.modal = { vendor_name: '', item_name: '', price: 0 }
    },
    openEdit(entry) {
      this.modal = { ...entry }
    },
    async save() {
      if (!this.modal.vendor_name.trim() || !this.modal.item_name.trim()) return
      this.saving = true
      try {
        await adminApi.upsertPriceCatalog([{
          vendor_name: this.modal.vendor_name.trim(),
          item_name: this.modal.item_name.trim(),
          price: this.modal.price
        }])
        this.$toast.add({ severity: 'success', summary: 'Tersimpan', detail: 'Entry berhasil disimpan', life: 2500 })
        this.modal = null
        await this.load()
      } catch (e) {
        this.$toast.add({ severity: 'error', summary: 'Error', detail: e.message, life: 3000 })
      } finally {
        this.saving = false
      }
    },
    confirmDelete(entry) {
      this.deleteTarget = entry
    },
    async doDelete() {
      this.deleting = true
      try {
        await adminApi.deletePriceCatalog(this.deleteTarget.id)
        this.$toast.add({ severity: 'success', summary: 'Dihapus', detail: 'Entry dihapus dari katalog', life: 2500 })
        this.deleteTarget = null
        await this.load()
      } catch (e) {
        this.$toast.add({ severity: 'error', summary: 'Error', detail: e.message, life: 3000 })
      } finally {
        this.deleting = false
      }
    },
    formatPrice(price) {
      return new Intl.NumberFormat('id-ID').format(price)
    },
    formatDate(dt) {
      if (!dt) return '—'
      return new Date(dt).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
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

.toolbar {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.filter-search-wrap {
  position: relative;
  flex: 1;
  min-width: 0;
}

.filter-search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: #64748b;
  font-size: 0.875rem;
}

.filter-search {
  width: 100%;
  padding: 0.625rem 0.75rem 0.625rem 2.25rem;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.875rem;
  box-sizing: border-box;
}

.filter-search::placeholder { color: #64748b; }
.filter-search:focus { outline: none; border-color: rgba(59, 130, 246, 0.4); }

.add-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.15s;
}

.add-btn:hover { background: rgba(59, 130, 246, 0.25); }

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

.stats-label {
  font-size: 0.8125rem;
  color: #64748b;
}

.logs-table-wrap {
  overflow-x: auto;
  border-radius: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
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
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  background: rgba(15, 23, 42, 0.6);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  white-space: nowrap;
}

.logs-table td {
  padding: 0.75rem 1rem;
  color: #cbd5e1;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  vertical-align: middle;
}

.logs-table tbody tr:hover { background: rgba(255, 255, 255, 0.02); }
.logs-table tbody tr:last-child td { border-bottom: none; }

.td-vendor { font-weight: 500; color: #f1f5f9; }
.td-price { font-weight: 600; color: #34d399; text-align: right; }
.th-price { text-align: right; }
.td-time { color: #64748b; font-size: 0.75rem; white-space: nowrap; }

.td-actions {
  display: flex;
  gap: 0.375rem;
  justify-content: flex-end;
}

.action-btn {
  width: 30px;
  height: 30px;
  border-radius: 0.375rem;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8125rem;
  transition: background 0.15s;
}

.action-btn--edit {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.action-btn--edit:hover { background: rgba(59, 130, 246, 0.2); }

.action-btn--del {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
}

.action-btn--del:hover { background: rgba(239, 68, 68, 0.2); }

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
  max-width: 440px;
  overflow: hidden;
}

.detail-modal--sm { max-width: 360px; }

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.detail-title {
  font-size: 1rem;
  font-weight: 600;
  color: #f1f5f9;
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
}

.detail-body {
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: #94a3b8;
}

.form-input {
  padding: 0.625rem 0.75rem;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.875rem;
}

.form-input:focus { outline: none; border-color: rgba(59, 130, 246, 0.4); }

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

.save-btn {
  padding: 0.5rem 1.25rem;
  background: #3b82f6;
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

.save-btn:disabled { opacity: 0.6; cursor: not-allowed; }

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

.confirm-sub { color: #34d399; font-weight: 600; }

.confirm-warn {
  font-size: 0.8125rem;
  color: #f87171;
}

@media (min-width: 768px) {
  .admin-container { padding: 1.5rem 2rem; padding-bottom: 2rem; }
}
</style>
