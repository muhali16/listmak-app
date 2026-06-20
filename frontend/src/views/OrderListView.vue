<template>
    <div class="order-list-container">
        <!-- Header -->
        <div class="page-header">
            <button
                class="back-btn"
                @click="goBack"
            >
                <i class="pi pi-arrow-left"></i>
                <span>Kembali</span>
            </button>
            <div class="header-title-row">
                <h1 class="page-title">
                    {{ listmakTitle }}
                </h1>
                <button
                    class="share-btn"
                    @click="openShare"
                >
                    <i
                        class="pi pi-share-alt"
                    ></i>
                    <span>Bagikan</span>
                </button>
            </div>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="state-block">
            <i class="pi pi-spin pi-spinner"></i>
            <p>Memuat pesanan...</p>
        </div>

        <!-- Error -->
        <div
            v-else-if="loadError"
            class="state-block state-error"
        >
            <i
                class="pi pi-exclamation-triangle"
            ></i>
            <p>{{ loadError }}</p>
            <button
                class="retry-btn"
                @click="loadData"
            >
                Coba lagi
            </button>
        </div>

        <template v-else>
            <!-- Active links section -->
            <div
                v-if="activeSharesLoaded && (activeShareLink || activeViewShare)"
                class="active-links-card"
            >
                <span class="active-links-label">🔗 Link aktif</span>
                <div v-if="activeShareLink" class="active-link-row">
                    <span class="active-link-icon">✏️</span>
                    <span class="active-link-text">Isi pesanan · {{ shareLinkExpiryLabel }}</span>
                    <div class="active-link-actions">
                        <button
                            class="active-link-btn"
                            @click="copyLinkUrl(shareLinkUrl)"
                            title="Salin link"
                        >
                            <i class="pi pi-copy"></i>
                        </button>
                        <button
                            class="active-link-btn"
                            @click="shareLinkViaWa('input', shareLinkUrl)"
                            title="Kirim WA"
                        >
                            <i class="pi pi-whatsapp"></i>
                        </button>
                    </div>
                </div>
                <div v-if="activeViewShare" class="active-link-row">
                    <span class="active-link-icon">👁️</span>
                    <span class="active-link-text">Lihat daftar</span>
                    <div class="active-link-actions">
                        <button
                            class="active-link-btn"
                            @click="copyLinkUrl(viewShareUrl)"
                            title="Salin link"
                        >
                            <i class="pi pi-copy"></i>
                        </button>
                        <button
                            class="active-link-btn"
                            @click="shareLinkViaWa('view', viewShareUrl)"
                            title="Kirim WA"
                        >
                            <i class="pi pi-whatsapp"></i>
                        </button>
                    </div>
                </div>
            </div>

            <!-- Summary 2-column -->
            <div class="summary-card">
                <div class="summary-item">
                    <span class="summary-value">{{
                        totalOrders
                    }}</span>
                    <span class="summary-label"
                        >Total pesanan</span
                    >
                </div>
                <div
                    class="summary-divider"
                ></div>
                <div class="summary-item">
                    <span
                        class="summary-value"
                        :class="{
                            'value-unpaid':
                                unpaidGroupCount >
                                0,
                        }"
                        >{{
                            unpaidGroupCount
                        }}</span
                    >
                    <span class="summary-label"
                        >Belum bayar</span
                    >
                </div>
            </div>

            <!-- Add order — full-width, explicit label -->
            <button
                class="add-btn"
                @click="openAddOrder"
            >
                <i class="pi pi-plus"></i>
                <span>Tambah pesanan</span>
            </button>

            <!-- Empty -->
            <div
                v-if="groups.length === 0"
                class="state-block empty"
            >
                <div class="empty-icon">
                    <i class="pi pi-inbox"></i>
                </div>
                <h3>Belum ada pesanan</h3>
                <p>
                    Tekan "+ Tambah pesanan" untuk
                    mulai.
                </p>
            </div>

            <!-- Grouped list -->
            <section v-else class="groups">
                <article
                    v-for="group in visibleGroups"
                    :key="group.key"
                    class="group-card"
                >
                    <!-- Group header: name, total, paid toggle -->
                    <div class="group-header">
                        <div class="group-meta">
                            <span
                                class="group-name"
                                >{{
                                    group.name
                                }}</span
                            >
                            <span
                                class="group-total"
                                >Rp
                                {{
                                    formatRupiah(
                                        group.total,
                                    )
                                }}</span
                            >
                        </div>
                        <button
                            class="paid-btn"
                            :class="{
                                'paid-btn--paid':
                                    group.allPaid &&
                                    !group.hasUnpriced,
                                'paid-btn--waiting':
                                    group.hasUnpriced,
                                'paid-btn--loading':
                                    group.loading,
                            }"
                            :disabled="
                                group.hasUnpriced ||
                                group.loading
                            "
                            @click="
                                togglePaid(group)
                            "
                        >
                            <i
                                v-if="
                                    group.loading
                                "
                                class="pi pi-spin pi-spinner"
                            ></i>
                            <i
                                v-else-if="
                                    group.hasUnpriced
                                "
                                class="pi pi-clock"
                            ></i>
                            <i
                                v-else-if="
                                    group.allPaid
                                "
                                class="pi pi-check-circle"
                            ></i>
                            <i
                                v-else
                                class="pi pi-circle"
                            ></i>
                            <span>{{
                                paidLabel(group)
                            }}</span>
                        </button>
                    </div>

                    <!-- Items — all shown, never collapsed -->
                    <ul class="item-list">
                        <li
                            v-for="order in group.orders"
                            :key="order.id"
                            class="item-row"
                        >
                            <div
                                class="item-info"
                            >
                                <span
                                    class="item-name"
                                    >{{
                                        order.order_detail
                                    }}</span
                                >
                                <span
                                    v-if="
                                        !order.price ||
                                        order.price ===
                                            0
                                    "
                                    class="item-no-price"
                                >
                                    <i
                                        class="pi pi-exclamation-circle"
                                    ></i>
                                    harga belum
                                    diisi
                                </span>
                                <span
                                    v-else
                                    class="item-price"
                                >
                                    Rp
                                    {{
                                        formatRupiah(
                                            (order.price ||
                                                0) *
                                                (order.qty ||
                                                    1),
                                        )
                                    }}
                                    <span
                                        v-if="
                                            order.qty >
                                            1
                                        "
                                        class="item-qty"
                                        >{{
                                            order.qty
                                        }}x</span
                                    >
                                </span>
                            </div>
                            <button
                                class="edit-btn"
                                @click="
                                    openEditOrder(
                                        order,
                                    )
                                "
                                title="Edit pesanan"
                            >
                                <i
                                    class="pi pi-pencil"
                                ></i>
                            </button>
                        </li>
                    </ul>
                </article>

                <!-- Explicit load-more — no infinite scroll -->
                <button
                    v-if="hasMore"
                    class="load-more-btn"
                    @click="loadMore"
                >
                    Muat lebih banyak ({{
                        groups.length -
                        visibleCount
                    }}
                    lagi)
                </button>
            </section>
        </template>

        <Toast position="top-center" />

        <!-- Share sheet backdrop -->
        <div
            v-if="showShareModal"
            class="sheet-backdrop"
            @click="closeShareModal"
        ></div>

        <!-- Share bottom sheet -->
        <div
            class="bottom-sheet"
            :class="{
                'bottom-sheet--open':
                    showShareModal,
            }"
        >
            <div class="sheet-handle"></div>

            <!-- Choose link type -->
            <div
                v-if="shareMode === 'choose'"
                class="sheet-section"
            >
                <h2 class="sheet-title">
                    Bagikan
                </h2>
                <p class="share-subtitle">
                    {{ listmakTitle }}
                </p>

                <div class="mode-buttons">
                    <button
                        class="mode-btn"
                        :disabled="shareLoading"
                        @click="createInputLink"
                    >
                        <i
                            class="pi pi-pencil"
                        ></i>
                        <span
                            class="mode-btn-label"
                            >Karyawan isi pesanan sendiri</span
                        >
                        <span
                            class="mode-btn-desc"
                            >Share ke grup, nanti karyawan tinggal klik dan isi pesanannya</span
                        >
                    </button>
                    <button
                        class="mode-btn"
                        :disabled="shareLoading"
                        @click="createViewLink"
                    >
                        <i class="pi pi-eye"></i>
                        <span
                            class="mode-btn-label"
                            >Lihat daftar pesanan</span
                        >
                        <span
                            class="mode-btn-desc"
                            >Buat yang cuma mau pantau daftar pesanan, tanpa bisa ubah apa-apa</span
                        >
                    </button>
                </div>

                <p
                    v-if="shareError"
                    class="form-error"
                >
                    {{ shareError }}
                </p>
                <div
                    v-if="shareLoading"
                    class="share-loading"
                >
                    <i
                        class="pi pi-spin pi-spinner"
                    ></i>
                    <span>Lagi bikin link...</span>
                </div>

                <button
                    class="sheet-cancel-btn"
                    :disabled="shareLoading"
                    @click="closeShareModal"
                >
                    Nanti dulu
                </button>
            </div>

            <!-- Result -->
            <div
                v-else-if="shareMode === 'result'"
                class="sheet-section"
            >
                <div class="share-result-icon">
                    <i
                        :class="
                            shareResult.type ===
                            'input'
                                ? 'pi pi-pencil'
                                : 'pi pi-eye'
                        "
                    ></i>
                </div>
                <h2 class="sheet-title">
                    Linknya udah jadi!
                </h2>
                <p class="share-result-type">
                    {{
                        shareResult.type === "input"
                            ? "Link isi pesanan · aktif 7 hari"
                            : "Link lihat daftar"
                    }}
                </p>

                <div class="share-url-box">
                    <span
                        class="share-url-text"
                        >{{
                            shareResult.url
                        }}</span
                    >
                </div>

                <button
                    class="submit-btn wa-btn"
                    @click="shareViaWhatsApp"
                >
                    <i class="pi pi-whatsapp"></i>
                    <span>Kirim ke WhatsApp</span>
                </button>

                <button
                    class="submit-btn"
                    style="margin-top: 0.5rem; background: rgba(30,41,59,0.8); border: 1px solid rgba(255,255,255,0.08);"
                    @click="copyShareUrl"
                >
                    <i class="pi pi-copy"></i>
                    <span>Salin link</span>
                </button>

                <button
                    class="sheet-cancel-btn"
                    style="margin-top: 0.75rem"
                    @click="closeShareModal"
                >
                    Tutup
                </button>
            </div>
        </div>

        <!-- Edit order backdrop -->
        <div
            v-if="showEditModal"
            class="sheet-backdrop"
            @click="closeEditModal"
        ></div>

        <!-- Edit order bottom sheet -->
        <div
            class="bottom-sheet"
            :class="{
                'bottom-sheet--open':
                    showEditModal,
            }"
        >
            <div class="sheet-handle"></div>
            <div class="sheet-section">
                <div class="sheet-nav">
                    <button
                        class="sheet-back-btn"
                        @click="closeEditModal"
                    >
                        <i
                            class="pi pi-arrow-left"
                        ></i>
                    </button>
                    <h2 class="sheet-title">
                        Edit pesanan
                    </h2>
                </div>

                <div class="form-group">
                    <label class="form-label"
                        >Nama pemesan</label
                    >
                    <input
                        v-model="editForm.name"
                        class="form-input"
                        type="text"
                        placeholder="contoh: Budi"
                        :disabled="editSubmitting"
                    />
                </div>
                <div class="form-group">
                    <label class="form-label"
                        >Pesanan</label
                    >
                    <input
                        v-model="
                            editForm.order_detail
                        "
                        class="form-input"
                        type="text"
                        placeholder="contoh: Nasi goreng"
                        :disabled="editSubmitting"
                    />
                </div>
                <div class="form-row">
                    <div
                        class="form-group form-group--flex"
                    >
                        <label class="form-label"
                            >Harga
                            <span
                                class="form-optional"
                                >(opsional)</span
                            ></label
                        >
                        <input
                            v-model.number="
                                editForm.price
                            "
                            class="form-input"
                            type="number"
                            placeholder="0"
                            min="0"
                            :disabled="
                                editSubmitting
                            "
                        />
                    </div>
                    <div
                        class="form-group form-group--flex"
                    >
                        <label class="form-label"
                            >Qty</label
                        >
                        <input
                            v-model.number="
                                editForm.qty
                            "
                            class="form-input"
                            type="number"
                            placeholder="1"
                            min="1"
                            :disabled="
                                editSubmitting
                            "
                        />
                    </div>
                </div>

                <p
                    v-if="editError"
                    class="form-error"
                >
                    {{ editError }}
                </p>

                <button
                    class="submit-btn"
                    :disabled="
                        editSubmitting ||
                        !editForm.name.trim() ||
                        !editForm.order_detail.trim()
                    "
                    @click="submitEdit"
                >
                    <i
                        v-if="editSubmitting"
                        class="pi pi-spin pi-spinner"
                    ></i>
                    <span>{{
                        editSubmitting
                            ? "Menyimpan..."
                            : "Simpan perubahan"
                    }}</span>
                </button>

                <button
                    class="delete-btn"
                    :disabled="editSubmitting"
                    @click="confirmDeleteOrder"
                >
                    <i class="pi pi-trash"></i>
                    <span>Hapus pesanan ini</span>
                </button>
            </div>
        </div>

        <!-- Bottom sheet backdrop -->
        <div
            v-if="showAddModal"
            class="sheet-backdrop"
            @click="closeAddModal"
        ></div>

        <!-- Add order bottom sheet -->
        <div
            class="bottom-sheet"
            :class="{
                'bottom-sheet--open':
                    showAddModal,
            }"
        >
            <div class="sheet-handle"></div>

            <!-- Mode choose -->
            <div
                v-if="addMode === 'choose'"
                class="sheet-section"
            >
                <h2 class="sheet-title">
                    Tambah Pesanan
                </h2>
                <div class="mode-buttons">
                    <button
                        class="mode-btn"
                        @click="
                            addMode = 'single'
                        "
                    >
                        <i class="pi pi-user"></i>
                        <span
                            class="mode-btn-label"
                            >Tambah
                            satu-satu</span
                        >
                        <span
                            class="mode-btn-desc"
                            >Isi satu form per
                            pesanan</span
                        >
                    </button>
                    <button
                        class="mode-btn"
                        @click="addMode = 'bulk'"
                    >
                        <i class="pi pi-list"></i>
                        <span
                            class="mode-btn-label"
                            >Tambah banyak
                            sekaligus</span
                        >
                        <span
                            class="mode-btn-desc"
                            >Paste atau ketik
                            beberapa baris</span
                        >
                    </button>
                </div>
                <button
                    class="sheet-cancel-btn"
                    @click="closeAddModal"
                >
                    Batal
                </button>
            </div>

            <!-- Single mode -->
            <div
                v-else-if="addMode === 'single'"
                class="sheet-section"
            >
                <div class="sheet-nav">
                    <button
                        class="sheet-back-btn"
                        @click="
                            addMode = 'choose'
                        "
                    >
                        <i
                            class="pi pi-arrow-left"
                        ></i>
                    </button>
                    <h2 class="sheet-title">
                        Tambah satu-satu
                    </h2>
                </div>

                <div class="form-group">
                    <label class="form-label"
                        >Nama pemesan</label
                    >
                    <input
                        v-model="singleForm.name"
                        class="form-input"
                        type="text"
                        placeholder="contoh: Budi"
                        :disabled="addSubmitting"
                    />
                </div>
                <div class="form-group">
                    <label class="form-label"
                        >Pesanan</label
                    >
                    <input
                        v-model="
                            singleForm.order_detail
                        "
                        class="form-input"
                        type="text"
                        placeholder="contoh: Nasi goreng"
                        :disabled="addSubmitting"
                    />
                </div>
                <div class="form-row">
                    <div
                        class="form-group form-group--flex"
                    >
                        <label class="form-label"
                            >Harga
                            <span
                                class="form-optional"
                                >(opsional)</span
                            ></label
                        >
                        <input
                            v-model.number="
                                singleForm.price
                            "
                            class="form-input"
                            type="number"
                            placeholder="0"
                            min="0"
                            :disabled="
                                addSubmitting
                            "
                        />
                    </div>
                    <div
                        class="form-group form-group--flex"
                    >
                        <label class="form-label"
                            >Qty</label
                        >
                        <input
                            v-model.number="
                                singleForm.qty
                            "
                            class="form-input"
                            type="number"
                            placeholder="1"
                            min="1"
                            :disabled="
                                addSubmitting
                            "
                        />
                    </div>
                </div>

                <p
                    v-if="addError"
                    class="form-error"
                >
                    {{ addError }}
                </p>

                <button
                    class="submit-btn"
                    :disabled="
                        addSubmitting ||
                        !singleForm.name.trim() ||
                        !singleForm.order_detail.trim()
                    "
                    @click="submitSingle"
                >
                    <i
                        v-if="addSubmitting"
                        class="pi pi-spin pi-spinner"
                    ></i>
                    <span>{{
                        addSubmitting
                            ? "Menyimpan..."
                            : "Simpan pesanan"
                    }}</span>
                </button>
            </div>

            <!-- Bulk mode -->
            <div
                v-else-if="addMode === 'bulk'"
                class="sheet-section"
            >
                <div class="sheet-nav">
                    <button
                        class="sheet-back-btn"
                        @click="
                            addMode = 'choose'
                        "
                    >
                        <i
                            class="pi pi-arrow-left"
                        ></i>
                    </button>
                    <h2 class="sheet-title">
                        Tambah banyak sekaligus
                    </h2>
                </div>

                <p class="bulk-hint">
                    Paste langsung dari WA/chat.
                    Format per baris:
                    <code
                        >1. Nama: Detail
                        pesanan</code
                    >
                    — baris lain diabaikan
                    otomatis.
                </p>

                <textarea
                    v-model="bulkText"
                    class="bulk-textarea"
                    placeholder="1. Rachel: nasi soto daging&#10;2. Dimas: ayam madura dada&#10;3. Nastha: ketoprak telur dadar"
                    rows="8"
                    :disabled="addSubmitting"
                ></textarea>

                <!-- Preview -->
                <div
                    v-if="
                        parsedBulkOrders.length >
                        0
                    "
                    class="bulk-preview"
                >
                    <p class="bulk-preview-title">
                        {{
                            parsedBulkOrders.length
                        }}
                        pesanan akan ditambahkan:
                    </p>
                    <ul class="bulk-preview-list">
                        <li
                            v-for="(
                                o, i
                            ) in parsedBulkOrders"
                            :key="i"
                            class="bulk-preview-item"
                        >
                            <span
                                class="preview-name"
                                >{{
                                    o.name
                                }}</span
                            >
                            <span
                                class="preview-detail"
                                >{{
                                    o.order_detail
                                }}</span
                            >
                            <span
                                v-if="o.price"
                                class="preview-price"
                                >Rp
                                {{
                                    formatRupiah(
                                        o.price,
                                    )
                                }}</span
                            >
                            <span
                                v-else
                                class="preview-no-price"
                                >harga belum
                                diisi</span
                            >
                        </li>
                    </ul>
                </div>

                <p
                    v-if="addError"
                    class="form-error"
                >
                    {{ addError }}
                </p>

                <button
                    class="submit-btn"
                    :disabled="
                        addSubmitting ||
                        parsedBulkOrders.length ===
                            0
                    "
                    @click="submitBulk"
                >
                    <i
                        v-if="addSubmitting"
                        class="pi pi-spin pi-spinner"
                    ></i>
                    <span>{{
                        addSubmitting
                            ? "Menyimpan..."
                            : `Simpan ${parsedBulkOrders.length} pesanan`
                    }}</span>
                </button>
            </div>
        </div>
    </div>
