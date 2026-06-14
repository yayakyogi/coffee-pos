# API Contract — Coffee Shop POS

**Versi:** 1.0
**Tanggal:** 14 Juni 2026
**Basis:** Mengacu pada [PRD.md](PRD.md) & [SCHEMA.md](SCHEMA.md)
**Format:** REST + JSON

---

## 1. Base URL & Versioning

```
Base URL : https://{host}/api/v1
```

- Semua endpoint diawali prefix `/api/v1`.
- `Content-Type: application/json` untuk semua request berbody, **kecuali** upload foto produk yang memakai `multipart/form-data`.
- Seluruh nilai uang dikirim & diterima sebagai **integer dalam satuan sen** (1 Rupiah = 100 sen), sesuai SCHEMA. Contoh: `Rp25.000` → `2500000`.
- Timestamp memakai format **ISO 8601 / RFC 3339** UTC (mis. `2026-06-14T09:30:00Z`).
- ID berupa **UUID** string.

---

## 2. Authentication

Sistem memakai **JWT Bearer token**. Token diperoleh dari `POST /auth/login`.

Semua **protected endpoint** wajib menyertakan header:

```
Authorization: Bearer <access_token>
```

Aturan akses:
- **Public** — tanpa token (mis. login, webhook Midtrans).
- **Owner** — token dengan `role = owner`.
- **Cashier** — token dengan `role = cashier`.

Path dikelompokkan berdasarkan akses lewat **prefix segmen role**:
- `/api/v1/owner/...` — hanya `role = owner`.
- `/api/v1/cashier/...` — hanya `role = cashier`.
- `/api/v1/auth/...` dan `/api/v1/payments/midtrans/webhook` — tanpa prefix role (publik/lintas-role).

Prefix ini **mengikat secara fungsional**, bukan sekadar penamaan: otorisasi ditegakkan di level **route group**, bukan per-handler. Middleware role dipasang sekali di group dan melindungi seluruh child route — sehingga mustahil ada endpoint Owner/Cashier yang lupa diproteksi.

```
/api/v1
├── /auth                 → JWTOptional (login & logout tanpa role check)
│
├── /owner    [RequireAuth → RequireRole("owner")]
│   ├── /categories ...
│   ├── /products ...        (termasuk /products/{id}/stock-movements)
│   ├── /tables ...
│   ├── /users ...
│   ├── /promos ...
│   └── /reports ...
│
├── /cashier  [RequireAuth → RequireRole("cashier")]
│   ├── /shifts ...
│   └── /orders ...          (termasuk /orders/{id}/checkout & /payment)
│
└── /payments/midtrans/webhook   → VerifyMidtransSignature (tanpa JWT)
```

Aturannya: `RequireAuth` memvalidasi token (gagal → `401`); `RequireRole(r)` memastikan `token.role == r` (gagal → `403`). Karena dipasang di group, keduanya berjalan sebelum handler manapun di bawah prefix tersebut.

Kegagalan otorisasi:
- Token tidak ada / tidak valid / kedaluwarsa → `401`.
- Token valid tetapi role tidak berhak → `403`.

---

## 3. Standard Response Format

### 3.1 Sukses

```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": { }
}
```

List dengan paginasi menyertakan objek `meta`:

```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": [ ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 135,
    "total_pages": 7
  }
}
```

### 3.2 Error

**Error validasi** (`422`) — menyertakan field `errors` berisi map `field → pesan`:

```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": {
    "price": "price must be greater than or equal to 0",
    "items": "items must not be empty"
  }
}
```

**Error umum** (selain validasi) — cukup `success` + `message`, tanpa `errors`:

```json
{
  "success": false,
  "message": "Produk tidak ditemukan"
}
```

### 3.3 HTTP Status Code

