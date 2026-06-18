# ListMak API Documentation

> Dokumentasi endpoint API untuk fitur ListMak.  
> Base URL: `http://localhost:9001/api/v1` (Development)

---

## 📋 Daftar Endpoint

### 1. Authentication

| Method | Endpoint                | Deskripsi                  | Auth Required |
| ------ | ----------------------- | -------------------------- | ------------- |
| GET    | `/auth/google/login`    | Redirect ke Google Login   | No            |
| GET    | `/auth/google/callback` | Callback dari Google       | No            |
| GET    | `/auth/user`            | Get current logged in user | Yes           |
| GET    | `/auth/logout`          | Logout (Clear cookie)      | Yes           |

### 2. ListMak (Daftar Pesanan Harian)

| Method | Endpoint               | Deskripsi                              | Auth Required |
| ------ | ---------------------- | -------------------------------------- | ------------- |
| GET    | `/listmaks`            | Ambil semua listmak (filter supported) | Yes           |
| POST   | `/listmaks`            | Buat listmak baru                      | Yes           |
| GET    | `/listmaks/:id`        | Ambil detail listmak by ID             | Yes           |
| PUT    | `/listmaks/:id`        | Update listmak                         | Yes           |
| DELETE | `/listmaks/:id`        | Hapus listmak                          | Yes           |
| GET    | `/listmaks/date/:date` | Ambil listmak berdasarkan tanggal      | Yes           |

### 3. Orders (Pesanan)

| Method | Endpoint                    | Deskripsi                         | Auth Required |
| ------ | --------------------------- | --------------------------------- | ------------- |
| GET    | `/listmaks/:id/orders`      | Ambil semua pesanan dalam listmak | Yes           |
| POST   | `/listmaks/:id/orders`      | Tambah pesanan baru ke listmak    | Yes           |
| POST   | `/listmaks/:id/orders/bulk` | Tambah pesanan bulk (parse WA)    | Yes           |
| PUT    | `/orders/:id`               | Update pesanan                    | Yes           |
| PATCH  | `/orders/:id/paid`          | Update status bayar pesanan       | Yes           |
| DELETE | `/orders/:id`               | Hapus pesanan                     | Yes           |

### 4. Share Links (Link Input Pesanan)

| Method | Endpoint                       | Deskripsi                             | Auth Required |
| ------ | ------------------------------ | ------------------------------------- | ------------- |
| POST   | `/share-links`                 | Buat share link baru                  | Yes           |
| DELETE | `/share-links/:id`             | Hapus/nonaktifkan share link          | Yes           |
| GET    | `/share-links/:shareId`        | Ambil data share link (Public Access) | **NO**        |
| POST   | `/share-links/:shareId/orders` | Submit pesanan via share link         | **NO**        |

### 5. View Shares (Link View-Only)

| Method | Endpoint               | Deskripsi                               | Auth Required |
| ------ | ---------------------- | --------------------------------------- | ------------- |
| POST   | `/view-shares`         | Buat view share link                    | Yes           |
| GET    | `/view-shares/:viewId` | Ambil data listmak view (Public Access) | **NO**        |

---

## 📝 Detail API Endpoint

---

### **AUTHENTICATION**

#### `GET /auth/user`

Mendapatkan data user yang sedang login (berdasarkan cookie).

**Response:**

```json
{
  "success": true,
  "message": "User authenticated",
  "data": {
    "id": 1,
    "google_id": "123456789",
    "email": "user@example.com",
    "name": "User Name",
    "avatar": "https://lh3.googleusercontent.com/...",
    "role": "user",
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  }
}
```

---

### **LISTMAKS**

#### `GET /listmaks`

Ambil daftar listmak dengan pagination dan filtering.

**Query Params:**

- `page`: (int) Halaman ke berapa (default: 1)
- `limit`: (int) Jumlah item per halaman (default: 10)
- `status`: (string) Filter status ('active', 'completed')
- `start_date`: (YYYY-MM-DD) Filter tanggal mulai
- `end_date`: (YYYY-MM-DD) Filter tanggal akhir

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "ListMak Sabtu, 4 Januari",
      "date": "2026-01-04T00:00:00Z",
      "total_orders": 15,
      "total_amount": 225000,
      "paid_amount": 150000,
      "status": "active",
      "created_at": "2026-01-04T08:00:00Z",
      "orders": [
        {
          "id": 1,
          "name": "Icha",
          "order_detail": "Nasi Ayam Madura PAHA",
          "price": 15000,
          "qty": 1,
          "total_price": 15000,
          "is_paid": true
        }
      ],
      "share_links": [
        {
          "id": 1,
          "share_id": "XyZ123AB",
          "title": "Link Order Siang",
          "expires_at": "2026-01-05T12:00:00Z",
          "is_active": true
        }
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 50
  }
}
```

#### `POST /listmaks`

Buat listmak baru.

**Request Body:**

```json
{
  "title": "ListMak Senin, 6 Januari",
  "date": "2026-01-06T00:00:00Z"
}
```

#### `GET /listmaks/:id`

Ambil detail listmak beserta semua orders di dalamnya (via Preload).

#### `GET /listmaks/date/:date`

Ambil semua listmak berdasarkan tanggal (Format: YYYY-MM-DD).

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "ListMak...",
      "date": "2026-01-04T00:00:00Z",
      "orders": [...],
      "share_links": [...]
    }
  ]
}
```