</template>

<script>
import Toast from "primevue/toast";
import { listmak, share } from "../api";

export default {
    name: "OrderListView",
    components: { Toast },

    data() {
        return {
            listmakTitle: "",
            orders: [],
            loading: false,
            loadError: "",
            loadingGroups: {},
            visibleCount: 10,
            showAddModal: false,
            addMode: "choose",
            singleForm: {
                name: "",
                order_detail: "",
                price: "",
                qty: 1,
            },
            bulkText: "",
            addSubmitting: false,
            addError: "",
            // edit modal
            showEditModal: false,
            editForm: {
                id: null,
                name: "",
                order_detail: "",
                price: "",
                qty: 1,
            },
            editSubmitting: false,
            editError: "",
            // share modal
            showShareModal: false,
            shareMode: "choose",
            shareResult: null,
            shareLoading: false,
            shareError: "",
            // active share state
            activeShareLink: null,
            activeViewShare: null,
            activeSharesLoaded: false,
            // expiry picker
            selectedExpiry: "7d",
            customExpiryDate: "",
            forceNew: false,
        };
    },

    computed: {
        listmakId() {
            return Number(this.$route.params.id);
        },

        groups() {
            const groupMap = new Map();
            for (const order of this.orders) {
                const key = order.name
                    .trim()
                    .toLowerCase();
                if (!groupMap.has(key)) {
                    groupMap.set(key, {
                        name: order.name.trim(),
                        orders: [],
                    });
                }
                groupMap
                    .get(key)
                    .orders.push(order);
            }

            return Array.from(
                groupMap.values(),
            ).map((g) => {
                const key = g.name.toLowerCase();
                const hasUnpriced = g.orders.some(
                    (o) =>
                        !o.price || o.price === 0,
                );
                const allPaid = g.orders.every(
                    (o) => o.is_paid,
                );
                const total = g.orders.reduce(
                    (sum, o) =>
                        sum +
                        (o.price || 0) *
                            (o.qty || 1),
                    0,
                );
                return {
                    key,
                    name: g.name,
                    orders: g.orders,
                    hasUnpriced,
                    allPaid,
                    total,
                    loading:
                        !!this.loadingGroups[key],
                };
            });
        },

        visibleGroups() {
            return this.groups.slice(
                0,
                this.visibleCount,
            );
        },

        hasMore() {
            return (
                this.visibleCount <
                this.groups.length
            );
        },

        totalOrders() {
            return this.orders.length;
        },

        unpaidGroupCount() {
            return this.groups.filter(
                (g) => !g.allPaid,
            ).length;
        },

        parsedBulkOrders() {
            const NUMBERED_RE =
                /^\d+\.\s+(.+?):\s+(.+)$/;
            const INVISIBLE = /[​‌‍⁠﻿]/g;
            return this.bulkText
                .split("\n")
                .map((line) =>
                    line
                        .replace(INVISIBLE, "")
                        .trim(),
                )
                .filter((line) => line)
                .map((line) => {
                    const m =
                        line.match(NUMBERED_RE);
                    if (!m) return null;
                    return {
                        name: m[1].trim(),
                        order_detail: m[2].trim(),
                        price: 0,
                        qty: 1,
                    };
                })
                .filter(Boolean);
        },

        shareLinkUrl() {
            if (!this.activeShareLink) return "";
            return `${window.location.origin}/listmak/order/${this.activeShareLink.share_id}`;
        },
        viewShareUrl() {
            if (!this.activeViewShare) return "";
            return `${window.location.origin}/listmak/view/${this.activeViewShare.view_id}`;
        },
        shareLinkExpiryLabel() {
            if (!this.activeShareLink) return "";
            const msLeft =
                new Date(this.activeShareLink.expires_at) - Date.now();
            const daysLeft = Math.ceil(
                msLeft / (1000 * 60 * 60 * 24),
            );
            if (daysLeft <= 0) return "kedaluwarsa";
            if (daysLeft === 1) return "exp. besok";
            return `exp. ${daysLeft} hari lagi`;
        },
        expiryOptions() {
            return [
                { value: "1d", label: "1 hari" },
                { value: "3d", label: "3 hari" },
                { value: "7d", label: "7 hari" },
                { value: "30d", label: "30 hari" },
            ];
        },
        minCustomExpiry() {
            const now = new Date(Date.now() + 60 * 60 * 1000);
            return now.toISOString().slice(0, 16);
        },
    },

    mounted() {
        this.loadData();
        this.loadActiveShares();
    },

    methods: {
        async loadActiveShares() {
            try {
                const res = await share.getActiveShares(
                    this.listmakId,
                );
                if (res.success && res.data) {
                    this.activeShareLink =
                        res.data.share_link;
                    this.activeViewShare =
                        res.data.view_share;
                }
            } catch {
                // silent — page still works without active shares info
            } finally {
                this.activeSharesLoaded = true;
            }
        },

        async loadData() {
            this.loading = true;
            this.loadError = "";
            try {
                const [lmRes, ordersRes] =
                    await Promise.all([
                        listmak.getListMakById(
                            this.listmakId,
                        ),
                        listmak.getOrders(
                            this.listmakId,
                        ),
                    ]);
                if (lmRes.success && lmRes.data) {
                    this.listmakTitle =
                        lmRes.data.title ||
                        `Listmak #${this.listmakId}`;
                }
                if (
                    ordersRes.success &&
                    ordersRes.data
                ) {
                    this.orders = Array.isArray(
                        ordersRes.data,
                    )
                        ? ordersRes.data
                        : [];
                }
            } catch {
                this.loadError =
                    "Gagal memuat data. Periksa koneksi lalu coba lagi.";
            } finally {
                this.loading = false;
            }
        },

        async refreshOrders() {
            try {
                const res =
                    await listmak.getOrders(
                        this.listmakId,
                    );
                if (res.success && res.data) {
                    this.orders = Array.isArray(
                        res.data,
                    )
                        ? res.data
                        : [];
                }
            } catch (err) {
                console.error(
                    "Failed to refresh orders:",
                    err,
                );
            }
        },

        async togglePaid(group) {
            const key = group.key;
            this.loadingGroups = {
                ...this.loadingGroups,
                [key]: true,
            };
            try {
                await listmak.updateOrdersPaidByName(
                    this.listmakId,
                    group.name,
                    !group.allPaid,
                );
                await this.refreshOrders();
            } catch (err) {
                this.$toast.add({
                    severity: "error",
                    summary: "Gagal update",
                    detail:
                        err.message ||
                        "Gagal mengubah status bayar. Coba lagi.",
                    life: 3000,
                });
            } finally {
                const { [key]: _, ...rest } =
                    this.loadingGroups;
                this.loadingGroups = rest;
            }
        },

        paidLabel(group) {
            if (group.loading) return "...";
            if (group.hasUnpriced)
                return "Tunggu";
            return group.allPaid
                ? "Lunas"
                : "Belum";
        },

        formatRupiah(amount) {
            return Number(
                amount || 0,
            ).toLocaleString("id-ID");
        },

        loadMore() {
            this.visibleCount += 10;
        },

        goBack() {
            this.$router.push("/today");
        },

        buildLinkMessage(type, url) {
            const title = this.listmakTitle;
            return type === "input"
                ? `Halo semua! 👋\n\nAda listmak baru nih: *${title}*\n\nSilakan isi pesanan kalian di sini ya:\n${url}\n\n_Klik linknya, masukin nama dan pesanan, selesai deh~_`
                : `Hai! Ada listmak nih: *${title}*\n\nIni daftar pesanannya ya:\n${url}\n\n_Bisa dipantau dari sini, tapi cuma bisa lihat aja ya._`;
        },
        async copyLinkUrl(url) {
            try {
                await navigator.clipboard.writeText(url);
                this.$toast.add({
                    severity: "success",
                    summary: "Link disalin!",
                    life: 2000,
                });
            } catch {
                this.$toast.add({
                    severity: "error",
                    summary: "Gagal menyalin",
                    life: 2000,
                });
            }
        },
        shareLinkViaWa(type, url) {
            window.open(
                `https://wa.me/?text=${encodeURIComponent(this.buildLinkMessage(type, url))}`,
                "_blank",
            );
        },

        openShare() {
            this.showShareModal = true;
            this.shareMode = "choose";
            this.shareResult = null;
            this.shareError = "";
            this.forceNew = false;
        },

        closeShareModal() {
            if (!this.shareLoading) {
                this.showShareModal = false;
                this.shareMode = "choose";
            }
        },

        async createInputLink() {
            this.shareLoading = true;
            this.shareError = "";
            try {
                const expiresAt = new Date(
                    Date.now() +
                        7 * 24 * 60 * 60 * 1000,
                ).toISOString();
                const res =
                    await share.createShareLink({
                        listmak_id:
                            this.listmakId,
                        title: this.listmakTitle,
                        expires_at: expiresAt,
                    });
                const shareId = res.data.share_id;
                const url = `${window.location.origin}/listmak/order/${shareId}`;
                this.shareResult = {
                    url,
                    type: "input",
                };
                this.shareMode = "result";
            } catch (err) {
                this.shareError =
                    err.message ||
                    "Gagal membuat link. Coba lagi.";
            } finally {
                this.shareLoading = false;
            }
        },

        async createViewLink() {
            this.shareLoading = true;
            this.shareError = "";
            try {
                const res =
                    await share.createViewShare({
                        listmak_id:
                            this.listmakId,
                        title: this.listmakTitle,
                    });
                const viewId = res.data.view_id;
                const url = `${window.location.origin}/listmak/view/${viewId}`;
                this.shareResult = {
                    url,
                    type: "view",
                };
                this.shareMode = "result";
            } catch (err) {
                this.shareError =
                    err.message ||
                    "Gagal membuat link. Coba lagi.";
            } finally {
                this.shareLoading = false;
            }
        },

        buildShareMessage() {
            const isInput = this.shareResult.type === "input";
            const title = this.listmakTitle;
            return isInput
                ? `Halo semua! 👋\n\nAda listmak baru nih: *${title}*\n\nSilakan isi pesanan kalian di sini ya:\n${this.shareResult.url}\n\n_Klik linknya, masukin nama dan pesanan, selesai deh~_`
                : `Hai! Ada listmak nih: *${title}*\n\nIni daftar pesanannya ya:\n${this.shareResult.url}\n\n_Bisa dipantau dari sini, tapi cuma bisa lihat aja ya._`;
        },

        async copyShareUrl() {
            const text = this.buildShareMessage();
            try {
                await navigator.clipboard.writeText(text);
                this.$toast.add({
                    severity: "success",
                    summary: "Teks disalin!",
                    life: 2000,
                });
            } catch {
                this.$toast.add({
                    severity: "error",
                    summary: "Gagal menyalin. Salin manual.",
                    life: 2000,
                });
            }
        },

        shareViaWhatsApp() {
            window.open(
                `https://wa.me/?text=${encodeURIComponent(this.buildShareMessage())}`,
                "_blank",
            );
        },

        openAddOrder() {
            this.showAddModal = true;
            this.addMode = "choose";
            this.singleForm = {
                name: "",
                order_detail: "",
                price: "",
                qty: 1,
            };
            this.bulkText = "";
            this.addError = "";
        },

        closeAddModal() {
            if (!this.addSubmitting) {
                this.showAddModal = false;
                this.addMode = "choose";
            }
        },

        async submitSingle() {
            this.addSubmitting = true;
            this.addError = "";
            try {
                await listmak.addOrder(
                    this.listmakId,
                    {
                        name: this.singleForm.name.trim(),
                        order_detail:
                            this.singleForm.order_detail.trim(),
                        price:
                            this.singleForm
                                .price || 0,
                        qty:
                            this.singleForm.qty ||
                            1,
                    },
                );
                await this.refreshOrders();
                this.singleForm = {
                    name: "",
                    order_detail: "",
                    price: "",
                    qty: 1,
                };
                this.showAddModal = false;
            } catch (err) {
                this.addError =
                    err.message ||
                    "Gagal menyimpan. Coba lagi.";
            } finally {
                this.addSubmitting = false;
            }
        },

        async submitBulk() {
            if (!this.parsedBulkOrders.length)
                return;
            this.addSubmitting = true;
            this.addError = "";
            try {
                await listmak.addBulkOrders(
                    this.listmakId,
                    this.parsedBulkOrders,
                );
                await this.refreshOrders();
                this.bulkText = "";
                this.showAddModal = false;
            } catch (err) {
                this.addError =
                    err.message ||
                    "Gagal menyimpan. Coba lagi.";
            } finally {
                this.addSubmitting = false;
            }
        },

        openEditOrder(order) {
            this.editForm = {
                id: order.id,
                name: order.name,
                order_detail: order.order_detail,
                price: order.price || "",
                qty: order.qty || 1,
            };
            this.editError = "";
            this.showEditModal = true;
        },

        closeEditModal() {
            if (!this.editSubmitting) {
                this.showEditModal = false;
            }
        },

        async submitEdit() {
            this.editSubmitting = true;
            this.editError = "";
            try {
                await listmak.updateOrder(
                    this.editForm.id,
                    {
                        name: this.editForm.name.trim(),
                        order_detail:
                            this.editForm.order_detail.trim(),
                        price:
                            this.editForm.price ||
                            0,
                        qty:
                            this.editForm.qty ||
                            1,
                    },
                );
                await this.refreshOrders();
                this.showEditModal = false;
            } catch (err) {
                this.editError =
                    err.message ||
                    "Gagal menyimpan. Coba lagi.";
            } finally {
                this.editSubmitting = false;
            }
        },

        async confirmDeleteOrder() {
            if (
                !confirm(
                    `Hapus pesanan "${this.editForm.order_detail}"?`,
                )
            )
                return;
            this.editSubmitting = true;
            this.editError = "";
            try {
                await listmak.deleteOrder(
                    this.editForm.id,
                );
                await this.refreshOrders();
                this.showEditModal = false;
            } catch (err) {
                this.editError =
                    err.message ||
                    "Gagal menghapus. Coba lagi.";
            } finally {
                this.editSubmitting = false;
            }
        },
    },
};
</script>