| HTTP | Kapan dipakai |
|------|---------------|
| 200 | Sukses `GET` dan `PUT`/`PATCH` (membaca atau memperbarui). |
| 201 | Sukses `POST` yang membuat resource baru. |
| 400 | Request tidak valid (mis. JSON rusak, param salah, pelanggaran aturan bisnis seperti stok kurang / shift sudah open). |
| 401 | Tidak ada token, atau token invalid/expired. |
| 403 | Token valid tetapi role tidak punya akses. |
| 404 | Data tidak ditemukan. |
| 422 | Validasi gagal (body bentuknya benar tapi nilainya tidak lolos validasi) — selalu mengembalikan field `errors`. |
| 429 | Rate limit terlampaui. |
| 500 | Galat server tak terduga. |

> Untuk ringkas, setiap endpoint di bawah hanya menyebut error yang **khas** untuknya. Error umum `401`, `403`, `429`, dan `500` berlaku di seluruh protected endpoint.

### 3.4 Query Params Umum (list)

| Param | Default | Keterangan |
|-------|---------|------------|
| `page` | 1 | Nomor halaman. |
| `per_page` | 20 | Item per halaman (maks 100). |
| `search` | — | Pencarian teks (per-domain). |
| `sort` | `created_at` | Kolom sort. |
| `order` | `desc` | `asc` / `desc`. |

---

## 4. Endpoints

Ringkasan akses:

| Domain | Akses | Prefix path |
|--------|-------|-------------|
| Auth | Public / All | `/api/v1/auth` |
| Categories, Products, Stock, Tables, Users, Promos, Reports | Owner | `/api/v1/owner` |
| Shifts, Orders | Cashier | `/api/v1/cashier` |
| Payments — checkout & status | Cashier | `/api/v1/cashier/orders/{id}/...` |
| Payments — webhook | Public (Midtrans) | `/api/v1/payments/midtrans/webhook` |

> Catatan: Stock berada di bawah Products (`/owner/products/{id}/stock-movements`). Checkout & cek status pembayaran berada di bawah Orders (`/cashier/orders/...`). Owner mengaudit shift kasir lewat `GET /owner/shifts/{id}`, sedangkan `GET /cashier/shifts/{id}` hanya untuk kasir pemilik shift — jadi tiap prefix murni satu role.

---

## 4.1 Auth

### `POST /auth/login`
**Akses:** Public

Request:
```json
{
  "email": "cashier1@coffee.id",
  "password": "secret123"
}
```

Response `200`:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOi...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": "9f1c...e2",
      "name": "Budi",
      "email": "cashier1@coffee.id",
      "role": "cashier"
    }
  }
}
```

Error: `401` (kredensial salah), `403` (akun nonaktif / `is_active=0`).

### `GET /auth/me`
**Akses:** Owner / Cashier

Response `200`:
```json
{
  "success": true,
  "message": "Current user",
  "data": {
    "id": "9f1c...e2",
    "name": "Budi",
    "email": "cashier1@coffee.id",
    "role": "cashier",
    "is_active": true
  }
}
```

### `POST /auth/logout`
**Akses:** Owner / Cashier

Menginvalidasi token saat ini (jika memakai blacklist/refresh strategy).

Response `200`:
```json
{ "success": true, "message": "Logged out successfully", "data": null }
```

---

## 4.2 Categories (Owner)

### `GET /owner/categories`
**Akses:** Owner
**Query:** `page`, `per_page`, `search`

Response `200`:
```json
{
  "success": true,
  "message": "Categories retrieved successfully",
  "data": [
    { "id": "c-001", "name": "Coffee", "description": "Menu kopi", "created_at": "2026-06-01T08:00:00Z", "updated_at": "2026-06-01T08:00:00Z" }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 3, "total_pages": 1 }
}
```

### `GET /owner/categories/{id}`
**Akses:** Owner
Response `200`: objek kategori. Error: `404`.

### `POST /owner/categories`
**Akses:** Owner

Request:
```json
{ "name": "Pastry", "description": "Roti & kue" }
```
Response `201`: objek kategori. Error: `422` (nama sudah ada).

### `PUT /owner/categories/{id}`
**Akses:** Owner

Request:
```json
{ "name": "Pastry & Cake", "description": "Roti, kue, dessert" }
```
Response `200`: objek kategori. Error: `404`, `422`.

### `DELETE /owner/categories/{id}`
**Akses:** Owner
Soft delete. Response `200`:
```json
{ "success": true, "message": "Category deleted successfully", "data": null }
```
Error: `404`, `400` (masih dipakai produk aktif).

---

## 4.3 Products (Owner)

### `GET /owner/products`
**Akses:** Owner
**Query:** `page`, `per_page`, `search`, `category_id`, `is_active`

Response `200`:
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": "p-101",
      "category_id": "c-001",
      "category_name": "Coffee",
      "name": "Caffe Latte",
      "description": "Espresso + steamed milk",
      "price": 2800000,
      "photo_url": "/uploads/products/p-101.jpg",
      "stock": 50,
      "is_active": true,
      "created_at": "2026-06-01T08:00:00Z",
      "updated_at": "2026-06-10T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 42, "total_pages": 3 }
}
```

