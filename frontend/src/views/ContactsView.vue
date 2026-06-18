<template>
  <div class="contacts-container">
    <!-- Header -->
    <header class="page-header">
      <h1 class="page-title">Kontak</h1>
      <p class="page-subtitle">Kelola daftar kontak peserta</p>
    </header>

    <!-- Tabs -->
    <section class="tabs-section">
      <div class="tabs">
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'list' }"
          @click="activeTab = 'list'"
        >
          <i class="pi pi-list"></i>
          <span>Daftar</span>
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'import' }"
          @click="activeTab = 'import'"
        >
          <i class="pi pi-upload"></i>
          <span>Import</span>
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'add' }"
          @click="activeTab = 'add'"
        >
          <i class="pi pi-plus"></i>
          <span>Tambah</span>
        </button>
      </div>
    </section>

    <!-- List Tab -->
    <section v-if="activeTab === 'list'" class="content-section">
      <div v-if="contacts.length > 0" class="contacts-list">
        <div 
          v-for="(contact, index) in contacts" 
          :key="index" 
          class="contact-card"
        >
          <div class="contact-avatar">
            {{ getInitials(contact.name) }}
          </div>
          <div class="contact-info">
            <span class="contact-name">{{ contact.name }}</span>
            <span class="contact-phone">{{ contact.phone || 'No phone' }}</span>
          </div>
          <Button 
            icon="pi pi-trash" 
            severity="danger" 
            text 
            rounded
            @click="removeContact(index)"
            class="delete-btn"
          />
        </div>
      </div>
      
      <div v-else class="empty-state">
        <div class="empty-icon">
          <i class="pi pi-users"></i>
        </div>
        <h3>Belum Ada Kontak</h3>
        <p>Import dari file atau tambah manual</p>
      </div>
    </section>

    <!-- Import Tab -->
    <section v-if="activeTab === 'import'" class="content-section">
      <!-- Import dari Kontak HP -->
      <div class="import-card import-contacts-card">
        <div class="import-icon contacts-icon">
          <i class="pi pi-mobile"></i>
        </div>
        <h3>Import dari Kontak HP</h3>
        <p>Pilih langsung dari daftar kontak di handphone</p>
        
        <Button 
          v-if="isContactPickerSupported"
          label="Pilih Kontak"
          icon="pi pi-users"
          @click="importFromPhoneContacts"
          class="contacts-import-btn"
          :loading="isImportingContacts"
        />
        
        <div v-else class="not-supported-notice">
          <i class="pi pi-info-circle"></i>
          <span>Import kontak tidak didukung di browser ini. Gunakan Chrome/Edge di Android atau import dari file.</span>
        </div>
      </div>

      <!-- Import dari File -->
      <div class="import-card">
        <div class="import-icon">
          <i class="pi pi-file-excel"></i>
        </div>
        <h3>Import dari File</h3>
        <p>Upload file Excel (.xlsx) atau CSV</p>
        
        <input 
          type="file" 
          ref="fileInput"
          accept=".csv,.xlsx,.xls"
          @change="handleFileUpload"
          class="file-input"
        />
        
        <Button 
          label="Pilih File"
          icon="pi pi-upload"
          @click="$refs.fileInput.click()"
          class="upload-btn"
        />
        
        <div class="import-format">
          <p><strong>Format yang didukung:</strong></p>
          <ul>
            <li>Kolom 1: Nama</li>
            <li>Kolom 2: No. Telepon (opsional)</li>
          </ul>
        </div>
      </div>
    </section>

    <!-- Add Tab -->
    <section v-if="activeTab === 'add'" class="content-section">
      <!-- Single Add -->
      <div class="add-card">
        <h3>Tambah Satu</h3>
        <div class="form-group">
          <label>Nama</label>
          <input 
            type="text"
            v-model="newContact.name" 
            placeholder="Nama kontak" 
            class="form-input" 
          />
        </div>
        <div class="form-group">
          <label>No. Telepon</label>
          <input 
            type="text"
            v-model="newContact.phone" 
            placeholder="08xxxxxxxxxx" 
            class="form-input" 
          />
        </div>
        <Button 
          label="Tambah"
          icon="pi pi-plus"
          size="small"
          @click="addSingleContact"
          class="add-btn"
        />
      </div>

      <!-- Bulk Add -->
      <div class="add-card">
        <h3>Tambah Banyak</h3>
        <p class="add-hint">Format: <strong>Nama, No HP</strong> (satu per baris)</p>
        <p class="add-hint-example">Contoh: Budi, 081234567890</p>
        <textarea
          v-model="bulkInput"
          placeholder="Budi, 081234567890
Ani, 082345678901
Citra, 083456789012"
          rows="5"
          class="bulk-textarea"
        ></textarea>
        <Button 
          label="Tambah Semua"
          icon="pi pi-plus-circle"
          size="small"
          @click="addBulkContacts"
          class="add-btn"
        />
      </div>
    </section>
  </div>
</template>

<script>
import Button from 'primevue/button'