<style scoped>
.order-list-container {
    padding: 1rem;
    padding-bottom: 2rem;
}

/* Header */
.page-header {
    margin-bottom: 1.25rem;
}

.back-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.375rem 0;
    background: transparent;
    border: none;
    color: #94a3b8;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    margin-bottom: 0.5rem;
}

.back-btn:hover {
    color: #cbd5e1;
}

.header-title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
}

.page-title {
    font-size: 1.375rem;
    font-weight: 700;
    color: #f1f5f9;
    overflow-wrap: anywhere;
    flex: 1;
    min-width: 0;
}

.share-btn {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.5rem 0.875rem;
    background: rgba(99, 102, 241, 0.15);
    border: 1px solid rgba(99, 102, 241, 0.3);
    border-radius: 0.625rem;
    color: #818cf8;
    font-size: 0.8125rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s;
}

.share-btn:hover {
    background: rgba(99, 102, 241, 0.25);
}

/* Summary card */
.summary-card {
    display: flex;
    align-items: stretch;
    background: rgba(30, 41, 59, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.05);
    border-radius: 0.875rem;
    padding: 1rem;
    margin-bottom: 1rem;
}

.summary-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
}

.summary-value {
    font-size: 1.75rem;
    font-weight: 700;
    color: #f1f5f9;
    line-height: 1;
}