### `GET /owner/products/{id}`
**Akses:** Owner
Response `200`: objek produk. Error: `404`.

### `POST /owner/products`
**Akses:** Owner
**Content-Type:** `application/json` (foto via endpoint terpisah) atau `multipart/form-data`.

Request:
```json
{
  "category_id": "c-001",
  "name": "Caffe Latte",
  "description": "Espresso + steamed milk",
  "price": 2800000,
  "stock": 50,
  "is_active": true
}
```
Response `201`: objek produk. Error: `404` (category_id tak ada), `422` (price < 0).

### `PUT /owner/products/{id}`
**Akses:** Owner

Request: sama seperti create (parsial diperbolehkan). Response `200`: objek produk. Error: `404`.

### `PATCH /owner/products/{id}/status`
**Akses:** Owner
Mengubah aktif/nonaktif.

Request:
```json
{ "is_active": false }
```
Response `200`: objek produk.

### `POST /owner/products/{id}/photo`
**Akses:** Owner
**Content-Type:** `multipart/form-data`, field `photo` (file gambar).

Response `200`:
```json
{ "success": true, "message": "Photo uploaded", "data": { "photo_url": "/uploads/products/p-101.jpg" } }
```
Error: `422` (format/ukuran tidak valid), `404`.

### `DELETE /owner/products/{id}`
**Akses:** Owner
Soft delete. Response `200`:
```json
{ "success": true, "message": "Product deleted successfully", "data": null }
```
Error: `404`.

---

## 4.4 Stock (Owner)

### `GET /owner/products/{id}/stock-movements`
**Akses:** Owner
**Query:** `page`, `per_page`, `type` (`in`/`out`/`adjustment`)

Response `200`:
```json
{
  "success": true,
  "message": "Stock movements retrieved successfully",
  "data": [
    {
      "id": "sm-9001",
      "product_id": "p-101",
      "type": "out",
      "quantity": 2,
      "stock_before": 52,
      "stock_after": 50,
      "reference_type": "sale",
      "reference_id": "trx-555",
      "note": "Sale TRX-20260614-0001",
      "created_by": "u-owner",
      "created_at": "2026-06-14T09:31:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 80, "total_pages": 4 }
}
```

### `POST /owner/products/{id}/stock-movements`
**Akses:** Owner
Restock atau adjustment manual (tipe `out` otomatis dari penjualan, tidak via endpoint ini).

Request:
```json
{
  "type": "in",
  "quantity": 20,
  "note": "Restock supplier A"
}
```
- `type`: `in` (restock) atau `adjustment`. Untuk `adjustment`, `quantity` adalah selisih (boleh negatif).

Response `201`:
```json
{
  "success": true,
  "message": "Stock movement recorded",
  "data": {
    "id": "sm-9002",
    "product_id": "p-101",
    "type": "in",
    "quantity": 20,
    "stock_before": 50,
    "stock_after": 70,
    "reference_type": "restock",
    "note": "Restock supplier A",
    "created_at": "2026-06-14T11:00:00Z"
  }
}
```
Error: `404`, `400` (adjustment membuat stok negatif).

