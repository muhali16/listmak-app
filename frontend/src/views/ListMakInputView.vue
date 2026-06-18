<template>
  <div class="listmak-container">
    <!-- Selection Header -->
    <header class="listmak-selection">
       <div class="selection-wrapper">
          <select 
            v-model="selectedListMak" 
            class="custom-select listmak-select"
            :disabled="availableListMaks.length === 0"
          >
            <option :value="null" disabled>Pilih ListMak Hari Ini</option>
            <option v-for="lm in availableListMaks" :key="lm.id" :value="lm">
               {{ lm.title }}
            </option>
          </select>
          <Button 
            icon="pi pi-plus" 
            aria-label="Buat Baru" 
            @click="showCreateListMakModal = true"
            severity="secondary"
            class="create-btn"
          />
       </div>
       <p v-if="!selectedListMak" class="selection-hint">
          Pilih ListMak atau buat baru untuk mulai input pesanan.
       </p>
    </header>

    <!-- Content Wrapper (Only if Listmak selected) -->
    <div v-if="selectedListMak" class="listmak-content">
      <!-- Summary Cards -->
      <section v-if="orders.length > 0" class="summary-section">
        <div class="summary-grid">
          <div class="summary-card">
            <span class="summary-label">Total</span>
            <span class="summary-value">{{ orders.length }}</span>
          </div>
          <div class="summary-card summary-green">
            <span class="summary-label">Bayar</span>
            <span class="summary-value">{{ paidCount }}</span>
          </div>
          <div class="summary-card summary-blue">
            <span class="summary-label">Total Biaya</span>
            <span class="summary-value">{{ formatShortCurrency(totalAmount) }}</span>
          </div>
          <div class="summary-card summary-purple">
            <span class="summary-label">Terbayar</span>
            <span class="summary-value">{{ formatShortCurrency(paidAmount) }}</span>
          </div>
        </div>
      </section>

      <!-- Add Order & Share Buttons -->
      <section class="add-order-section">
        <Button 
          @click="showAddModal = true" 
          label="Tambah Pesanan"
          icon="pi pi-plus"
          class="add-order-btn"
        />
        <Button 
          @click="showShareModal = true" 
          label="Share Link"
          icon="pi pi-share-alt"
          severity="secondary"
          class="share-btn"
        />
      </section>

    <!-- Active Share Links -->
    <section v-if="selectedListMak && selectedListMak.shareLinks && selectedListMak.shareLinks.length > 0" class="active-shares-section">
      <h3 class="section-title-small">Link Aktif</h3>
      <div class="share-links-list">
        <div v-for="link in selectedListMak.shareLinks" :key="link.id" class="share-link-item">
           <div class="share-info">
             <span class="share-title">{{ link.title }}</span>
             <span class="share-expiry" :class="{ 'expired': isExpired(link.expires_at) }">
                <i class="pi pi-clock"></i> {{ formatExpiry(link.expires_at) }}
             </span>
           </div>
           <button @click="copyExistingShareLink(link.share_id)" class="share-copy-btn" title="Copy Link">
             <i class="pi pi-copy"></i>
           </button>
        </div>
      </div>
    </section>
    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Konfirmasi Hapus</h3>
          <button @click="showDeleteModal = false" class="modal-close">
            <i class="pi pi-times"></i>
          </button>
        </div>
        
        <div class="modal-body">
            <p>Apakah Anda yakin ingin menghapus pesanan ini?</p>
            <p v-if="orderToDelete && orderToDelete.data" class="modal-desc" style="margin-top: 0.5rem">
               "{{ orderToDelete.data.name }}: {{ orderToDelete.data.order }}"
            </p>
        </div>
        
        <div class="modal-footer">
             <Button label="Batal" severity="secondary" @click="showDeleteModal = false" class="cancel-btn" />
             <Button label="Hapus" severity="danger" @click="confirmDeleteOrder" :loading="deletingOrder" class="delete-confirm-btn" />
        </div>
      </div>
    </div>

    <!-- Checkbox Info -->
    <section v-if="orders.length > 0" class="info-section">
      <div class="checkbox-info">
        <i class="pi pi-info-circle"></i>
        <span>Centang (✓) menandakan pesanan <strong>sudah dibayar</strong>, bukan sudah dibeli.</span>
      </div>
    </section>

    <!-- Orders Table -->
    <section v-if="orders.length > 0" class="table-section">
      <!-- Search Input -->
      <div class="search-wrapper">
        <i class="pi pi-search search-icon"></i>
        <input 
          type="text"
          v-model="searchQuery"
          placeholder="Cari nama atau pesanan..."
          class="search-input"
        />
        <button v-if="searchQuery" @click="searchQuery = ''" class="search-clear">
          <i class="pi pi-times"></i>
        </button>
      </div>

      <!-- Mobile Card Layout -->
      <div class="mobile-cards">
        <div 
          v-for="(order, index) in filteredOrders" 
          :key="'mobile-' + order.originalIndex" 
          class="order-card"
          :class="{ 'card-paid': orders[order.originalIndex].paid }"
        >
          <div class="card-header">
            <input 
              type="checkbox" 
              v-model="orders[order.originalIndex].paid"
              @change="updateSingleOrder(orders[order.originalIndex])"
              class="custom-checkbox"
            />
            <input 
              type="text"
              v-model="orders[order.originalIndex].name" 
              class="card-name-input"
              @change="updateSingleOrder(orders[order.originalIndex])"
              placeholder="Nama"
            />
            <button @click="removeOrder(order.originalIndex)" class="card-delete-btn">
              <i class="pi pi-trash"></i>
            </button>
          </div>
          
          <div class="card-body">
            <textarea 
              v-model="orders[order.originalIndex].order" 
              class="card-order-input"
              @input="autoResizeTextarea($event); saveOrders()"
              @change="updateSingleOrder(orders[order.originalIndex])"
              @focus="autoResizeTextarea($event)"
              rows="1"
              placeholder="Pesanan..."
            ></textarea>
          </div>
          
          <div class="card-footer">
            <div class="card-price-group">
              <label>Harga</label>
              <input 
                type="number"
                v-model.number="orders[order.originalIndex].price" 
                class="card-number-input"
                @change="updateSingleOrder(orders[order.originalIndex])"
                placeholder="0"
              />
            </div>
            <div class="card-qty-group">
              <label>Qty</label>
              <input 
                type="number"
                v-model.number="orders[order.originalIndex].qty" 
                class="card-number-input card-qty-input"
                @change="updateSingleOrder(orders[order.originalIndex])"
                min="1"
              />
            </div>
            <div class="card-total-group">
              <label>Total</label>
              <span class="card-total-value">
                {{ orders[order.originalIndex].price ? formatShortCurrency(orders[order.originalIndex].price * orders[order.originalIndex].qty) : '-' }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Desktop Table Layout -->
      <div class="table-wrapper desktop-table">
        <table class="orders-table">
          <thead>
            <tr>
              <th class="th-check">
                <input 
                  type="checkbox" 
                  v-model="selectAll" 
                  @change="toggleSelectAll"
                  class="custom-checkbox"
                />
              </th>
              <th>Nama</th>
              <th>Pesanan</th>
              <th class="th-price">Harga</th>
              <th class="th-qty">Qty</th>
              <th class="th-total">Total</th>
              <th class="th-action"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(order, index) in filteredOrders" :key="order.originalIndex" :class="{ 'row-paid': order.paid }">
              <td class="td-check">
                <input 
                  type="checkbox" 
                  v-model="orders[order.originalIndex].paid"
                  @change="updateSingleOrder(orders[order.originalIndex])"
                  class="custom-checkbox"
                />
              </td>
              <td class="td-name">
                <input 
                  type="text"
                  v-model="orders[order.originalIndex].name" 
                  class="cell-input"
                  @change="updateSingleOrder(orders[order.originalIndex])"
                />
              </td>
              <td class="td-order">
                <textarea 
                  :ref="'orderTextarea_' + order.originalIndex"
                  v-model="orders[order.originalIndex].order" 
                  class="cell-input cell-order cell-textarea"
                  @input="autoResizeTextarea($event); saveOrders()"
                  @change="updateSingleOrder(orders[order.originalIndex])"
                  @focus="autoResizeTextarea($event)"
                  rows="1"
                ></textarea>
              </td>
              <td class="td-price">
                <input 
                  type="number"
                  v-model.number="orders[order.originalIndex].price" 
                  class="cell-input cell-number"
                  @change="updateSingleOrder(orders[order.originalIndex])"
                  placeholder="0"
                />
              </td>
              <td class="td-qty">
                <input 
                  type="number"
                  v-model.number="orders[order.originalIndex].qty" 
                  class="cell-input cell-number cell-qty"
                  @change="updateSingleOrder(orders[order.originalIndex])"
                  min="1"
                />
              </td>
              <td class="td-total">
                {{ orders[order.originalIndex].price ? formatShortCurrency(orders[order.originalIndex].price * orders[order.originalIndex].qty) : '' }}
              </td>
              <td class="td-action">
                <button @click="removeOrder(order.originalIndex)" class="delete-btn">
                  <i class="pi pi-trash"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Action Buttons -->
      <div class="action-buttons">
        <Button 
          @click="copyToClipboard" 
          label="Copy"
          icon="pi pi-copy"
          severity="secondary"
          size="small"
          class="copy-btn"
        />
        <Button 
          @click="shareViewList" 
          label="Share"
          icon="pi pi-external-link"
          severity="secondary"
          size="small"
          class="share-view-btn"
        />
        <Button 
          @click="saveAndNotify" 
          label="Simpan"
          icon="pi pi-save"
          size="small"
          class="save-btn"
        />
      </div>
    </section>

    <!-- Empty State -->
    <section v-if="orders.length === 0" class="empty-state">
      <div class="empty-icon">
        <i class="pi pi-inbox"></i>
      </div>
      <h3>Belum Ada Data</h3>
      <p>Klik "Tambah Pesanan" untuk mulai</p>
    </section>
  </div>

    <!-- Create ListMak Modal (Manual) -->
    <div v-if="showCreateListMakModal" class="modal-overlay" @click.self="showCreateListMakModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Buat ListMak Baru</h3>
          <button @click="showCreateListMakModal = false" class="modal-close">
            <i class="pi pi-times"></i>
          </button>
        </div>
        
        <div class="modal-body">
            <div class="form-group">
                <label>Judul ListMak</label>
                <input 
                  type="text" 
                  v-model="newListMakTitle" 
                  placeholder="Contoh: Makan Siang Ops"
                  class="modal-input"
                  @keyup.enter="createNewListMak"
                />
            </div>
        </div>
        
        <div class="modal-footer">
             <Button label="Buat ListMak" @click="createNewListMak" :loading="creatingListMak" :disabled="!newListMakTitle" class="submit-btn" />
        </div>
      </div>
    </div>

    <!-- Add Order Modal -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>Tambah Pesanan</h3>
          <button @click="showAddModal = false" class="modal-close">
            <i class="pi pi-times"></i>
          </button>
        </div>
        
        <!-- Input Mode Tabs -->
        <div class="modal-tabs">
          <button 
            class="modal-tab" 
            :class="{ active: addInputMode === 'parse' }"
            @click="addInputMode = 'parse'"
          >
            <i class="pi pi-whatsapp"></i>
            <span>Parse WA</span>
          </button>
          <button 
            class="modal-tab" 
            :class="{ active: addInputMode === 'manual' }"
            @click="addInputMode = 'manual'"
          >
            <i class="pi pi-pencil"></i>
            <span>Manual</span>
          </button>
        </div>
        
        <div class="modal-body">
          <!-- Parse WA Tab -->
          <div v-if="addInputMode === 'parse'">
            <p class="modal-desc">Paste daftar pesanan dari WhatsApp</p>
            <div class="form-group">
              <textarea
                v-model="rawInput"
                placeholder="1. Icha: Nasi Ayam Madura PAHA
2. Ali: Nasi Ayam Madura DADA
3. Budi: Nasi Goreng Spesial"
                rows="8"
                class="modal-textarea"
              ></textarea>
            </div>
          </div>
          
          <!-- Manual Tab -->
          <div v-if="addInputMode === 'manual'">
            <p class="modal-desc">Input pesanan satu per satu</p>
            <div class="form-group">
              <label>Nama Pemesan</label>
              <input 
                type="text" 
                v-model="manualForm.name" 
                placeholder="Nama pemesan"
                class="modal-input"
              />
            </div>
            <div class="form-group">
              <label>Pesanan</label>
              <textarea 
                v-model="manualForm.order" 
                placeholder="Detail pesanan..."
                rows="3"
                class="modal-textarea"
              ></textarea>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label>Harga</label>
                <input 
                  type="number" 
                  v-model.number="manualForm.price" 
                  placeholder="0"
                  class="modal-input"
                />
              </div>
              <div class="form-group form-group-small">
                <label>Qty</label>
                <input 
                  type="number" 
                  v-model.number="manualForm.qty" 
                  min="1"
                  class="modal-input"
                />
              </div>
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <Button 
            v-if="addInputMode === 'parse'"
            @click="parseInputAndClose" 
            label="Parse & Tambah"
            icon="pi pi-bolt"
            class="add-manual-btn"
            :disabled="!rawInput.trim()"
          />
          <Button 
            v-if="addInputMode === 'manual'"
            @click="addManualOrderAndClose" 
            label="Tambah Pesanan"
            icon="pi pi-plus"
            class="add-manual-btn"
            :disabled="!manualForm.name.trim() || !manualForm.order.trim()"
          />
        </div>
      </div>
    </div>

    <!-- Share Link Modal -->
    <div v-if="showShareModal" class="modal-overlay" @click.self="showShareModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Share Link ListMak</h3>
          <button @click="showShareModal = false" class="modal-close">
            <i class="pi pi-times"></i>
          </button>
        </div>
        
        <div class="modal-body">
          <p class="modal-desc">Bagikan link untuk pemesan input pesanan sendiri</p>
          
          <div class="form-group">
            <label>Judul ListMak</label>
            <input 
              type="text" 
              v-model="shareForm.title" 
              placeholder="Contoh: Makan Siang Kantor"
              class="modal-input"
            />
          </div>
          
          <div class="form-group">
            <label>Batas Waktu Input</label>
            <input 
              type="datetime-local" 
              v-model="shareForm.expiresAt"
              :min="minDateTime"
              class="modal-input"
            />
          </div>
          
          <div v-if="generatedShareLink" class="share-link-result">
            <label>Link Share:</label>
            <div class="share-link-box">
              <input type="text" :value="generatedShareLink" readonly class="share-link-input" />
              <button @click="copyShareLink" class="copy-link-btn">
                <i class="pi pi-copy"></i>
              </button>
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <Button 
            v-if="!generatedShareLink"
            @click="generateShareLink" 
            label="Generate Link"
            icon="pi pi-link"
            class="generate-btn"
            :disabled="!shareForm.expiresAt"
          />
          <Button 
            v-else
            @click="shareViaWhatsApp" 
            label="Share via WhatsApp"
            icon="pi pi-whatsapp"
            class="wa-share-btn"
          />
        </div>
      </div>
    </div>

    <!-- Manual Input Modal (OB) -->
    <div v-if="showManualInputModal" class="modal-overlay" @click.self="showManualInputModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Tambah Pesanan Manual</h3>
          <button @click="showManualInputModal = false" class="modal-close">
            <i class="pi pi-times"></i>
          </button>
        </div>
        
        <div class="modal-body">
          <p class="modal-desc">Tambahkan pesanan di luar waktu share link</p>
          
          <div class="form-group">
            <label>Nama Pemesan</label>
            <input 
              type="text" 
              v-model="manualForm.name" 
              placeholder="Nama pemesan"
              class="modal-input"
            />
          </div>
          
          <div class="form-group">
            <label>Pesanan</label>
            <textarea 
              v-model="manualForm.order" 
              placeholder="Detail pesanan..."
              rows="3"
              class="modal-textarea"
            ></textarea>
          </div>
          
          <div class="form-row">
            <div class="form-group">
              <label>Harga</label>
              <input 
                type="number" 
                v-model.number="manualForm.price" 
                placeholder="0"
                class="modal-input"
              />
            </div>
            <div class="form-group form-group-small">
              <label>Qty</label>
              <input 
                type="number" 
                v-model.number="manualForm.qty" 
                min="1"
                class="modal-input"
              />
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <Button 
            @click="addManualOrder" 
            label="Tambah Pesanan"
            icon="pi pi-plus"
            class="add-manual-btn"
            :disabled="!manualForm.name.trim() || !manualForm.order.trim()"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Button from 'primevue/button'
import { listmak, share } from '../api'

export default {
  name: 'ListMakInputView',
  components: {
    Button
  },
  data() {
    return {
      // ListMak Selection
      availableListMaks: [],
      selectedListMak: null,
      showCreateListMakModal: false,
      newListMakTitle: '',
      creatingListMak: false,
      contextDate: new Date().toISOString().split('T')[0], // Default today

      rawInput: '',
      orders: [],
      selectAll: false,
      searchQuery: '',
      listmakId: null, // Kept for compatibility, sync with selectedListMak.id
      syncing: false,
      // Modal states
      showAddModal: false,
      showShareModal: false,
      showDeleteModal: false, // Konfirmasi Hapus
      orderToDelete: null, // Index or Object
      deletingOrder: false,
      addInputMode: 'parse',
      generatedShareLink: '',
      // Share form
      shareForm: {
        title: '',
        expiresAt: ''
      },
      // Manual input form
      manualForm: {
        name: '',
        order: '',
        price: 0,
        qty: 1
      }
    }
  },
  computed: {
    filteredOrders() {
      if (!this.searchQuery.trim()) {
        return this.orders.map((order, index) => ({ ...order, originalIndex: index }))
      }
      const query = this.searchQuery.toLowerCase()
      return this.orders
        .map((order, index) => ({ ...order, originalIndex: index }))
        .filter(o => 
          o.name.toLowerCase().includes(query) || 
          o.order.toLowerCase().includes(query)
        )
    },
    paidCount() {
      return this.orders.filter(o => o.paid).length
    },
    totalAmount() {
      return this.orders.reduce((sum, o) => sum + (o.price * o.qty), 0)
    },
    paidAmount() {
      return this.orders.filter(o => o.paid).reduce((sum, o) => sum + (o.price * o.qty), 0)
    },
    minDateTime() {
      const now = new Date()
      now.setMinutes(now.getMinutes() + 5) // Minimum 5 minutes from now
      return now.toISOString().slice(0, 16)
    }
  },
  watch: {
    selectedListMak(newVal) {
      if (newVal) {
        this.listmakId = newVal.id
        this.loadOrdersFromSelection(newVal)
      } else {
        this.listmakId = null
        this.orders = []
      }
    }
  },
  mounted() {
    this.loadTodayListMaks()
  },
  updated() {
    this.$nextTick(() => {
      this.resizeAllTextareas()
    })
  },
  methods: {
    async loadTodayListMaks() {
      const queryDate = this.$route.query.date
      const targetDate = queryDate || new Date().toISOString().split('T')[0]
      this.contextDate = targetDate

      try {
        const response = await listmak.getListMakByDate(targetDate)
        if (response.success && response.data) {
          // Response data is array of listmaks
          this.availableListMaks = response.data.map(lm => ({
             id: lm.id,
             title: lm.title,
             date: lm.date,
             orderCount: lm.orders ? lm.orders.length : 0,
             orders: lm.orders || [],
             shareLinks: lm.share_links || [] // Map share links from API
          }))
          
          // Auto select first if available
          if (this.availableListMaks.length > 0) {
            this.selectedListMak = this.availableListMaks[0]
          }
        }
      } catch (error) {
        console.error('Failed to load listmaks:', error)
      }
    },
    async createNewListMak() {
       if (!this.newListMakTitle.trim()) return

       this.creatingListMak = true
       try {
          const dateObj = new Date(this.contextDate)
          const safeDate = !isNaN(dateObj.getTime()) ? dateObj.toISOString() : new Date().toISOString()
          
          const response = await listmak.createListMak({
             title: this.newListMakTitle,
             date: safeDate
          })

          if (response.success && response.data) {
             const newItem = {
                id: response.data.id,
                title: response.data.title,
                date: response.data.date,
                orders: [],
                shareLinks: []
             }
             this.availableListMaks.push(newItem)
             this.selectedListMak = newItem
             this.showCreateListMakModal = false
             this.newListMakTitle = ''
             
             this.$toast.add({
                severity: 'success',
                summary: 'Berhasil',
                detail: 'ListMak baru dibuat',
                life: 2000
             })
          }
       } catch(error) {
          this.$toast.add({
             severity: 'error',
             summary: 'Gagal',
             detail: 'Gagal membuat ListMak',
             life: 3000
          })
       } finally {
          this.creatingListMak = false
       }
    },
    loadOrdersFromSelection(listmakData) {
       this.orders = (listmakData.orders || []).map(o => ({
            id: o.id,
            name: o.name,
            order: o.order_detail || o.order,
            price: o.price,
            qty: o.qty,
            paid: o.is_paid,
            addedVia: o.added_via,
            addedAt: o.added_at
       }))
    },
    parseInput() {
      if (!this.rawInput.trim()) {
        this.$toast.add({
          severity: 'warn',
          summary: 'Perhatian',
          detail: 'Paste daftar pesanan terlebih dahulu',
          life: 3000
        })
        return false
      }

      const lines = this.rawInput.split('\n').filter(line => line.trim())
      const parsedOrders = []

      for (const line of lines) {
        // Match patterns like "1. Name: Order" or "Name: Order"
        const match = line.match(/^(?:\d+[\.\)\-]\s*)?(?:⁠)?(.+?)[\:\-;]\s*(.+)$/i)
        
        if (match) {
          const name = match[1].trim()
          const order = match[2].trim()
          
          parsedOrders.push({
            name,
            order,
            price: 0,
            qty: 1,
            paid: false
          })
        }
      }

      if (parsedOrders.length > 0) {
        // Append to existing orders instead of replacing
        this.orders = [...this.orders, ...parsedOrders]
        this.saveOrders()
        this.$toast.add({
          severity: 'success',
          summary: 'Berhasil',
          detail: `${parsedOrders.length} pesanan ditambahkan`,
          life: 2000
        })
        this.rawInput = ''
        return true
      } else {
        this.$toast.add({
          severity: 'error',
          summary: 'Gagal',
          detail: 'Format tidak dikenali',
          life: 3000
        })
        return false
      }
    },
    parseInputAndClose() {
      const success = this.parseInput()
      if (success) {
        this.showAddModal = false
      }
    },
    addManualOrderAndClose() {
      this.addManualOrder()
      this.showAddModal = false
    },
    toggleSelectAll() {
      this.orders.forEach(o => {
        o.paid = this.selectAll
      })
      this.saveOrders()
    },
    removeOrder(index) {
      this.orderToDelete = { index, data: this.orders[index] }
      this.showDeleteModal = true
    },
    async confirmDeleteOrder() {
       if (!this.orderToDelete) return
       
       const order = this.orderToDelete.data
       const index = this.orderToDelete.index
       
       this.deletingOrder = true
       try {
           if (order.id) {
               // Delete from API
               await listmak.deleteOrder(order.id)
           }
           
           // Remove locally
           this.orders.splice(index, 1)
           this.saveOrders()
           
           this.$toast.add({
                severity: 'success',
                summary: 'Terhapus',
                detail: 'Pesanan berhasil dihapus',
                life: 2000
           })
           this.showDeleteModal = false
       } catch (error) {
           console.error('Failed to delete order:', error)
           this.$toast.add({
                severity: 'error',
                summary: 'Gagal',
                detail: 'Gagal menghapus pesanan',
                life: 3000
           })
       } finally {
           this.deletingOrder = false
       }
    },
    clearAll() {
      this.rawInput = ''
      this.orders = []
      this.selectAll = false
      const today = new Date().toISOString().split('T')[0]
      localStorage.removeItem(`listmak_${today}`)
    },
    saveOrders() {
      const today = new Date().toISOString().split('T')[0]
      localStorage.setItem(`listmak_${today}`, JSON.stringify(this.orders))
    },
    async loadOrders() {
      const today = new Date().toISOString().split('T')[0]
      
      // Try loading from API first
      try {
        const response = await listmak.getListMakByDate(today)
        if (response.success && response.data) {
          this.listmakId = response.data.id
          this.orders = response.data.orders.map(o => ({
            id: o.id, // Keep backend ID
            name: o.name,
            order: o.order_detail,
            price: o.price,
            qty: o.qty,
            paid: o.is_paid,
            addedVia: o.added_via,
            addedAt: o.added_at
          }))
          return
        }
      } catch (error) {
        console.log('Daily listmak not found on server or offline, loading local:', error)
      }

      // Fallback to local storage
      const saved = localStorage.getItem(`listmak_${today}`)
      if (saved) {
        this.orders = JSON.parse(saved)
      }
    },
    async saveAndNotify() {
      // Validation: Must select listmak first
      if (!this.selectedListMak) {
         this.$toast.add({
            severity: 'warn',
            summary: 'Perhatian',
            detail: 'Pilih ListMak terlebih dahulu atau buat baru',
            life: 3000
         })
         return
      }

      this.saveOrders() // Local backup
      this.syncing = true
      const targetListMakId = this.selectedListMak.id

      try {
        // Sync orders
        const promises = this.orders.map(async (order) => {
           if (order.id) {
              // Update existing order
              await listmak.updateOrder(order.id, {
                 listmak_id: targetListMakId, // Explicitly sending listmak_id as requested
                 name: order.name,
                 order_detail: order.order,
                 price: order.price,
                 qty: order.qty
              })
              
              // Sync paid status check
              if (order.paid !== undefined) {
                  await listmak.updateOrderPaidStatus(order.id, order.paid)
              }
           } else {
              // Create new order
              const res = await listmak.addOrder(targetListMakId, {
                 listmak_id: targetListMakId, // Explicitly sending listmak_id as requested
                 name: order.name,
                 order_detail: order.order,
                 price: order.price,
                 qty: order.qty,
                 added_via: order.addedVia || 'manual'
              })
              if (res.success && res.data) {
                 order.id = res.data.id
                 // If paid status was set locally before sync
                 if (order.paid) {
                    await listmak.updateOrderPaidStatus(order.id, true)
                 }
              }
           }
        })
        
        await Promise.all(promises)
        
        // Update local storage with new IDs
        this.saveOrders() 
        
        // Refresh orders from API to ensure sync state
        // this.loadOrdersFromSelection(...) - Optional but good practice

        this.$toast.add({
          severity: 'success',
          summary: 'Tersimpan',
          detail: 'Data berhasil disinkronkan ke server',
          life: 2000
        })
      } catch (error) {
        console.error('Sync error:', error)
        this.$toast.add({
          severity: 'warn',
          summary: 'Tersimpan Lokal',
          detail: 'Gagal sinkron ke server: ' + error.message,
          life: 3000
        })
      } finally {
        this.syncing = false
      }
    },
    async updateSingleOrder(order) {
       this.saveOrders() // Local backup first
       
       if (!order.id || !this.selectedListMak) return 

       try {
          const listmakId = this.selectedListMak.id
          
          // Update details
          await listmak.updateOrder(order.id, {
             listmak_id: listmakId,
             name: order.name,
             order_detail: order.order,
             price: order.price,
             qty: order.qty
          })
          
          // Update paid status if changed (always sending it is safer for sync)
          if (order.paid !== undefined) {
              await listmak.updateOrderPaidStatus(order.id, order.paid)
          }
       } catch (error) {
          console.error('Failed to auto-update order:', error)
          // Silent fail or small toast? Let's toast for now but minimal
          this.$toast.add({
             severity: 'warn',
             summary: 'Sync Pending',
             detail: 'Perubahan belum tersimpan di server',
             life: 1000
          })
       }
    },
    copyToClipboard() {
      if (this.orders.length === 0) return

      let text = 'ListMak:\n'
      
      this.orders.forEach((order, index) => {
        const total = this.formatCurrency(order.price * order.qty)
        const status = order.paid ? '✅' : '⏳'
        text += `${index + 1}. ${order.name}: ${order.order} - ${total} ${status}\n`
      })

      text += `\n📊 Total: ${this.formatCurrency(this.totalAmount)}`
      text += `\n✅ Sudah Bayar: ${this.paidCount}/${this.orders.length} (${this.formatCurrency(this.paidAmount)})`
      text += `\n⏳ Belum Bayar: ${this.orders.length - this.paidCount} (${this.formatCurrency(this.totalAmount - this.paidAmount)})`

      // Fallback untuk mobile browser
      this.copyTextFallback(text)
    },
    copyTextFallback(text) {
      // Coba menggunakan Clipboard API modern
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text).then(() => {
          this.showCopySuccess()
        }).catch(() => {
          this.fallbackCopyText(text)
        })
      } else {
        this.fallbackCopyText(text)
      }
    },
    fallbackCopyText(text) {
      // Fallback menggunakan textarea dan execCommand
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      textArea.style.opacity = '0'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      
      try {
        const successful = document.execCommand('copy')
        if (successful) {
          this.showCopySuccess()
        } else {
          this.showCopyError()
        }
      } catch (err) {
        this.showCopyError()
      }
      
      document.body.removeChild(textArea)
    },
    showCopySuccess() {
      this.$toast.add({
        severity: 'success',
        summary: 'Disalin',
        detail: 'List berhasil dicopy ke clipboard',
        life: 2000
      })
    },
    showCopyError() {
      this.$toast.add({
        severity: 'error',
        summary: 'Gagal',
        detail: 'Tidak bisa menyalin ke clipboard',
        life: 3000
      })
    },
    formatCurrency(value) {
      return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0
      }).format(value)
    },
    formatShortCurrency(value) {
      if (value >= 1000000) {
        return `${(value / 1000000).toFixed(1)}jt`
      }
      if (value >= 1000) {
        return `${Math.round(value / 1000)}rb`
      }
      return value.toString()
    },
    autoResizeTextarea(event) {
      const textarea = event.target
      textarea.style.height = 'auto'
      textarea.style.height = textarea.scrollHeight + 'px'
    },
    resizeAllTextareas() {
      // Target both desktop and mobile textareas
      const textareas = this.$el.querySelectorAll('.cell-textarea, .card-order-input')
      textareas.forEach(textarea => {
        textarea.style.height = 'auto'
        textarea.style.height = textarea.scrollHeight + 'px'
      })
    },
    // Share Link Helpers
    copyExistingShareLink(shareId) {
       const baseUrl = window.location.origin
       const link = `${baseUrl}/listmak/order/${shareId}`
       this.copyTextFallback(link)
       this.$toast.add({
          severity: 'success',
          summary: 'Disalin',
          detail: 'Link share disalin ke clipboard',
          life: 2000
       })
    },
    formatExpiry(isoDate) {
       if (!isoDate) return '-'
       const date = new Date(isoDate)
       // Check if date is valid
       if (isNaN(date.getTime())) return '-'
       
       const options = { hour: '2-digit', minute: '2-digit' }
       const timeStr = date.toLocaleTimeString('id-ID', options)
       
       const today = new Date()
       const isToday = date.getDate() === today.getDate() && 
                       date.getMonth() === today.getMonth() && 
                       date.getFullYear() === today.getFullYear()
                       
       return isToday ? timeStr : `${date.getDate()}/${date.getMonth()+1} ${timeStr}`
    },
    isExpired(isoDate) {
       if (!isoDate) return false
       return new Date(isoDate) < new Date()
    },
    // Share Link Methods
    async generateShareLink() {
      if (!this.shareForm.expiresAt) {
        this.$toast.add({
          severity: 'warn',
          summary: 'Perhatian',
          detail: 'Pilih batas waktu input terlebih dahulu',
          life: 3000
        })
        return
      }

      // Sync first
      if (!this.listmakId) {
          await this.saveAndNotify()
      }
      
      if (!this.listmakId) {
          this.$toast.add({
            severity: 'error',
            summary: 'Gagal',
            detail: 'Gagal sinkronisasi data ListMak',
            life: 3000
          })
          return 
      }

      try {
          const response = await share.createShareLink({
              listmak_id: this.listmakId,
              title: this.shareForm.title || `Input ListMak ${new Date().toLocaleDateString('id-ID')}`,
              expires_at: new Date(this.shareForm.expiresAt).toISOString()
          })
          
          if (response.success && response.data) {
             const baseUrl = window.location.origin
             this.generatedShareLink = `${baseUrl}/listmak/order/${response.data.share_id}`
             
             this.$toast.add({
                severity: 'success',
                summary: 'Berhasil',
                detail: 'Link share berhasil dibuat',
                life: 2000
             })
          }
      } catch (error) {
          console.error('Failed to create share link:', error)
          this.$toast.add({
            severity: 'error',
            summary: 'Gagal',
            detail: error.message || 'Gagal membuat link share',
            life: 3000
          })
      }
    },
    generateShareId() {
      // Helper still used for random ID if needed, but backend handles share IDs.
      // Keeping it if needed for other things or remove if unused.
      const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
      let result = ''
      for (let i = 0; i < 8; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length))
      }
      return result
    },
    copyShareLink() {
      this.copyTextFallback(this.generatedShareLink)
    },
    shareViaWhatsApp() {
      const title = this.shareForm.title || 'ListMak Hari Ini'
      const expires = new Date(this.shareForm.expiresAt).toLocaleString('id-ID', {
        weekday: 'long',
        day: 'numeric',
        month: 'long',
        hour: '2-digit',
        minute: '2-digit'
      })
      
      const message = `📋 *${title}*\n\nSilakan input pesanan Anda melalui link berikut:\n${this.generatedShareLink}\n\n⏰ Batas waktu: ${expires}\n\n_Pesanan yang masuk setelah batas waktu tidak akan diproses._`
      
      const waUrl = `https://wa.me/?text=${encodeURIComponent(message)}`
      window.open(waUrl, '_blank')
    },
    // Manual Input Methods
    addManualOrder() {
      if (!this.manualForm.name.trim() || !this.manualForm.order.trim()) {
        this.$toast.add({
          severity: 'warn',
          summary: 'Perhatian',
          detail: 'Nama dan pesanan harus diisi',
          life: 3000
        })
        return
      }

      this.orders.push({
        name: this.manualForm.name.trim(),
        order: this.manualForm.order.trim(),
        price: this.manualForm.price || 0,
        qty: this.manualForm.qty || 1,
        paid: false,
        addedVia: 'manual',
        addedAt: new Date().toISOString()
      })

      this.saveOrders()
      this.showManualInputModal = false
      
      // Reset form
      this.manualForm = {
        name: '',
        order: '',
        price: 0,
        qty: 1
      }

      this.$toast.add({
        severity: 'success',
        summary: 'Berhasil',
        detail: 'Pesanan berhasil ditambahkan',
        life: 2000
      })
    },
    async shareViewList() {
      if (this.orders.length === 0) {
        this.$toast.add({
          severity: 'warn',
          summary: 'Perhatian',
          detail: 'Tidak ada pesanan untuk di-share',
          life: 3000
        })
        return
      }

      // Sync first
      if (!this.listmakId) {
          await this.saveAndNotify()
      }
      
      if (!this.listmakId) return

      try {
          const response = await share.createViewShare({
              listmak_id: this.listmakId,
              title: `ListMak ${new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long' })}`
          })
          
          if (response.success && response.data) {
             const baseUrl = window.location.origin
             const viewLink = `${baseUrl}/listmak/view/${response.data.view_id}`
             
             this.copyTextFallback(viewLink)
             
             this.$toast.add({
                severity: 'success',
                summary: 'Link Disalin',
                detail: 'Link view listmak berhasil disalin ke clipboard',
                life: 3000
             })
          }
      } catch (error) {
          console.error('Failed to create view share:', error)
          this.$toast.add({
            severity: 'error',
            summary: 'Gagal',
            detail: 'Gagal membuat link view',
            life: 3000
          })
      }
    }
  }
}
</script>

