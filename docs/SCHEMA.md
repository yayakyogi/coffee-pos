# SCHEMA — Coffee Shop POS (MySQL)

**Versi:** 1.0
**Tanggal:** 14 Juni 2026
**Basis:** Mengacu pada [PRD.md](PRD.md)
**Engine:** InnoDB, charset `utf8mb4`, collation `utf8mb4_unicode_ci`

Konvensi umum yang berlaku untuk **semua tabel**:
- **Primary key**: `id VARCHAR(36)` berisi UUID (bukan AUTO_INCREMENT).
- **Uang**: seluruh nilai moneter `BIGINT` dalam satuan **sen** (1 Rupiah = 100 sen).
- **Timestamp**: setiap tabel memiliki `created_at` dan `updated_at` (`TIMESTAMP`).
- **Soft delete**: tabel yang membutuhkannya memiliki `deleted_at` (`TIMESTAMP NULL`).
- **Status terbatas**: menggunakan `ENUM`.

---

## 1. Daftar Tabel

| # | Tabel | Deskripsi |
|---|-------|-----------|
| 1 | `users` | Akun pengguna sistem (Owner & Cashier) beserta kredensial dan status aktif. |
| 2 | `categories` | Kategori produk (mis. Coffee, Non-Coffee, Pastry). |
| 3 | `products` | Master produk yang dijual, dengan soft delete, foto, dan status aktif. |
| 4 | `stock_movements` | Riwayat pergerakan stok (in/out/adjustment) per produk. |
| 5 | `tables` | Master meja untuk transaksi dine-in. |
| 6 | `promos` | Promo & diskon (persentase / nominal) dengan periode berlaku. |
| 7 | `shifts` | Shift kasir: modal kas awal, rekap penutupan, dan selisih kas. |
| 8 | `transactions` | Header transaksi/order: meja, kasir, shift, promo, total, status. |
| 9 | `transaction_items` | Baris item per transaksi (snapshot nama & harga produk). |
| 10 | `payments` | Data pembayaran Midtrans Snap & status webhook per transaksi. |

> Catatan: tidak ada tabel khusus untuk laporan — dashboard & export CSV dihitung (aggregate query) dari `transactions`, `transaction_items`, dan `shifts`.

---

## 2. Definisi Tabel

### 2.1 `users`
Menyimpan akun Owner dan Cashier. Login ditolak bila `is_active = 0` atau `deleted_at` terisi.

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| name | VARCHAR(100) | NOT NULL | Nama lengkap pengguna. |
| email | VARCHAR(50) | UNIQUE, NOT NULL | Dipakai untuk login. |
| password_hash | VARCHAR(255) | NOT NULL | Hash bcrypt — tidak pernah plaintext. |
| role | ENUM('owner','cashier') | NOT NULL | Menentukan hak akses. |
| is_active | TINYINT(1) | NOT NULL, DEFAULT 1 | 0 = nonaktif (tidak bisa login). |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |
| deleted_at | TIMESTAMP | NULL | Soft delete akun. |

### 2.2 `categories`
Kategori untuk mengelompokkan produk. Setiap produk wajib punya satu kategori.

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| name | VARCHAR(100) | UNIQUE, NOT NULL | Nama kategori. |
| description | VARCHAR(255) | NULL | Keterangan opsional. |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |
| deleted_at | TIMESTAMP | NULL | Soft delete agar produk historis tetap punya nama kategori. |