---

## 4.5 Tables (Owner)

### `GET /owner/tables`
**Akses:** Owner
**Query:** `status` (`available`/`inactive`)

Response `200`:
```json
{
  "success": true,
  "message": "Tables retrieved successfully",
  "data": [
    { "id": "t-01", "name": "Meja 1", "capacity": 4, "status": "available", "created_at": "2026-06-01T08:00:00Z", "updated_at": "2026-06-01T08:00:00Z" }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 12, "total_pages": 1 }
}
```

### `POST /owner/tables`
**Akses:** Owner

Request:
```json
{ "name": "Meja 13", "capacity": 2, "status": "available" }
```
Response `201`: objek meja. Error: `422` (nama sudah ada).

### `PUT /owner/tables/{id}`
**Akses:** Owner
Request: sama seperti create. Response `200`. Error: `404`, `422`.

### `PATCH /owner/tables/{id}/status`
**Akses:** Owner
```json
{ "status": "inactive" }
```
Response `200`: objek meja.

### `DELETE /owner/tables/{id}`
**Akses:** Owner
Response `200`. Error: `404`.

---

## 4.6 Users / Cashier Management (Owner)

### `GET /owner/users`
**Akses:** Owner
**Query:** `role`, `is_active`, `search`

Response `200`:
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    { "id": "u-201", "name": "Budi", "email": "cashier1@coffee.id", "role": "cashier", "is_active": true, "created_at": "2026-06-01T08:00:00Z" }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

### `POST /owner/users`
**Akses:** Owner
Membuat cashier baru.

Request:
```json
{
  "name": "Sari",
  "email": "cashier2@coffee.id",
  "password": "secret123",
  "role": "cashier"
}
```
Response `201`:
```json
{
  "success": true,
  "message": "User created successfully",
  "data": { "id": "u-202", "name": "Sari", "email": "cashier2@coffee.id", "role": "cashier", "is_active": true }
}
```
Error: `422` (email sudah dipakai).

### `PUT /owner/users/{id}`
**Akses:** Owner
Ubah nama/email. Response `200`. Error: `404`, `422`.

### `PATCH /owner/users/{id}/status`
**Akses:** Owner
Aktif/nonaktifkan akun.
```json
{ "is_active": false }
```
Response `200`. Error: `400` (cashier punya shift `open` — lihat PRD 3.5).

### `PATCH /owner/users/{id}/password`
**Akses:** Owner
Reset password cashier.
```json
{ "new_password": "newSecret456" }
```
Response `200`:
```json
{ "success": true, "message": "Password reset successfully", "data": null }
```

---

## 4.7 Promos (Owner)

### `GET /owner/promos`
**Akses:** Owner
**Query:** `is_active`, `type`

Response `200`:
```json
{
  "success": true,
  "message": "Promos retrieved successfully",
  "data": [
    {
      "id": "promo-01",
      "name": "Diskon Pagi 10%",
      "type": "percentage",
      "value": 10,
      "min_purchase": 5000000,
      "start_date": "2026-06-01T00:00:00Z",
      "end_date": "2026-06-30T23:59:59Z",
      "is_active": true,
      "created_at": "2026-06-01T08:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 4, "total_pages": 1 }
}
```

### `POST /owner/promos`
**Akses:** Owner

Request (persentase):
```json
{
  "name": "Diskon Pagi 10%",
  "type": "percentage",
  "value": 10,
  "min_purchase": 5000000,
  "start_date": "2026-06-01T00:00:00Z",
  "end_date": "2026-06-30T23:59:59Z",
  "is_active": true
}
```
Request (nominal):
```json
{
  "name": "Potongan 15rb",
  "type": "fixed",
  "value": 1500000,
  "min_purchase": 0,
  "start_date": "2026-06-01T00:00:00Z",
  "end_date": "2026-06-30T23:59:59Z",
  "is_active": true
}
```
Response `201`: objek promo.
Error: `422` (percentage di luar 0–100, value negatif, `end_date` < `start_date`).

