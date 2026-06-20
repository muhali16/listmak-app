# Design: Auto Vendor Categorization + Food Suggestions

**Date:** 2026-06-21  
**Status:** Approved

---

## Overview

Dua fitur yang saling berkaitan:

1. **Auto Vendor Categorization** — LLM (Fireworks.ai) mengekstrak nama resto/vendor dari `order_detail` secara async setelah order disimpan. Hasilnya tersimpan ke DB. OB bisa override manual.
2. **Food Suggestions + Restaurant Nudge** — SharedListMakView.vue menampilkan autocomplete makanan dari histori order (Postgres ILIKE), dan mengingatkan user jika belum mencantumkan lokasi beli.

---

## Architecture Overview

```
orders.vendor_name (new nullable column)

Fitur A — Auto Vendor (backend):
  Submit order → save DB → goroutine → Fireworks.ai API → UPDATE vendor_name
  (non-blocking: order langsung tersimpan, vendor diisi async)

Fitur B — Food Suggestions (frontend + backend):
  User ketik → debounce 300ms → GET /share/:id/food-suggestions?q=X
  → Postgres ILIKE → dropdown max 8 item
  Priority: same listmak (5 item) → global fallback (padded ke 8)

Fitur B2 — Restaurant Nudge (frontend only):
  Single mode: onBlur order_detail → JS keyword check → hint text
  Bulk mode: debounce 1s → parse orders → JS keyword check per order → warning chip
```

---

## Data Model

### Schema Migration

```sql
ALTER TABLE orders ADD COLUMN vendor_name VARCHAR(100);

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_orders_detail_trgm ON orders USING GIN (order_detail gin_trgm_ops);
```

### Go Model (`internal/models/order_model.go`)

Tambah 1 field:

```go
VendorName string `gorm:"type:varchar(100)" json:"vendor_name"`
```

---

## Backend

### Env Vars Baru

```
FIREWORKS_API_KEY=fw_xxx
FIREWORKS_MODEL=accounts/fireworks/models/llama-v3p1-8b-instruct
```

### File Baru: `internal/services/ai_service.go`

```go
type AIService struct {
    apiKey string
    model  string
    client *http.Client
}

func (s *AIService) ExtractVendor(orderDetail string) (string, error) {
    // POST https://api.fireworks.ai/inference/v1/chat/completions
    // Prompt:
    // "Dari pesanan makanan berikut, ekstrak nama resto atau kategori tempatnya.
    //  Jika ada nama tempat/orang (contoh: 'pak donan', 'madura') → pakai itu.
    //  Jika tidak ada → infer dari jenis makanan (contoh: 'soto daging' → 'Soto',
    //  'dimsum' → 'Dimsum'). Jawab HANYA nama tempatnya saja, tanpa penjelasan.
    //  Pesanan: <orderDetail>"
    // Timeout: 10s
    // Return: trimmed string vendor name
}
```

### Perubahan Order Service

Setelah `CreateOrder()` berhasil, spawn goroutine:

```go
go func(id uint, detail string) {
    vendor, err := s.ai.ExtractVendor(detail)
    if err != nil || vendor == "" {
        return // fail silently, log error only
    }
    s.repo.UpdateVendorName(id, vendor)
}(order.ID, order.OrderDetail)
```

### Perubahan Order Repository

Tambah method:

```go
func (r *OrderRepository) UpdateVendorName(id uint, vendor string) error
```

### Endpoint Baru 1: Food Suggestions

```
GET /api/v1/share/:shareId/food-suggestions?q=<query>
Auth: Share token validation (sama seperti endpoint share lainnya)
```

Logic:
1. Validate share token → dapat `listmak_id`
2. Query same listmak (`ILIKE '%q%'`, DISTINCT, limit 5)
3. Jika hasil < 8 → pad dengan global results (exclude duplikat, limit sisa)
4. Return array string

Response:
```json
{ "success": true, "data": ["Nasi goreng", "Ayam madura dada", "Soto daging"] }
```

Jika `q` kosong → return popular items dari listmak (ORDER BY frequency/count DESC, limit 8).

### Endpoint Baru 2: Override Vendor

```
PATCH /api/v1/orders/:id/vendor
Auth: JWT, hanya listmak owner
Body: { "vendor_name": "Pak Donan" }
```

---

## Frontend

### SharedListMakView.vue — Single Mode

**Autocomplete field:**
- `<textarea>` order_detail tetap (jangan ganti ke `<input>` — konten bisa panjang multi-line)
- Tambahkan floating dropdown di bawah textarea (absolute positioned)
- On focus → fetch suggestions (query kosong, ambil popular items)
- On input → debounce 300ms, min 2 char → fetch suggestions berdasarkan kata terakhir yang diketik
- Dropdown max 8 item, tap to append/replace ke textarea
- Tetap bisa ketik bebas (bukan forced select)
- Tekan Escape atau klik luar → tutup dropdown

**Restaurant nudge:**
- Trigger: onBlur pada order_detail input
- JS check:
  ```js
  const hasLocationHint = (detail) => {
    const keywords = /\b(pak|bu|mas|mba|mbak|warung|resto|kedai|depot|kantin|gerobak)\b/i
    const dashSeparator = /[-–—]/
    return keywords.test(detail) || dashSeparator.test(detail)
  }
  ```
- Jika tidak ada lokasi → tampilkan hint muted di bawah field:
  `"Tulis juga belinya dimana, contoh: Ayam Goreng - Warung Bu Sari"`
- Styling: warna info/muted, bukan error merah

### SharedListMakView.vue — Bulk Mode

- Textarea tetap seperti sekarang (tidak ada autocomplete)
- Debounce 1s setelah user berhenti ketik → parse orders → JS check per order
- Order tanpa location hint → tampil warning chip `⚠️ Belum tulis lokasi` di bulk preview list
- User tetap bisa submit (warning, bukan block)
- Setelah submit → AI backend handle ekstraksi vendor async

### OrderListView (OB View) — Vendor Grouping

- Orders digroup by `vendor_name`
- Header per vendor group
- Orders tanpa `vendor_name` → grup "Lainnya" di bagian bawah
- Tiap order: vendor chip kecil → tap → inline text input → save → `PATCH /orders/:id/vendor`
- Autocomplete vendor dari existing vendor names di listmak yang sama

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Fireworks.ai timeout/error | Log error, order tetap tersimpan tanpa `vendor_name` |
| Suggestions endpoint gagal | Dropdown tidak muncul, user tetap bisa ketik manual |
| PATCH vendor gagal | Toast error "Gagal simpan", chip kembali ke nilai sebelumnya |
| pg_trgm extension tidak ada | Query fallback ke plain `ILIKE` (tetap works, hanya tanpa index) |

---

## Out of Scope

- Vendor grouping di SharedListMakView (share page hanya untuk input order)
- Meilisearch (ditunda, Postgres cukup untuk skala saat ini)
- On-demand batch categorization (bisa ditambah nanti jika user count bertambah)
- Vendor analytics / report per vendor

---

## Future Considerations

- Jika user count bertambah besar → migrate suggestions ke Meilisearch
- Batch re-categorization endpoint untuk backfill orders lama
- On-demand "Auto Kategorikan" button per listmak (hemat token untuk listmak besar)
