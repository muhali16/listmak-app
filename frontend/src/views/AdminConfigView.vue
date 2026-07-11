<template>
  <div class="admin-container">
    <div class="admin-header">
      <h1 class="admin-title">Konfigurasi</h1>
      <span class="admin-badge">Admin</span>
    </div>

    <div v-if="loading" class="state-block">
      <i class="pi pi-spin pi-spinner"></i>
      <p>Memuat...</p>
    </div>

    <template v-else>
      <!-- Testing mode toggle -->
      <div class="cfg-card" :class="{ 'cfg-card--on': testing }">
        <div class="cfg-info">
          <div class="cfg-title-row">
            <i class="pi pi-exclamation-triangle"></i>
            <h2 class="cfg-title">Mode Testing</h2>
            <span class="cfg-state" :class="testing ? 'cfg-state--on' : 'cfg-state--off'">
              {{ testing ? 'AKTIF' : 'NONAKTIF' }}
            </span>
          </div>
          <p class="cfg-desc">
            Saat aktif, semua data baru (listmak, pesanan, pembayaran) ditandai
            <strong>sandbox</strong> dan hanya muncul di mode testing. Di mode produksi
            data sandbox disembunyikan dari pengguna, tapi tetap terlihat + berlabel di admin.
            Banner kuning "TESTING" tampil di semua halaman.
          </p>
        </div>

        <button
          class="cfg-switch"
          :class="{ 'cfg-switch--on': testing }"
          :disabled="saving"
          role="switch"
          :aria-checked="testing"
          @click="toggle"
        >
          <span class="cfg-knob">
            <i v-if="saving" class="pi pi-spin pi-spinner"></i>
          </span>
        </button>
      </div>

      <p class="cfg-note">
        <i class="pi pi-info-circle"></i>
        Perubahan berlaku langsung untuk semua pengguna. Data sandbox lama tidak dihapus —
        muncul lagi jika mode testing diaktifkan kembali.
      </p>

      <!-- AI model -->
      <div class="cfg-card" style="margin-top: 1rem">
        <div class="cfg-info">
          <div class="cfg-title-row">
            <i class="pi pi-microchip-ai" style="color:#818cf8"></i>
            <h2 class="cfg-title">Model AI (Fireworks)</h2>
          </div>
          <p class="cfg-desc">
            Model yang dipakai untuk parsing pesanan, estimasi harga, dan ringkasan.
            Ubah tanpa redeploy. API key tetap di server (.env).
          </p>
          <div class="cfg-model-row">
            <input
              v-model="model"
              class="cfg-input"
              type="text"
              placeholder="accounts/fireworks/models/..."
              :disabled="savingModel"
            />
            <button class="cfg-save-btn" :disabled="savingModel || !model.trim()" @click="saveModel">
              <i v-if="savingModel" class="pi pi-spin pi-spinner"></i>
              <span>{{ savingModel ? 'Menyimpan...' : 'Simpan' }}</span>
            </button>
          </div>
        </div>
      </div>
    </template>

    <Toast position="top-center" />
  </div>
</template>

<script>
import Toast from 'primevue/toast'
import admin from '../api/admin'

export default {
  name: 'AdminConfigView',
  components: { Toast },

  data() {
    return {
      testing: false,
      model: '',
      loading: false,
      saving: false,
      savingModel: false,
    }
  },

  mounted() {
    this.load()
  },

  methods: {
    async load() {
      this.loading = true
      try {
        const res = await admin.getAdminConfig()
        this.testing = !!res.data?.testing_mode
        this.model = res.data?.fireworks_model || ''
      } catch (err) {
        this.$toast.add({ severity: 'error', summary: 'Gagal', detail: err.message, life: 3000 })
      } finally {
        this.loading = false
      }
    },

    async toggle() {
      const next = !this.testing
      if (!confirm(next
        ? 'Aktifkan mode testing? Data baru akan ditandai sandbox.'
        : 'Nonaktifkan mode testing? Aplikasi kembali ke mode produksi.')) return
      this.saving = true
      try {
        const res = await admin.updateConfig({ testing_mode: next })
        this.testing = !!res.data?.testing_mode
        this.$toast.add({
          severity: 'success',
          summary: 'Tersimpan',
          detail: this.testing ? 'Mode testing AKTIF' : 'Mode produksi',
          life: 3000,
        })
      } catch (err) {
        this.$toast.add({ severity: 'error', summary: 'Gagal', detail: err.message, life: 3000 })
      } finally {
        this.saving = false
      }
    },

    async saveModel() {
      this.savingModel = true
      try {
        const res = await admin.updateConfig({ fireworks_model: this.model.trim() })
        this.model = res.data?.fireworks_model || ''
        this.$toast.add({ severity: 'success', summary: 'Tersimpan', detail: 'Model AI diperbarui', life: 3000 })
      } catch (err) {
        this.$toast.add({ severity: 'error', summary: 'Gagal', detail: err.message, life: 3000 })
      } finally {
        this.savingModel = false
      }
    },
  },
}
</script>

<style scoped>
.admin-container { padding: 1.5rem; max-width: 720px; }

.admin-header { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1.25rem; }
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

.state-block {
  display: flex; flex-direction: column; align-items: center;
  padding: 2.5rem 1rem; color: #64748b; gap: 0.75rem;
}
.state-block > i { font-size: 2rem; color: #3b82f6; }

.cfg-card {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 0.875rem;
  padding: 1.25rem;
}

.cfg-card--on {
  border-color: rgba(250, 204, 21, 0.4);
  background: rgba(250, 204, 21, 0.06);
}

.cfg-info { flex: 1; min-width: 0; }

.cfg-title-row { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.cfg-title-row i { color: #facc15; }
.cfg-title { font-size: 1rem; font-weight: 700; color: #f1f5f9; margin: 0; }

.cfg-state {
  font-size: 0.6875rem; font-weight: 800; letter-spacing: 0.05em;
  padding: 0.1rem 0.45rem; border-radius: 999px;
}
.cfg-state--on { background: rgba(250, 204, 21, 0.18); color: #facc15; }
.cfg-state--off { background: rgba(148, 163, 184, 0.15); color: #94a3b8; }

.cfg-desc { font-size: 0.8125rem; line-height: 1.6; color: #94a3b8; margin: 0; }
.cfg-desc strong { color: #fcd34d; }

.cfg-switch {
  flex-shrink: 0;
  width: 3.25rem;
  height: 1.85rem;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(15, 23, 42, 0.8);
  cursor: pointer;
  padding: 0.15rem;
  display: flex;
  transition: background 0.15s, border-color 0.15s;
}

.cfg-switch--on { background: #facc15; border-color: #facc15; justify-content: flex-end; }
.cfg-switch:disabled { opacity: 0.6; cursor: not-allowed; }

.cfg-knob {
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  background: #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.7rem;
  color: #422006;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.cfg-note {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin-top: 1rem;
  font-size: 0.75rem;
  line-height: 1.55;
  color: #64748b;
}
.cfg-note i { color: #64748b; margin-top: 0.1rem; flex-shrink: 0; }

.cfg-model-row { display: flex; gap: 0.5rem; margin-top: 0.75rem; }

.cfg-input {
  flex: 1;
  min-width: 0;
  padding: 0.5rem 0.75rem;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.8125rem;
  font-family: monospace;
}
.cfg-input:focus { outline: none; border-color: rgba(99, 102, 241, 0.5); }

.cfg-save-btn {
  flex-shrink: 0;
  padding: 0.5rem 1rem;
  background: rgba(99, 102, 241, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 0.5rem;
  color: #818cf8;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
}
.cfg-save-btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