.summary-value.value-unpaid {
    color: #eab308;
}

.summary-label {
    font-size: 0.75rem;
    color: #64748b;
    text-align: center;
}

.summary-divider {
    width: 1px;
    background: rgba(255, 255, 255, 0.06);
    margin: 0 0.5rem;
}

/* Add order button — full-width, labelled */
.add-btn {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.875rem 1rem;
    margin-bottom: 1.25rem;
    background: linear-gradient(
        135deg,
        #3b82f6,
        #1d4ed8
    );
    border: none;
    border-radius: 0.875rem;
    color: white;
    font-size: 0.9375rem;
    font-weight: 600;
    cursor: pointer;
    transition:
        transform 0.15s,
        opacity 0.15s;
}

.add-btn:hover {
    opacity: 0.95;
    transform: translateY(-1px);
}

/* Groups */
.groups {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.group-card {
    background: rgba(30, 41, 59, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.05);
    border-radius: 0.875rem;
    overflow: hidden;
}

.group-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.875rem 1rem;
    border-bottom: 1px solid
        rgba(255, 255, 255, 0.04);
}

.group-meta {
    flex: 1;
    min-width: 0;
}

.group-name {
    display: block;
    font-size: 0.9375rem;
    font-weight: 700;
    color: #f1f5f9;
    overflow-wrap: anywhere;
}