### `PUT /owner/promos/{id}`
**Akses:** Owner
Request: sama seperti create. Response `200`. Error: `404`.

### `PATCH /owner/promos/{id}/status`
**Akses:** Owner
```json
{ "is_active": false }
```
Response `200`.

### `DELETE /owner/promos/{id}`
**Akses:** Owner
Soft delete. Response `200`. Error: `404`.

---

## 4.8 Reports (Owner)

Semua laporan hanya menghitung transaksi berstatus `paid` (PRD Business Rule #7).

### `GET /owner/reports/revenue`
**Akses:** Owner
**Query:** `period` (`daily`/`weekly`/`monthly`), `start_date`, `end_date`

Response `200`:
```json
{
  "success": true,
  "message": "Revenue report retrieved successfully",
  "data": {
    "period": "daily",
    "start_date": "2026-06-01",
    "end_date": "2026-06-14",
    "total_revenue": 152500000,
    "total_transactions": 320,
    "series": [
      { "date": "2026-06-14", "revenue": 12500000, "transactions": 28 }
    ]
  }
}
```

### `GET /owner/reports/best-sellers`
**Akses:** Owner
**Query:** `start_date`, `end_date`, `limit` (default 10)

Response `200`:
```json
{
  "success": true,
  "message": "Best sellers retrieved successfully",
  "data": [
    { "product_id": "p-101", "product_name": "Caffe Latte", "quantity_sold": 240, "revenue": 67200000 }
  ]
}
```

### `GET /owner/reports/cashier-performance`
**Akses:** Owner
**Query:** `start_date`, `end_date`

Response `200`:
```json
{
  "success": true,
  "message": "Cashier performance retrieved successfully",
  "data": [
    { "cashier_id": "u-201", "cashier_name": "Budi", "total_transactions": 120, "total_revenue": 56000000 }
  ]
}
```

### `GET /owner/reports/transactions/export`
**Akses:** Owner
**Query:** `start_date`, `end_date`
Export CSV (PRD 3.8).

Response `200`:
- Header: `Content-Type: text/csv; charset=utf-8`, `Content-Disposition: attachment; filename="transactions_2026-06-01_2026-06-14.csv"`
- Body: file CSV (kolom: `transaction_number,date,cashier,table,subtotal,discount,total,status`).

Error: `422` (rentang tanggal tidak valid).

### `GET /owner/shifts/{id}`
**Akses:** Owner
Audit shift kasir mana pun (read-only). Owner tidak dibatasi kepemilikan shift.

Response `200`:
```json
{
  "success": true,
  "message": "Shift retrieved successfully",
  "data": {
    "id": "shift-77",
    "cashier_id": "u-201",
    "cashier_name": "Budi",
    "opening_balance": 50000000,
    "total_sales": 8500000,
    "total_transactions": 14,
    "expected_balance": 58500000,
    "closing_balance": 58400000,
    "difference": -100000,
    "status": "closed",
    "opened_at": "2026-06-14T07:00:00Z",
    "closed_at": "2026-06-14T15:00:00Z"
  }
}
```
Error: `404` (shift tidak ditemukan).

---

## 4.9 Shifts (Cashier)

### `POST /cashier/shifts/open`
**Akses:** Cashier
Buka shift dengan modal kas.

Request:
```json
{ "opening_balance": 50000000 }
```
Response `201`:
```json
{
  "success": true,
  "message": "Shift opened successfully",
  "data": {
    "id": "shift-77",
    "cashier_id": "u-201",
    "opening_balance": 50000000,
    "status": "open",
    "opened_at": "2026-06-14T07:00:00Z"
  }
}
```
Error: `422` (sudah ada shift `open` — PRD Business Rule #2/#14).

### `GET /cashier/shifts/current`
**Akses:** Cashier
Shift `open` milik kasir saat ini.

Response `200`:
```json
{
  "success": true,
  "message": "Current shift",
  "data": {
    "id": "shift-77",
    "opening_balance": 50000000,
    "total_sales": 8500000,
    "total_transactions": 14,
    "status": "open",
    "opened_at": "2026-06-14T07:00:00Z"
  }
}
```
Error: `404` (tidak ada shift terbuka).

### `POST /cashier/shifts/{id}/close`
**Akses:** Cashier (pemilik shift)
Tutup shift dengan kas akhir aktual; sistem menghitung rekap & selisih.

Request:
```json
{ "closing_balance": 58400000 }
```
Response `200`:
```json
{
  "success": true,
  "message": "Shift closed successfully",
  "data": {
    "id": "shift-77",
    "opening_balance": 50000000,
    "total_sales": 8500000,
    "total_transactions": 14,
    "expected_balance": 58500000,
    "closing_balance": 58400000,
    "difference": -100000,
    "status": "closed",
    "opened_at": "2026-06-14T07:00:00Z",
    "closed_at": "2026-06-14T15:00:00Z"
  }
}
```
Error: `404`, `403` (bukan pemilik shift), `400` (shift sudah `closed`).

### `GET /cashier/shifts/{id}`
**Akses:** Cashier (hanya pemilik shift)
Detail & rekap shift milik kasir sendiri. Response `200`: objek shift. Error: `404`, `403` (bukan pemilik shift).

> Owner mengakses shift kasir mana pun lewat endpoint terpisah `GET /owner/shifts/{id}` (di akhir §4.8 Reports).

---

## 4.10 Orders (Cashier)

Order dibuat dalam status `pending`. Item tidak dapat diubah setelah checkout (PRD Business Rule #16). Stok berkurang hanya setelah webhook Midtrans confirmed (Business Rule #13).

### `GET /cashier/orders`
**Akses:** Cashier
Riwayat transaksi shift hari ini milik kasir sendiri (PRD 4.5).
**Query:** `status`, `date` (default hari ini)

Response `200`:
```json
{
  "success": true,
  "message": "Orders retrieved successfully",
  "data": [
    {
      "id": "trx-555",
      "transaction_number": "TRX-20260614-0001",
      "table_id": "t-01",
      "table_name": "Meja 1",
      "subtotal": 5600000,
      "discount_amount": 560000,
      "total": 5040000,
      "status": "paid",
      "created_at": "2026-06-14T09:30:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 14, "total_pages": 1 }
}
```

### `POST /cashier/orders`
**Akses:** Cashier
Buat order baru. Wajib ada shift `open`.

Request:
```json
{
  "table_id": "t-01",
  "items": [
    { "product_id": "p-101", "quantity": 2 }
  ]
}
```
Response `201`:
```json
{
  "success": true,
  "message": "Order created successfully",
  "data": {
    "id": "trx-555",
    "transaction_number": "TRX-20260614-0001",
    "shift_id": "shift-77",
    "table_id": "t-01",
    "promo_id": null,
    "items": [
      { "id": "ti-1", "product_id": "p-101", "product_name": "Caffe Latte", "price": 2800000, "quantity": 2, "subtotal": 5600000 }
    ],
    "subtotal": 5600000,
    "discount_amount": 0,
    "total": 5600000,
    "status": "pending",
    "created_at": "2026-06-14T09:30:00Z"
  }
}
```
Error:
- `400` — tidak ada shift terbuka (Business Rule #14).
- `400` — stok tidak cukup / produk nonaktif.
- `404` — produk/meja tidak ditemukan.
- `422` — items kosong / quantity < 1.

### `GET /cashier/orders/{id}`
**Akses:** Cashier (pemilik)
Detail order (header + items + promo). Response `200`. Error: `404`.

### `POST /cashier/orders/{id}/apply-promo`
**Akses:** Cashier
Terapkan satu promo aktif (maks. satu — Business Rule #15). Hanya untuk order `pending`.

Request:
```json
{ "promo_id": "promo-01" }
```
Response `200`:
```json
{
  "success": true,
  "message": "Promo applied successfully",
  "data": {
    "id": "trx-555",
    "promo_id": "promo-01",
    "subtotal": 5600000,
    "discount_amount": 560000,
    "total": 5040000,
    "status": "pending"
  }
}
```
Error:
- `422` — subtotal di bawah `min_purchase`.
- `400` — promo tidak aktif / di luar periode, atau order bukan `pending`.
- `404` — promo/order tidak ditemukan.

### `DELETE /cashier/orders/{id}/promo`
**Akses:** Cashier
Lepas promo dari order `pending`. Response `200`: order dengan `discount_amount=0`.

### `POST /cashier/orders/{id}/cancel`
**Akses:** Cashier
Batalkan order `pending` (belum dibayar). Response `200`:
```json
{ "success": true, "message": "Order cancelled", "data": { "id": "trx-555", "status": "cancelled" } }
```
Error: `400` (order sudah `paid`).

---

## 4.11 Payments

### `POST /cashier/orders/{id}/checkout`
**Akses:** Cashier
Membuat transaksi pembayaran Midtrans Snap. Setelah checkout, item order terkunci (Business Rule #16).

Request: _(tanpa body)_

Response `201`:
```json
{
  "success": true,
  "message": "Checkout initiated",
  "data": {
    "transaction_id": "trx-555",
    "payment": {
      "id": "pay-900",
      "midtrans_order_id": "TRX-20260614-0001",
      "snap_token": "66e4fa55-fdac-4ef9-91b5-733b97d1b862",
      "snap_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/66e4fa55...",
      "gross_amount": 5040000,
      "status": "pending"
    }
  }
}
```
Error:
- `400` — order kosong / sudah `paid` / sudah di-checkout.
- `404` — order tidak ditemukan.

### `GET /cashier/orders/{id}/payment`
**Akses:** Cashier
Status pembayaran terkini suatu order.

Response `200`:
```json
{
  "success": true,
  "message": "Payment status",
  "data": {
    "id": "pay-900",
    "transaction_id": "trx-555",
    "status": "settlement",
    "payment_type": "qris",
    "gross_amount": 5040000,
    "paid_at": "2026-06-14T09:33:00Z"
  }
}
```
Error: `404`.

### `POST /payments/midtrans/webhook`
**Akses:** Public (dipanggil server Midtrans)
Notifikasi pembayaran. Sistem **memverifikasi signature** sebelum memproses (Business Rule #8). Jika status `settlement`/`capture`, sistem menandai transaksi `paid`, mengurangi stok, dan mencatat `stock_movements` bertipe `out` (Business Rule #5 & #13).

Request (payload Midtrans, ringkas):
```json
{
  "order_id": "TRX-20260614-0001",
  "transaction_status": "settlement",
  "payment_type": "qris",
  "gross_amount": "50400.00",
  "fraud_status": "accept",
  "signature_key": "a1b2c3...",
  "status_code": "200"
}
```
Response `200`:
```json
{ "success": true, "message": "Notification processed", "data": null }
```
Error:
- `401` — signature tidak valid.
- `404` — `order_id` tidak dikenal.
- Respons tetap `200` untuk notifikasi duplikat (idempotent) agar Midtrans berhenti retry.

---

## 5. Catatan Implementasi

- **Idempotensi webhook:** notifikasi Midtrans bisa terkirim berkali-kali; pemrosesan harus idempotent (cek status transaksi sebelum mengubah & sebelum mengurangi stok).
- **Atomicity:** pembuatan order + item, dan transisi `paid` (update transaksi + pengurangan stok + insert stock_movement), dijalankan dalam satu transaksi DB.
- **Ownership check:** endpoint Cashier hanya boleh mengakses order/shift miliknya sendiri; pelanggaran → `403`.
- **Locking:** "satu shift open per kasir" dan pengecekan stok memakai `SELECT ... FOR UPDATE` (lihat SCHEMA §4.9).
- **Uang dalam sen:** seluruh field uang konsisten integer-sen; klien bertanggung jawab memformat tampilan Rupiah.