export default {
  name: 'ContactsView',
  components: {
    Button
  },
  data() {
    return {
      activeTab: 'list',
      contacts: [],
      newContact: {
        name: '',
        phone: ''
      },
      bulkInput: '',
      isImportingContacts: false
    }
  },
  computed: {
    isContactPickerSupported() {
      return 'contacts' in navigator && 'ContactsManager' in window
    }
  },
  mounted() {
    this.loadContacts()
  },
  methods: {
    loadContacts() {
      const saved = localStorage.getItem('listmak_contacts')
      this.contacts = saved ? JSON.parse(saved) : []
    },
    saveContacts() {
      localStorage.setItem('listmak_contacts', JSON.stringify(this.contacts))
    },
    getInitials(name) {
      return name
        .split(' ')
        .map(word => word[0])
        .join('')
        .toUpperCase()
        .slice(0, 2)
    },
    addSingleContact() {
      if (!this.newContact.name.trim()) {
        this.$toast.add({
          severity: 'warn',
          summary: 'Perhatian',
          detail: 'Nama tidak boleh kosong',
          life: 3000
        })
        return
      }

      this.contacts.push({
        name: this.newContact.name.trim(),
        phone: this.newContact.phone.trim()
      })
      this.saveContacts()
      
      this.newContact = { name: '', phone: '' }
      this.$toast.add({
        severity: 'success',
        summary: 'Berhasil',
        detail: 'Kontak berhasil ditambahkan',
        life: 2000
      })
    },
    addBulkContacts() {
      if (!this.bulkInput.trim()) {
        this.$toast.add({
          severity: 'warn',
          summary: 'Perhatian',
          detail: 'Masukkan daftar nama terlebih dahulu',
          life: 3000
        })
        return
      }

      const lines = this.bulkInput.split('\n').filter(n => n.trim())
      let addedCount = 0
      
      for (const line of lines) {
        // Parse format: "Nama, No HP" atau "Nama"
        const parts = line.split(',').map(p => p.trim())
        const name = parts[0]
        const phone = parts[1] || ''
        
        if (name) {
          this.contacts.push({
            name: name,
            phone: phone
          })
          addedCount++
        }
      }
      
      this.saveContacts()
      this.bulkInput = ''
      
      this.$toast.add({
        severity: 'success',
        summary: 'Berhasil',
        detail: `${addedCount} kontak berhasil ditambahkan`,
        life: 3000
      })
      
      this.activeTab = 'list'
    },
    removeContact(index) {
      this.contacts.splice(index, 1)
      this.saveContacts()
      this.$toast.add({
        severity: 'info',
        summary: 'Dihapus',
        detail: 'Kontak berhasil dihapus',
        life: 2000
      })
    },
    handleFileUpload(event) {
      const file = event.target.files[0]
      if (!file) return

      const reader = new FileReader()
      
      if (file.name.endsWith('.csv')) {
        reader.onload = (e) => {
          const text = e.target.result
          const lines = text.split('\n').filter(l => l.trim())
          
          // Skip header row if exists
          const startIndex = lines[0].toLowerCase().includes('nama') ? 1 : 0
          
          for (let i = startIndex; i < lines.length; i++) {
            const cols = lines[i].split(',')
            if (cols[0]) {
              this.contacts.push({
                name: cols[0].trim().replace(/"/g, ''),
                phone: cols[1] ? cols[1].trim().replace(/"/g, '') : ''
              })
            }
          }
          
          this.saveContacts()
          this.$toast.add({
            severity: 'success',
            summary: 'Berhasil',
            detail: `${lines.length - startIndex} kontak berhasil diimport`,
            life: 3000
          })
          this.activeTab = 'list'
        }
        reader.readAsText(file)
      } else {
        this.$toast.add({
          severity: 'warn',
          summary: 'Format Tidak Didukung',
          detail: 'Silakan gunakan file CSV untuk saat ini',
          life: 5000
        })
      }
      
      // Reset input
      event.target.value = ''
    },
    async importFromPhoneContacts() {
      if (!this.isContactPickerSupported) {
        this.$toast.add({
          severity: 'error',
          summary: 'Tidak Didukung',
          detail: 'Browser tidak mendukung import kontak',
          life: 3000
        })
        return
      }

      this.isImportingContacts = true
      
      try {
        const props = ['name', 'tel']
        const opts = { multiple: true }
        
        const contacts = await navigator.contacts.select(props, opts)
        
        if (contacts && contacts.length > 0) {
          let addedCount = 0
          
          for (const contact of contacts) {
            const name = contact.name?.[0] || ''
            const phone = contact.tel?.[0] || ''
            
            if (name) {
              // Check for duplicate
              const exists = this.contacts.some(
                c => c.name.toLowerCase() === name.toLowerCase() && c.phone === phone
              )
              
              if (!exists) {
                this.contacts.push({ name, phone })
                addedCount++
              }
            }
          }
          
          if (addedCount > 0) {
            this.saveContacts()
            this.$toast.add({
              severity: 'success',
              summary: 'Berhasil',
              detail: `${addedCount} kontak berhasil diimport`,
              life: 3000
            })
            this.activeTab = 'list'
          } else {
            this.$toast.add({
              severity: 'info',
              summary: 'Info',
              detail: 'Semua kontak yang dipilih sudah ada',
              life: 3000
            })
          }
        }
      } catch (error) {
        if (error.name !== 'NotAllowedError') {
          this.$toast.add({
            severity: 'error',
            summary: 'Gagal',
            detail: 'Terjadi kesalahan saat import kontak',
            life: 3000
          })
        }
      } finally {
        this.isImportingContacts = false
      }
    }
  }
}
</script>

<style scoped>
.contacts-container {
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

/* Tabs */
.tabs-section {
  margin-bottom: 1.25rem;
}

.tabs {
  display: flex;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.75rem;
  padding: 0.25rem;
  gap: 0.25rem;
}

.tab-btn {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  padding: 0.625rem 0.5rem;
  background: transparent;
  border: none;
  color: #64748b;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.6875rem;
}

.tab-btn i {
  font-size: 1.125rem;
}

.tab-btn.active {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

/* Contacts List */
.contacts-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.contact-card {
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.75rem;
  padding: 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.contact-avatar {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 0.75rem;
  color: white;
  flex-shrink: 0;
}

.contact-info {
  flex: 1;
  min-width: 0;
}

.contact-name {
  display: block;
  font-weight: 500;
  color: #f1f5f9;
  font-size: 0.9375rem;
}

.contact-phone {
  font-size: 0.75rem;
  color: #64748b;
}

.delete-btn {
  color: #64748b !important;
}

.delete-btn:hover {
  color: #ef4444 !important;
  background: rgba(239, 68, 68, 0.1) !important;
}

/* Import Card */
.import-card {
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.75rem;
  padding: 1.5rem;
  text-align: center;
  border: 1px solid rgba(255, 255, 255, 0.05);
  margin-bottom: 1rem;
}

.import-contacts-card {
  border: 1px solid rgba(59, 130, 246, 0.2);
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.8), rgba(59, 130, 246, 0.05));
}

.contacts-icon {
  background: rgba(59, 130, 246, 0.15) !important;
}

.contacts-icon i {
  color: #3b82f6 !important;
}

.contacts-import-btn {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8) !important;
  border: none !important;
}

.not-supported-notice {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  text-align: left;
  background: rgba(234, 179, 8, 0.1);
  border: 1px solid rgba(234, 179, 8, 0.2);
  border-radius: 0.5rem;
  padding: 0.75rem;
  margin-top: 0.5rem;
}

.not-supported-notice i {
  color: #eab308;
  font-size: 1rem;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.not-supported-notice span {
  font-size: 0.75rem;
  color: #94a3b8;
  line-height: 1.4;
}

.import-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 1rem;
  background: rgba(34, 197, 94, 0.15);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.import-icon i {
  font-size: 1.75rem;
  color: #22c55e;
}

.import-card h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.25rem;
}

.import-card > p {
  font-size: 0.8125rem;
  color: #64748b;
  margin-bottom: 1rem;
}

.file-input {
  display: none;
}

.upload-btn {
  margin-bottom: 1.5rem;
  background: linear-gradient(135deg, #22c55e, #16a34a) !important;
  border: none !important;
}

.import-format {
  text-align: left;
  background: rgba(15, 23, 42, 0.5);
  border-radius: 0.5rem;
  padding: 0.75rem;
}

.import-format p {
  font-size: 0.75rem;
  color: #94a3b8;
  margin-bottom: 0.5rem;
}

.import-format ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.import-format li {
  font-size: 0.6875rem;
  color: #64748b;
  padding-left: 1rem;
  position: relative;
}

.import-format li::before {
  content: '•';
  position: absolute;
  left: 0;
  color: #3b82f6;
}

/* Add Card */
.add-card {
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.75rem;
  padding: 1rem;
  margin-bottom: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.add-card h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.75rem;
}

.add-hint {
  font-size: 0.75rem;
  color: #64748b;
  margin-bottom: 0.25rem;
}

.add-hint-example {
  font-size: 0.6875rem;
  color: #475569;
  font-style: italic;
  margin-bottom: 0.75rem;
}

.form-group {
  margin-bottom: 0.75rem;
}

.form-group label {
  display: block;
  font-size: 0.75rem;
  color: #94a3b8;
  margin-bottom: 0.375rem;
}

.form-input {
  width: 100%;
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.8125rem;
  padding: 0.5rem 0.75rem;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-input::placeholder {
  color: #64748b;
}

.bulk-textarea {
  width: 100%;
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.8125rem;
  padding: 0.5rem 0.75rem;
  margin-bottom: 0.75rem;
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
}

.bulk-textarea:focus {
  outline: none;
  border-color: #3b82f6;
}

.bulk-textarea::placeholder {
  color: #64748b;
}

.add-btn {
  width: 100%;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8) !important;
  border: none !important;
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 3rem 1rem;
}

.empty-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 1rem;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-icon i {
  font-size: 2rem;
  color: #475569;
}

.empty-state h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.5rem;
}

.empty-state p {
  font-size: 0.875rem;
  color: #64748b;
}

@media (min-width: 768px) {
  .contacts-container {
    padding: 1.5rem 2rem;
  }
}
</style>