.group-total {
    display: block;
    font-size: 0.8125rem;
    color: #94a3b8;
    margin-top: 0.125rem;
}

/* Paid toggle button */
.paid-btn {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0.5rem 0.75rem;
    border-radius: 0.5rem;
    font-size: 0.8125rem;
    font-weight: 600;
    cursor: pointer;
    transition:
        background 0.15s,
        opacity 0.15s;
    border: 1px solid transparent;
    /* default: belum bayar */
    background: rgba(234, 179, 8, 0.12);
    border-color: rgba(234, 179, 8, 0.25);
    color: #eab308;
}

.paid-btn--paid {
    background: rgba(34, 197, 94, 0.12);
    border-color: rgba(34, 197, 94, 0.25);
    color: #22c55e;
}

.paid-btn--waiting {
    background: rgba(100, 116, 139, 0.12);
    border-color: rgba(100, 116, 139, 0.2);
    color: #64748b;
    cursor: not-allowed;
}

.paid-btn--loading {
    opacity: 0.6;
    cursor: not-allowed;
}

/* Item list */
.item-list {
    list-style: none;
    padding: 0;
    margin: 0;
}

.item-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.625rem 1rem;
    border-bottom: 1px solid
        rgba(255, 255, 255, 0.03);
}

.item-row:last-child {
    border-bottom: none;
}