---

### **ORDERS**

#### `GET /listmaks/:id/orders`

Ambil semua order untuk listmak tertentu.

**Query Params:**

- `is_paid`: (true/false) Filter status bayar
- `search`: (string) Cari nama atau detail pesanan

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "listmak_id": 1,
      "name": "Icha",
      "order_detail": "Nasi Ayam Madura PAHA",
      "price": 15000,
      "qty": 1,
      "total_price": 15000,
      "is_paid": true,
      "added_via": "parse"
    }
  ],
  "summary": {
    "total_orders": 1,
    "paid_count": 1,
    "total_amount": 15000,
    "paid_amount": 15000
  }
}
```

#### `POST /listmaks/:id/orders`

Tambah satu pesanan.

**Request Body:**

```json
{
  "name": "Budi",
  "order_detail": "Es Teh Manis",
  "price": 5000,
  "qty": 2,
  "added_via": "manual" // optional
}
```

#### `POST /listmaks/:id/orders/bulk`

Tambah banyak pesanan sekaligus.

**Request Body:**

```json
{
  "orders": [
    {
      "name": "Ali",
      "order_detail": "Ayam Geprek",
      "price": 15000,
      "qty": 1
    },
    {
      "name": "Siti",
      "order_detail": "Mie Goreng",
      "price": 12000,
      "qty": 1
    }
  ],
  "added_via": "parse"
}
```

#### `PUT /orders/:id`

Update data pesanan.

**Request Body:**

```json
{
  "name": "Budi Update",
  "order_detail": "Es Teh Tawar",
  "price": 3000,
  "qty": 1
}
```

#### `PATCH /orders/:id/paid`

Update status pembayaran.

**Request Body:**

```json
{
  "is_paid": true
}
```

#### `DELETE /orders/:id`

Hapus pesanan.

---

### **SHARE LINKS (Public)**

#### `POST /share-links`

Buat link untuk share ke grup (agar user lain bisa input).

**Request Body:**

```json
{
  "listmak_id": 1,
  "title": "Link Order Siang",
  "expires_at": "2026-01-05T12:00:00Z"
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "share_id": "XyZ123AB",
    "listmak_id": 1,
    "title": "Link Order Siang",
    "expires_at": "..."
  }
}
```

#### `GET /share-links/:shareId` (PUBLIC)

Endpoint untuk halaman publik input pesanan. Mengecek apakah link valid/expired.

**Response Success:**

```json
{
  "success": true,
  "data": {
    "share_id": "XyZ123AB",
    "title": "Link Order Siang",
    "expires_at": "...",
    "is_expired": false,
    "listmak_date": "2026-01-04"
  }
}
```

**Response Expired (HTTP 410):**

```json
{
  "success": false,
  "error": "EXPIRED",
  "message": "Waktu input pesanan telah berakhir",
  "data": { "expires_at": "..." }
}
```

#### `POST /share-links/:shareId/orders` (PUBLIC)

Submit pesanan dari halaman publik. Bisa single atau bulk.

**Request Body:**

```json
{
  "name": "Tamu",
  "order_detail": "Nasi Kuning"
}
```

_Atau Bulk:_

```json
{
  "orders": [
    { "name": "A", "order_detail": "Menu A" },
    { "name": "B", "order_detail": "Menu B" }
  ]
}
```

---

### **VIEW SHARES (Public)**

#### `POST /view-shares`

Buat link view-only (snapshot/readonly).

**Request Body:**

```json
{
  "listmak_id": 1,
  "title": "Rekapan Makan Siang"
}
```

#### `GET /view-shares/:viewId` (PUBLIC)

Lihat data listmak (snapshot).

**Response:**

```json
{
  "success": true,
  "data": {
    "view_id": "vIeW123",
    "title": "Rekapan Makan Siang",
    "snapshot": {
      "id": 1,
      "title": "ListMak...",
      "orders": [...] // Data orders saat snapshot dibuat
    }
  }
}
```