<style scoped>
.listmak-container {
  padding: 1rem;
  padding-bottom: 2rem;
}

/* Selection Header */
.listmak-selection {
  margin-bottom: 1.5rem;
}

.selection-wrapper {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.custom-select {
  flex: 1;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #f1f5f9;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  font-size: 0.9375rem;
  font-family: inherit;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2394a3b8'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.75rem center;
  background-size: 1rem;
  padding-right: 2.5rem;
}

.custom-select:focus {
  outline: none;
  border-color: #3b82f6;
  background-color: rgba(30, 41, 59, 0.8);
}

.custom-select option {
  background: #1e293b;
  color: #f1f5f9;
  padding: 0.5rem;
}

.create-btn {
  flex-shrink: 0;
  width: 2.5rem !important;
  height: 2.5rem !important;
  padding: 0 !important;
}

.selection-hint {
  font-size: 0.75rem;
  color: #64748b;
  text-align: center;
}

/* Add Order Section */
.add-order-section {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.add-order-btn {
  flex: 1;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8) !important;
  border: none !important;
  font-size: 0.9375rem !important;
  padding: 0.875rem !important;
}

.share-btn {
  background: rgba(30, 41, 59, 0.8) !important;
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  color: #94a3b8 !important;
  padding: 0.875rem !important;
}

/* Active Shares Section */
.active-shares-section {
  background: rgba(30, 41, 59, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.5rem;
  padding: 0.75rem;
  margin-bottom: 1rem;
}

.section-title-small {
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 0.5rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.share-links-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.share-link-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(15, 23, 42, 0.4);
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  border: 1px solid rgba(255, 255, 255, 0.02);
}

.share-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.share-title {
  font-size: 0.8125rem;
  color: #f1f5f9;
  font-weight: 500;
}

.share-expiry {
  font-size: 0.6875rem;
  color: #22c55e;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.share-expiry.expired {
  color: #ef4444;
}

.share-copy-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 0.25rem;
  transition: color 0.2s;
}

.share-copy-btn:hover {
  color: #3b82f6;
}

/* Modal Tabs */
.modal-tabs {
  display: flex;
  background: rgba(15, 23, 42, 0.5);
  padding: 0.25rem;
  gap: 0.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.modal-tab {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  padding: 0.625rem 0.5rem;
  background: transparent;
  border: none;
  color: #64748b;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.6875rem;
}

.modal-tab i {
  font-size: 1.125rem;
}

.modal-tab.active {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.modal-tab:hover:not(.active) {
  background: rgba(255, 255, 255, 0.05);
}

.modal-large {
  max-width: 450px;
}

/* Input Section */
.input-section {
  margin-bottom: 1rem;
}

.input-textarea {
  width: 100%;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.8125rem;
  padding: 0.625rem 0.75rem;
  margin-bottom: 0.5rem;
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
}

.input-textarea::placeholder {
  color: #64748b;
}

.input-textarea:focus {
  outline: none;
  border-color: #3b82f6;
}

.input-actions {
  display: flex;
  gap: 0.5rem;
}

.parse-btn {
  flex: 1;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8) !important;
  border: none !important;
  font-size: 0.8125rem !important;
}

.clear-btn {
  background: rgba(30, 41, 59, 0.8) !important;
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  color: #94a3b8 !important;
  font-size: 0.8125rem !important;
}

/* Summary Section */
.summary-section {
  margin-bottom: 1rem;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.375rem;
}

.summary-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.5rem;
  padding: 0.5rem;
  text-align: center;
}

.summary-label {
  display: block;
  font-size: 0.5625rem;
  color: #64748b;
  margin-bottom: 0.125rem;
}

.summary-value {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #f1f5f9;
}

.summary-green .summary-value { color: #22c55e; }
.summary-blue .summary-value { color: #3b82f6; }
.summary-purple .summary-value { color: #a855f7; }

/* Table Section */
.table-section {
  margin-bottom: 1rem;
}

.table-wrapper {
  overflow-x: auto;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.05);
  /* Hide scrollbar */
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE and Edge */
}

.table-wrapper::-webkit-scrollbar {
  display: none; /* Chrome, Safari, Opera */
}

/* Search Wrapper */
.search-wrapper {
  position: relative;
  margin-bottom: 0.75rem;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: #64748b;
  font-size: 0.875rem;
}

.search-input {
  width: 100%;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.8125rem;
  padding: 0.625rem 2rem 0.625rem 2.25rem;
  font-family: inherit;
  transition: border-color 0.2s ease;
}

.search-input::placeholder {
  color: #64748b;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.search-clear {
  position: absolute;
  right: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 0.25rem;
  font-size: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.search-clear:hover {
  color: #f1f5f9;
  background: rgba(255, 255, 255, 0.1);
}

.orders-table {
  width: 100%;
  min-width: 550px;
  border-collapse: collapse;
}

.orders-table th {
  background: rgba(15, 23, 42, 0.5);
  padding: 0.5rem 0.375rem;
  text-align: left;
  font-size: 0.625rem;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.orders-table td {
  padding: 0.25rem 0.375rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.02);
  vertical-align: middle;
}

.th-check, .td-check { width: 32px; text-align: center; }
.th-price, .td-price { width: 75px; }
.th-qty, .td-qty { width: 50px; }
.th-total, .td-total { width: 70px; text-align: right; }
.th-action, .td-action { width: 32px; text-align: center; }

.td-total {
  font-weight: 600;
  color: #f1f5f9;
  font-size: 0.75rem;
}

.row-paid {
  background: rgba(34, 197, 94, 0.05);
}

.row-paid .cell-input {
  text-decoration: line-through;
  color: #64748b !important;
}

/* Custom Checkbox */
.custom-checkbox {
  width: 16px;
  height: 16px;
  accent-color: #22c55e;
  cursor: pointer;
}

/* Cell Inputs */
.cell-input {
  width: 100%;
  background: transparent;
  border: 1px solid transparent;
  color: #f1f5f9;
  font-size: 0.75rem;
  padding: 0.25rem 0.375rem;
  border-radius: 0.25rem;
  font-family: inherit;
}

.cell-input:hover {
  border-color: rgba(255, 255, 255, 0.1);
}

.cell-input:focus {
  outline: none;
  border-color: #3b82f6;
  background: rgba(255, 255, 255, 0.03);
}

.cell-order {
  min-width: 150px;
  max-width: 250px;
}

/* Textarea auto-fit */
.cell-textarea {
  resize: none;
  overflow: hidden;
  min-height: 24px;
  line-height: 1.4;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.cell-number {
  text-align: right;
  width: 60px;
}

.cell-qty {
  width: 45px;
}

/* Remove number input spinners */
.cell-number::-webkit-outer-spin-button,
.cell-number::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.cell-number[type=number] {
  -moz-appearance: textfield;
  appearance: textfield;
}

.delete-btn {
  width: 24px;
  height: 24px;
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.25rem;
  font-size: 0.75rem;
}

.delete-btn:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

/* Action Buttons */
.action-buttons {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.copy-btn {
  flex: 1;
  background: rgba(30, 41, 59, 0.8) !important;
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  color: #94a3b8 !important;
  font-size: 0.8125rem !important;
}

.share-view-btn {
  flex: 1;
  background: rgba(59, 130, 246, 0.15) !important;
  border: 1px solid rgba(59, 130, 246, 0.3) !important;
  color: #3b82f6 !important;
  font-size: 0.8125rem !important;
}

.save-btn {
  flex: 1;
  background: linear-gradient(135deg, #22c55e, #16a34a) !important;
  border: none !important;
  font-size: 0.8125rem !important;
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 2rem 1rem;
}

.empty-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 0.75rem;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-icon i {
  font-size: 1.5rem;
  color: #475569;
}

.empty-state h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #f1f5f9;
  margin-bottom: 0.25rem;
}

.empty-state p {
  font-size: 0.8125rem;
  color: #64748b;
}

@media (min-width: 768px) {
  .listmak-container {
    padding: 1.5rem 2rem;
  }
  
  .mobile-cards {
    display: none;
  }
  
  .desktop-table {
    display: block;
  }
}

/* Mobile Cards - Default visible on mobile */
.mobile-cards {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.desktop-table {
  display: none;
}

.order-card {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 0.75rem;
  padding: 0.75rem;
  transition: all 0.2s ease;
}

.order-card.card-paid {
  background: rgba(34, 197, 94, 0.08);
  border-color: rgba(34, 197, 94, 0.2);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.card-name-input {
  flex: 1;
  background: transparent;
  border: 1px solid transparent;
  color: #f1f5f9;
  font-size: 0.9375rem;
  font-weight: 600;
  padding: 0.25rem 0.5rem;
  border-radius: 0.375rem;
  font-family: inherit;
}

.card-name-input:focus {
  outline: none;
  border-color: #3b82f6;
  background: rgba(255, 255, 255, 0.03);
}

.card-name-input::placeholder {
  color: #64748b;
}

.card-delete-btn {
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.card-delete-btn:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

.card-body {
  margin-bottom: 0.625rem;
}

.card-order-input {
  width: 100%;
  background: rgba(15, 23, 42, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #f1f5f9;
  font-size: 0.8125rem;
  padding: 0.5rem 0.625rem;
  border-radius: 0.5rem;
  font-family: inherit;
  resize: none;
  overflow: hidden;
  min-height: 32px;
  line-height: 1.4;
}

.card-order-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.card-order-input::placeholder {
  color: #64748b;
}

.card-footer {
  display: flex;
  gap: 0.5rem;
  align-items: flex-end;
}

.card-price-group,
.card-qty-group,
.card-total-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.card-price-group {
  flex: 1;
}

.card-qty-group {
  width: 60px;
}

.card-total-group {
  width: 70px;
  text-align: right;
}

.card-footer label {
  font-size: 0.625rem;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.card-number-input {
  width: 100%;
  background: rgba(15, 23, 42, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #f1f5f9;
  font-size: 0.8125rem;
  padding: 0.375rem 0.5rem;
  border-radius: 0.375rem;
  font-family: inherit;
  text-align: right;
}

.card-number-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.card-number-input::placeholder {
  color: #64748b;
}

/* Remove number input spinners for card inputs */
.card-number-input::-webkit-outer-spin-button,
.card-number-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.card-number-input[type=number] {
  -moz-appearance: textfield;
  appearance: textfield;
}

.card-total-value {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #22c55e;
  display: block;
  padding: 0.375rem 0;
}

.card-paid .card-name-input {
  text-decoration: line-through;
  color: #64748b;
}

.card-paid .card-order-input {
  text-decoration: line-through;
  color: #64748b;
}

/* Info Section */
.info-section {
  margin-bottom: 1rem;
}

.checkbox-info {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 0.5rem;
  padding: 0.75rem;
  font-size: 0.75rem;
  color: #94a3b8;
}

.checkbox-info i {
  color: #3b82f6;
  font-size: 1rem;
  flex-shrink: 0;
  margin-top: 0.0625rem;
}

.checkbox-info strong {
  color: #22c55e;
}

/* Quick Actions */
.quick-actions {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.share-link-btn {
  flex: 1;
  font-size: 0.8125rem !important;
}

.manual-input-btn {
  flex: 1;
  background: rgba(30, 41, 59, 0.8) !important;
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  color: #94a3b8 !important;
  font-size: 0.8125rem !important;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: #1e293b;
  border-radius: 1rem;
  width: 100%;
  max-width: 400px;
  max-height: 90vh;
  overflow-y: auto;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.modal-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #f1f5f9;
}

.modal-close {
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  font-size: 1rem;
}

.modal-close:hover {
  color: #f1f5f9;
  background: rgba(255, 255, 255, 0.1);
}

.modal-body {
  padding: 1.25rem;
}

.modal-desc {
  font-size: 0.8125rem;
  color: #64748b;
  margin-bottom: 1.25rem;
}

.modal-body .form-group {
  margin-bottom: 1rem;
}

.modal-body .form-group label {
  display: block;
  font-size: 0.75rem;
  font-weight: 500;
  color: #94a3b8;
  margin-bottom: 0.375rem;
}

.modal-input,
.modal-textarea {
  width: 100%;
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #f1f5f9;
  font-size: 0.875rem;
  padding: 0.625rem 0.75rem;
  font-family: inherit;
}

.modal-input:focus,
.modal-textarea:focus {
  outline: none;
  border-color: #3b82f6;
}

.modal-input::placeholder,
.modal-textarea::placeholder {
  color: #64748b;
}

.modal-textarea {
  resize: vertical;
  min-height: 80px;
}

.form-row {
  display: flex;
  gap: 0.75rem;
}

.form-group-small {
  width: 80px;
  flex-shrink: 0;
}

.share-link-result {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.share-link-result label {
  display: block;
  font-size: 0.75rem;
  color: #94a3b8;
  margin-bottom: 0.5rem;
}

.share-link-box {
  display: flex;
  gap: 0.5rem;
}

.share-link-input {
  flex: 1;
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
  color: #22c55e;
  font-size: 0.75rem;
  padding: 0.5rem 0.75rem;
  font-family: inherit;
}

.copy-link-btn {
  width: 40px;
  background: rgba(59, 130, 246, 0.2);
  border: 1px solid rgba(59, 130, 246, 0.3);
  color: #3b82f6;
  cursor: pointer;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.copy-link-btn:hover {
  background: rgba(59, 130, 246, 0.3);
}

.modal-footer {
  padding: 1rem 1.25rem;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.generate-btn {
  width: 100%;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8) !important;
  border: none !important;
}

.generate-btn:disabled {
  opacity: 0.5;
}

.wa-share-btn {
  width: 100%;
  background: linear-gradient(135deg, #22c55e, #16a34a) !important;
  border: none !important;
}

.add-manual-btn {
  width: 100%;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8) !important;
  border: none !important;
}

.add-manual-btn:disabled {
  opacity: 0.5;
}
</style>