.item-info {
    flex: 1;
    min-width: 0;
}

.item-name {
    display: block;
    font-size: 0.875rem;
    color: #e2e8f0;
    overflow-wrap: anywhere;
}

.item-price {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.8125rem;
    color: #94a3b8;
    margin-top: 0.15rem;
}

.item-qty {
    font-size: 0.75rem;
    color: #64748b;
}

.item-no-price {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.75rem;
    color: #f97316;
    margin-top: 0.15rem;
}

.item-no-price i {
    font-size: 0.75rem;
}

.edit-btn {
    flex-shrink: 0;
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 0.5rem;
    color: #64748b;
    cursor: pointer;
    transition:
        color 0.15s,
        background 0.15s;
}

.edit-btn:hover {
    color: #94a3b8;
    background: rgba(255, 255, 255, 0.08);
}

.edit-btn i {
    font-size: 0.8rem;
}

/* Load more */
.load-more-btn {
    width: 100%;
    padding: 0.875rem;
    background: rgba(30, 41, 59, 0.4);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 0.875rem;
    color: #94a3b8;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition:
        background 0.15s,
        color 0.15s;
}

.load-more-btn:hover {
    background: rgba(30, 41, 59, 0.7);
    color: #cbd5e1;
}

/* State blocks */
.state-block {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
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
    margin-top: 0.5rem;
    padding: 0.6rem 1.1rem;
    background: rgba(59, 130, 246, 0.15);
    border: 1px solid rgba(59, 130, 246, 0.3);
    border-radius: 0.625rem;
    color: #3b82f6;
    font-weight: 600;
    font-size: 0.875rem;
    cursor: pointer;
}