### 2.3 `products`
Master produk. Soft delete menjaga integritas transaksi historis (PRD Business Rule #3).

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| category_id | VARCHAR(36) | FK → categories.id, NOT NULL | Kategori produk. |
| name | VARCHAR(150) | NOT NULL | Nama produk. |
| description | TEXT | NULL | Deskripsi produk. |
| price | BIGINT | NOT NULL, DEFAULT 0 | Harga jual dalam **sen**. CHECK ≥ 0. |
| photo_url | VARCHAR(255) | NULL | Path/URL foto produk. |
| stock | INT | NOT NULL, DEFAULT 0 | Stok saat ini (cache); sumber kebenaran = stock_movements. |
| is_active | TINYINT(1) | NOT NULL, DEFAULT 1 | 0 = nonaktif, tidak bisa dipilih kasir. |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |
| deleted_at | TIMESTAMP | NULL | Soft delete produk. |

### 2.4 `stock_movements`
Audit trail setiap perubahan stok. Insert-only (tidak pernah di-update/hapus).

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| product_id | VARCHAR(36) | FK → products.id, NOT NULL | Produk terkait. |
| type | ENUM('in','out','adjustment') | NOT NULL | Jenis pergerakan. |
| quantity | INT | NOT NULL | Jumlah unit yang bergerak (selalu positif). |
| stock_before | INT | NOT NULL | Saldo stok sebelum pergerakan. |
| stock_after | INT | NOT NULL | Saldo stok setelah pergerakan. |
| reference_type | ENUM('restock','sale','adjustment') | NOT NULL | Asal pergerakan. |
| reference_id | VARCHAR(36) | NULL | ID transaksi (bila `sale`); NULL untuk restock/adjustment manual. |
| note | VARCHAR(255) | NULL | Keterangan/alasan. |
| created_by | VARCHAR(36) | FK → users.id, NOT NULL | User yang melakukan. |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |

### 2.5 `tables`
Master meja dine-in. (Nama tabel di-quote dengan backtick karena `tables` mendekati kata kunci MySQL.)

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| name | VARCHAR(50) | UNIQUE, NOT NULL | Nomor/nama meja. |
| capacity | INT | NULL | Kapasitas kursi (opsional). |
| status | ENUM('available','inactive') | NOT NULL, DEFAULT 'available' | Meja `inactive` tidak bisa dipilih. |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |

### 2.6 `promos`
Promo/diskon. Lihat keputusan desain #5 untuk semantik kolom `value`.

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| name | VARCHAR(150) | NOT NULL | Nama promo. |
| type | ENUM('percentage','fixed') | NOT NULL | Tipe diskon. |
| value | BIGINT | NOT NULL | Bila `percentage`: 0–100 (persen). Bila `fixed`: nominal dalam **sen**. |
| min_purchase | BIGINT | NOT NULL, DEFAULT 0 | Minimum subtotal (sen) agar promo berlaku. |
| start_date | DATETIME | NOT NULL | Awal periode berlaku. |
| end_date | DATETIME | NOT NULL | Akhir periode berlaku. |
| is_active | TINYINT(1) | NOT NULL, DEFAULT 1 | Toggle aktif/nonaktif manual. |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |
| deleted_at | TIMESTAMP | NULL | Soft delete agar transaksi historis tetap mereferensi promo. |

### 2.7 `shifts`
Shift kasir. Satu kasir maksimal satu shift `open` (PRD Business Rule #2/#14).

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| cashier_id | VARCHAR(36) | FK → users.id, NOT NULL | Kasir pemilik shift. |
| opening_balance | BIGINT | NOT NULL, DEFAULT 0 | Modal kas awal (sen). |
| expected_balance | BIGINT | NULL | Ekspektasi kas akhir saat tutup (sen). |
| closing_balance | BIGINT | NULL | Kas akhir aktual yang dihitung kasir (sen). |
| difference | BIGINT | NULL | Selisih over/short = closing − expected (sen, bisa negatif). |
| total_sales | BIGINT | NULL | Total penjualan paid selama shift (sen). |
| total_transactions | INT | NOT NULL, DEFAULT 0 | Jumlah transaksi paid. |
| status | ENUM('open','closed') | NOT NULL, DEFAULT 'open' | |
| opened_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Waktu buka shift. |
| closed_at | TIMESTAMP | NULL | Waktu tutup shift. |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |

### 2.8 `transactions`
Header transaksi/order. Status pembayaran final ditentukan oleh webhook Midtrans (PRD Business Rule #13).

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| transaction_number | VARCHAR(30) | UNIQUE, NOT NULL | Nomor transaksi human-readable (mis. TRX-20260614-0001). |
| shift_id | VARCHAR(36) | FK → shifts.id, NOT NULL | Shift saat transaksi dibuat. |
| cashier_id | VARCHAR(36) | FK → users.id, NOT NULL | Kasir pembuat. |
| table_id | VARCHAR(36) | FK → tables.id, NULL | Meja dine-in (boleh kosong). |
| promo_id | VARCHAR(36) | FK → promos.id, NULL | Promo yang dipakai (maks. satu — Business Rule #15). |
| subtotal | BIGINT | NOT NULL, DEFAULT 0 | Total sebelum diskon (sen). |
| discount_amount | BIGINT | NOT NULL, DEFAULT 0 | Besaran potongan promo (sen). |
| total | BIGINT | NOT NULL, DEFAULT 0 | subtotal − discount_amount (sen, ≥ 0). |
| status | ENUM('pending','paid','failed','cancelled') | NOT NULL, DEFAULT 'pending' | Status order; jadi `paid` hanya setelah webhook confirmed. |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |

### 2.9 `transaction_items`
Baris item. Menyimpan **snapshot** nama & harga agar laporan historis tidak berubah saat master produk diedit.

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| transaction_id | VARCHAR(36) | FK → transactions.id, NOT NULL | Transaksi induk. |
| product_id | VARCHAR(36) | FK → products.id, NOT NULL | Referensi produk. |
| product_name | VARCHAR(150) | NOT NULL | Snapshot nama produk saat transaksi. |
| price | BIGINT | NOT NULL | Snapshot harga satuan (sen) saat transaksi. |
| quantity | INT | NOT NULL | Jumlah, ≥ 1. |
| subtotal | BIGINT | NOT NULL | price × quantity (sen). |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |

### 2.10 `payments`
Satu baris pembayaran per transaksi. Menyimpan token Snap dan status dari notifikasi Midtrans.

| nama_kolom | tipe_data | constraint | keterangan |
|------------|-----------|------------|------------|
| id | VARCHAR(36) | PK, NOT NULL | UUID. |
| transaction_id | VARCHAR(36) | FK → transactions.id, UNIQUE, NOT NULL | Relasi 1:1 ke transaksi. |
| midtrans_order_id | VARCHAR(50) | UNIQUE, NOT NULL | `order_id` yang dikirim ke Midtrans. |
| snap_token | VARCHAR(255) | NULL | Token Snap dari Midtrans. |
| snap_url | VARCHAR(255) | NULL | Redirect URL Snap. |
| gross_amount | BIGINT | NOT NULL | Nilai yang ditagih (sen). |
| payment_type | VARCHAR(50) | NULL | Metode (gopay, qris, dll) dari Midtrans. |
| status | ENUM('pending','settlement','capture','deny','cancel','expire','failure') | NOT NULL, DEFAULT 'pending' | Status mentah dari Midtrans. |
| raw_notification | JSON | NULL | Payload webhook terakhir untuk audit/rekonsiliasi. |
| paid_at | TIMESTAMP | NULL | Waktu pembayaran terkonfirmasi. |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | |

---

## 3. Relasi Antar Tabel (Foreign Keys)

| Child | Kolom | → Parent | Kolom | On Delete | Keterangan |
|-------|-------|----------|-------|-----------|------------|
| products | category_id | categories | id | RESTRICT | Kategori tak bisa dihapus selama dipakai. |
| stock_movements | product_id | products | id | RESTRICT | Jaga audit trail tetap valid. |
| stock_movements | created_by | users | id | RESTRICT | Akuntabilitas pelaku. |
| shifts | cashier_id | users | id | RESTRICT | |
| transactions | shift_id | shifts | id | RESTRICT | |
| transactions | cashier_id | users | id | RESTRICT | |
| transactions | table_id | tables | id | SET NULL | Meja boleh dihapus tanpa merusak transaksi lama. |
| transactions | promo_id | promos | id | SET NULL | Promo boleh dihapus tanpa merusak transaksi lama. |
| transaction_items | transaction_id | transactions | id | CASCADE | Item ikut terhapus bila transaksi (pending) dibatalkan/dihapus. |
| transaction_items | product_id | products | id | RESTRICT | |
| payments | transaction_id | transactions | id | CASCADE | Pembayaran 1:1 mengikuti transaksi. |

### Daftar relasi (format ringkas)

```
[products.category_id]            → [categories.id]    | Setiap produk milik satu kategori (RESTRICT)
[stock_movements.product_id]      → [products.id]      | Pergerakan stok merujuk produk (RESTRICT)
[stock_movements.created_by]      → [users.id]         | Pelaku pergerakan stok (RESTRICT)
[shifts.cashier_id]               → [users.id]         | Pemilik shift (RESTRICT)
[transactions.shift_id]           → [shifts.id]        | Transaksi terikat satu shift (RESTRICT)
[transactions.cashier_id]         → [users.id]         | Kasir pembuat transaksi (RESTRICT)
[transactions.table_id]           → [tables.id]        | Meja dine-in, nullable (SET NULL)
[transactions.promo_id]           → [promos.id]        | Promo dipakai, maks. satu, nullable (SET NULL)
[transaction_items.transaction_id]→ [transactions.id]  | Item milik satu transaksi (CASCADE)
[transaction_items.product_id]    → [products.id]      | Item merujuk produk (RESTRICT)
[payments.transaction_id]         → [transactions.id]  | Pembayaran 1:1 dengan transaksi (CASCADE)
```

### Diagram relasi (ringkas)

```
categories 1───* products 1───* stock_movements
                       │              ▲
                       │              │ created_by
users 1──* shifts 1──* transactions *─┘
  │              │          │ 1
  │ created_by   │          ├──* transaction_items *── products
  └──────────────┘          ├──1 payments
                            *─1 tables (nullable)
                            *─1 promos (nullable)
```

### Index yang disarankan
- `users(email)` UNIQUE; `users(role)`.
- `products(category_id)`, `products(is_active, deleted_at)`.
- `stock_movements(product_id, created_at)`, `stock_movements(reference_type, reference_id)`.
- `transactions(shift_id)`, `transactions(cashier_id, created_at)`, `transactions(status, created_at)` — untuk laporan & filter periode.
- `transaction_items(transaction_id)`, `transaction_items(product_id)` — laporan produk terlaris.
- Partial uniqueness "satu shift open per kasir" tidak didukung native MySQL; ditegakkan di layer aplikasi (transaksi + `SELECT ... FOR UPDATE`).

---

## 4. Keputusan Desain Penting

**1. UUID `VARCHAR(36)` sebagai PK, bukan AUTO_INCREMENT.**
ID tidak dapat ditebak/diurut (lebih aman untuk endpoint publik), dapat di-generate di sisi aplikasi sebelum insert (memudahkan operasi multi-tabel dalam satu transaksi, mis. transaksi + item + payment sekaligus), dan menghindari konflik bila kelak ada sinkronisasi/multi-instance. Trade-off: ukuran index lebih besar dan ordering tidak natural — dimitigasi dengan index pada `created_at`. (Bila butuh performa lebih, UUID dapat disimpan sebagai `BINARY(16)`, namun `VARCHAR(36)` dipilih demi keterbacaan sesuai ketentuan.)

**2. Uang sebagai `BIGINT` dalam satuan sen.**
Menghindari galat pembulatan floating point (`FLOAT`/`DOUBLE`) pada perhitungan subtotal, diskon, dan total — selaras dengan PRD Business Rule #12. Satuan sen (1/100 Rupiah) memberi ruang presisi untuk perhitungan persentase diskon sebelum dibulatkan, tanpa kehilangan nilai. `BIGINT` (hingga ~9,2 × 10¹⁸) jauh melebihi kebutuhan nominal coffee shop. Dibanding `DECIMAL`, integer lebih cepat dan bebas ambiguitas pembulatan.

**3. `created_at` / `updated_at` di semua tabel; `deleted_at` hanya pada tabel master.**
Audit dasar diberikan untuk semua entitas. Soft delete (`deleted_at`) diterapkan pada master data yang direferensikan transaksi historis — `users`, `categories`, `products`, `promos` — sehingga laporan & transaksi lama tetap utuh meski data master "dihapus" (PRD Business Rule #3). Tabel transaksional (`transactions`, `transaction_items`, `payments`, `stock_movements`, `shifts`) **tidak** memakai soft delete: transaksi yang batal cukup berstatus `cancelled/failed`, dan `stock_movements` bersifat insert-only sebagai audit trail permanen.

**4. `ENUM` untuk kolom berstatus terbatas.**
`role`, `type` (stock & promo), `status` (table, shift, transaction, payment), dan `reference_type` memakai `ENUM` agar nilai tervalidasi di level DB, hemat storage, dan self-documenting. Konsekuensi: menambah nilai baru perlu `ALTER TABLE` — dapat diterima karena himpunan nilai stabil sesuai PRD. Untuk status Midtrans yang lebih cair, payload mentah tetap disimpan di `payments.raw_notification` sebagai cadangan.

**5. Kolom `promos.value` polimorfik + snapshot di `transactions`.**
Satu kolom `value BIGINT` melayani dua tipe: persen (0–100) untuk `percentage`, dan nominal sen untuk `fixed`, dibedakan oleh kolom `type`. Hasil perhitungan diskon final tetap "dibekukan" ke `transactions.discount_amount` sehingga perubahan/penghapusan promo di kemudian hari tidak mengubah nilai transaksi historis.

**6. Snapshot nama & harga di `transaction_items`.**
`product_name` dan `price` disalin saat transaksi dibuat. Ketika Owner mengubah harga atau menonaktifkan/menghapus (soft delete) produk, struk dan laporan lama tetap menampilkan nilai yang benar pada saat transaksi terjadi.

**7. Stok: kolom cache `products.stock` + sumber kebenaran `stock_movements`.**
`products.stock` menyimpan saldo terkini untuk pembacaan cepat (cek ketersediaan saat kasir input), sedangkan `stock_movements` adalah ledger yang dapat direkonsiliasi. Pengurangan stok + insert movement bertipe `out` dilakukan dalam **satu transaksi DB** dan hanya dipicu setelah webhook Midtrans `settlement/capture` diterima (PRD Business Rule #5 & #13).

**8. Pemisahan `transactions` dan `payments` (1:1).**
Detail pembayaran (token Snap, status Midtrans, payload webhook) dipisah dari header transaksi agar tabel `transactions` tetap ramping untuk query laporan, dan agar siklus hidup pembayaran (retry, notifikasi berulang) tidak mengotori data penjualan. `raw_notification JSON` disimpan untuk audit & rekonsiliasi bila status meragukan.

**9. Aturan integritas yang ditegakkan di aplikasi (bukan FK).**
Beberapa invariant tidak bisa dijamin constraint MySQL: "satu shift open per kasir", "transaksi hanya saat shift open", "order tidak bisa diubah setelah checkout" (Business Rule #14/#16). Ini ditegakkan di service layer Go menggunakan transaksi DB dan locking (`SELECT ... FOR UPDATE`), didukung index pendukung di atas.