.empty .empty-icon {
    width: 72px;
    height: 72px;
    border-radius: 50%;
    background: rgba(30, 41, 59, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
}

.empty .empty-icon i {
    font-size: 1.75rem;
    color: #475569;
}

.empty h3 {
    font-size: 1.0625rem;
    font-weight: 600;
    color: #f1f5f9;
}

.empty p {
    font-size: 0.875rem;
    color: #64748b;
}

@media (min-width: 768px) {
    .order-list-container {
        padding: 1.5rem 2rem;
        max-width: 720px;
    }
}

/* Bottom sheet */
.sheet-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    z-index: 100;
}

.bottom-sheet {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    background: #0f172a;
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 1.25rem 1.25rem 0 0;
    z-index: 101;
    max-height: 90vh;
    overflow-y: auto;
    transform: translateY(100%);
    transition: transform 0.28s ease;
}

.bottom-sheet--open {
    transform: translateY(0);
}

.sheet-handle {
    width: 40px;
    height: 4px;
    background: rgba(255, 255, 255, 0.12);
    border-radius: 2px;
    margin: 0.75rem auto 0;
}

.sheet-section {
    padding: 1rem 1.25rem 2rem;
}

.sheet-nav {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 1.25rem;
}

.sheet-back-btn {
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

.sheet-title {
    font-size: 1.0625rem;
    font-weight: 700;
    color: #f1f5f9;
    margin: 0 0 1.25rem;
}

.sheet-nav .sheet-title {
    margin: 0;
}

/* Mode choose buttons */
.mode-buttons {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 1rem;
}

.mode-btn {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.25rem;
    padding: 1rem 1.125rem;
    background: rgba(30, 41, 59, 0.7);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 0.875rem;
    cursor: pointer;
    text-align: left;
    transition:
        background 0.15s,
        border-color 0.15s;
}

.mode-btn:hover {
    background: rgba(30, 41, 59, 1);
    border-color: rgba(99, 102, 241, 0.3);
}

.mode-btn i {
    font-size: 1.25rem;
    color: #818cf8;
    margin-bottom: 0.25rem;
}

.mode-btn-label {
    font-size: 0.9375rem;
    font-weight: 600;
    color: #f1f5f9;
}

.mode-btn-desc {
    font-size: 0.8125rem;
    color: #64748b;
}

.sheet-cancel-btn {
    width: 100%;
    padding: 0.75rem;
    background: transparent;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 0.75rem;
    color: #64748b;
    font-size: 0.875rem;
    cursor: pointer;
}

/* Form */
.form-group {
    margin-bottom: 1rem;
}

.form-row {
    display: flex;
    gap: 0.75rem;
}

.form-group--flex {
    flex: 1;
}

.form-label {
    display: block;
    font-size: 0.8125rem;
    font-weight: 600;
    color: #94a3b8;
    margin-bottom: 0.375rem;
}

.form-optional {
    font-weight: 400;
    color: #475569;
}

.form-input {
    width: 100%;
    padding: 0.625rem 0.75rem;
    background: rgba(30, 41, 59, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 0.625rem;
    color: #f1f5f9;
    font-size: 0.9375rem;
    box-sizing: border-box;
    transition: border-color 0.15s;
}

.form-input:focus {
    outline: none;
    border-color: rgba(99, 102, 241, 0.5);
}

.form-input:disabled {
    opacity: 0.5;
}

.form-input::placeholder {
    color: #334155;
}

.form-error {
    font-size: 0.8125rem;
    color: #ef4444;
    margin-bottom: 0.75rem;
}

/* Bulk textarea */
.bulk-hint {
    font-size: 0.8125rem;
    color: #64748b;
    margin-bottom: 0.75rem;
}

.bulk-hint code {
    background: rgba(255, 255, 255, 0.06);
    padding: 0.125rem 0.375rem;
    border-radius: 0.3rem;
    font-size: 0.8125rem;
    color: #94a3b8;
}

.bulk-textarea {
    width: 100%;
    padding: 0.75rem;
    background: rgba(30, 41, 59, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 0.625rem;
    color: #f1f5f9;
    font-size: 0.875rem;
    font-family: inherit;
    resize: vertical;
    box-sizing: border-box;
    margin-bottom: 1rem;
}

.bulk-textarea:focus {
    outline: none;
    border-color: rgba(99, 102, 241, 0.5);
}

.bulk-textarea:disabled {
    opacity: 0.5;
}

/* Bulk preview */
.bulk-preview {
    background: rgba(30, 41, 59, 0.5);
    border: 1px solid rgba(255, 255, 255, 0.05);
    border-radius: 0.625rem;
    padding: 0.75rem;
    margin-bottom: 1rem;
}

.bulk-preview-title {
    font-size: 0.75rem;
    font-weight: 600;
    color: #64748b;
    margin-bottom: 0.5rem;
}

.bulk-preview-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
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
}

.preview-detail {
    flex: 1;
    color: #94a3b8;
}

.preview-price {
    color: #64748b;
    font-size: 0.75rem;
}

.preview-no-price {
    color: #f97316;
    font-size: 0.75rem;
}

/* Share modal specifics */
.share-subtitle {
    font-size: 0.875rem;
    color: #64748b;
    margin: -0.75rem 0 1.25rem;
}

.share-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    color: #64748b;
    font-size: 0.875rem;
    padding: 0.5rem 0;
}

.share-result-icon {
    width: 3rem;
    height: 3rem;
    background: rgba(99, 102, 241, 0.12);
    border-radius: 0.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 0.75rem;
}

.share-result-icon i {
    font-size: 1.25rem;
    color: #818cf8;
}

.share-result-type {
    font-size: 0.8125rem;
    color: #64748b;
    margin: -0.5rem 0 1.25rem;
}

.share-url-box {
    background: rgba(30, 41, 59, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 0.625rem;
    padding: 0.75rem;
    margin-bottom: 1rem;
    word-break: break-all;
}

.share-url-text {
    font-size: 0.8125rem;
    color: #94a3b8;
    font-family: monospace;
}

/* WhatsApp button */
.wa-btn {
    background: linear-gradient(
        135deg,
        #25d366,
        #128c3e
    ) !important;
}

/* Submit button */
.submit-btn {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.875rem;
    background: linear-gradient(
        135deg,
        #3b82f6,
        #1d4ed8
    );
    border: none;
    border-radius: 0.875rem;
    color: white;
    font-size: 0.9375rem;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.15s;
    margin-top: 0.5rem;
}

.submit-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
}

.delete-btn {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.75rem;
    background: transparent;
    border: 1px solid rgba(239, 68, 68, 0.25);
    border-radius: 0.875rem;
    color: #ef4444;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    margin-top: 0.75rem;
    transition: background 0.15s;
}

.delete-btn:hover {
    background: rgba(239, 68, 68, 0.08);
}

.delete-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
}

/* ── Active links section ── */
.active-links-card {
    background: rgba(30, 41, 59, 0.6);
    border: 1px solid rgba(99, 179, 237, 0.15);
    border-radius: 0.875rem;
    padding: 0.75rem 1rem;
    margin-bottom: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.active-links-label {
    font-size: 0.75rem;
    font-weight: 700;
    color: #63b3ed;
    letter-spacing: 0.04em;
    text-transform: uppercase;
}

.active-link-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.active-link-icon {
    font-size: 0.9rem;
}

.active-link-text {
    flex: 1;
    font-size: 0.8125rem;
    color: #94a3b8;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.active-link-actions {
    display: flex;
    gap: 0.375rem;
    flex-shrink: 0;
}

.active-link-btn {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 0.5rem;
    color: #94a3b8;
    padding: 0.375rem 0.5rem;
    cursor: pointer;
    font-size: 0.8125rem;
    display: flex;
    align-items: center;
    transition: background 0.15s;
}

.active-link-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #e2e8f0;
}
</style>
