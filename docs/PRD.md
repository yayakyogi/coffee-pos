# PRD — Coffee Shop Point of Sales (POS)

**Versi:** 1.0
**Tanggal:** 14 Juni 2026
**Stack:** Go (backend) + MySQL (database)
**Status:** Draft

---

## 1. Deskripsi Produk

**Coffee POS** adalah sistem Point of Sales berbasis web untuk coffee shop skala menengah (1 outlet, beberapa kasir, kapasitas puluhan meja). Sistem ini menangani seluruh alur operasional harian: pencatatan produk dan stok, pengelolaan meja, pembuatan transaksi penjualan, penerapan promo, hingga pembayaran digital melalui Midtrans Snap. Selain itu sistem menyediakan dashboard pelaporan bagi pemilik usaha untuk memantau kinerja penjualan.

**Untuk siapa:**
- **Owner / Pemilik** — membutuhkan kontrol penuh atas master data (produk, kategori, stok, meja, user) serta visibilitas terhadap performa bisnis melalui laporan.
- **Cashier / Kasir** — staf operasional yang menjalankan shift, melayani pelanggan, membuat transaksi, dan memproses pembayaran.

**Tujuan utama:**
- Mempercepat proses transaksi di kasir.
- Menjaga akurasi stok melalui pencatatan pergerakan stok otomatis.
- Memberikan akuntabilitas per kasir melalui mekanisme shift dengan modal kas.
- Menyediakan data penjualan yang dapat dianalisis oleh owner.

---

## 2. User Roles

Sistem memiliki **dua role** dengan hak akses yang berbeda. Otentikasi wajib untuk semua akses; setiap aksi tercatat terhadap user yang melakukannya.

### 2.1 Owner
Akses penuh ke seluruh fitur manajemen dan pelaporan. Owner **tidak** melakukan transaksi penjualan harian (bukan operasional kasir), namun dapat melihat seluruh data transaksi.

Akses Owner mencakup:
- Manajemen produk, kategori, dan stok.
- Manajemen meja.
- Manajemen user cashier (membuat, menonaktifkan, reset password).
- Manajemen promo dan diskon.
- Dashboard laporan dan export CSV.

### 2.2 Cashier
Akses terbatas pada operasional penjualan. Cashier **tidak dapat** mengubah master data (produk, harga, stok, meja, user) maupun melihat laporan agregat lintas-kasir.

Akses Cashier mencakup:
- Manajemen shift miliknya sendiri (buka & tutup shift).
- Membuat transaksi penjualan baru.
- Menerapkan promo aktif ke transaksi.
- Checkout & pembayaran via Midtrans Snap.
- Melihat riwayat transaksi pada shift hari ini (miliknya sendiri).

**Acceptance Criteria (Roles):**
- Login dengan kredensial salah ditolak; sesi memiliki masa berlaku.
- Cashier yang mengakses endpoint khusus Owner menerima respons `403 Forbidden`.
- Setiap entitas transaksional menyimpan `created_by` / `cashier_id` yang valid.

---

## 3. Fitur Owner

### 3.1 Manajemen Produk
CRUD produk dengan dukungan **soft delete**, upload **foto**, dan **status aktif/nonaktif**.

Atribut produk minimal: nama, kategori, harga, deskripsi, foto, status (aktif/nonaktif), flag terhapus (`deleted_at`).

**Acceptance Criteria:**
- Owner dapat membuat, melihat, mengubah, dan menghapus produk.
- Hapus produk bersifat soft delete: data tidak hilang dari DB (`deleted_at` terisi) dan tidak muncul di daftar produk aktif maupun di pilihan produk kasir.
- Produk berstatus **nonaktif** tidak dapat dipilih saat membuat transaksi, namun tetap muncul di daftar manajemen produk Owner.
- Foto produk dapat diunggah (format gambar umum, dengan batas ukuran) dan ditampilkan; produk tanpa foto tetap valid.
- Harga harus berupa angka ≥ 0 dan wajib diisi.

### 3.2 Manajemen Kategori Produk
CRUD kategori untuk mengelompokkan produk (mis. Coffee, Non-Coffee, Pastry).

**Acceptance Criteria:**
- Owner dapat membuat, mengubah, dan menghapus kategori.
- Kategori yang masih dipakai oleh produk aktif tidak dapat dihapus, atau dihapus dengan mekanisme yang menjaga integritas (produk tidak menjadi tanpa kategori secara tidak sengaja).
- Setiap produk wajib memiliki tepat satu kategori.

### 3.3 Manajemen Stok dengan Riwayat Pergerakan
Pengelolaan stok per produk beserta **riwayat pergerakan stok (stock movement)** yang mencatat setiap perubahan.

Setiap pergerakan stok mencatat: produk, tipe (`in` / `out` / `adjustment`), jumlah, stok sebelum, stok sesudah, alasan/keterangan, user, dan waktu.

**Acceptance Criteria:**
- Owner dapat menambah stok (restock) dan melakukan penyesuaian (adjustment) dengan keterangan.
- Setiap transaksi penjualan yang berhasil dibayar otomatis mengurangi stok produk terkait dan mencatat pergerakan bertipe `out`.
- Riwayat pergerakan stok dapat dilihat per produk, terurut waktu, dan menampilkan saldo stok sebelum & sesudah.
- Stok tidak dapat menjadi negatif; upaya transaksi melebihi stok tersedia ditolak (lihat Business Rules).

### 3.4 Manajemen Meja
CRUD meja untuk dipilih saat transaksi (dine-in).

Atribut meja minimal: nomor/nama meja, kapasitas (opsional), status (tersedia/nonaktif).

**Acceptance Criteria:**
- Owner dapat membuat, mengubah, dan menonaktifkan meja.
- Meja nonaktif tidak muncul sebagai pilihan saat membuat transaksi.
- Nomor/nama meja bersifat unik.

### 3.5 Manajemen User Cashier
Owner mengelola akun kasir.

**Acceptance Criteria:**
- Owner dapat membuat akun cashier (username unik, password, nama).
- Owner dapat menonaktifkan akun cashier; akun nonaktif tidak dapat login.
- Owner dapat mereset password cashier.
- Password disimpan dalam bentuk hash (mis. bcrypt), tidak pernah plaintext.
- Cashier yang sedang memiliki shift terbuka tidak dapat dinonaktifkan sebelum shift ditutup (atau ditangani dengan peringatan eksplisit).

### 3.6 Dashboard Laporan
Ringkasan kinerja penjualan untuk Owner.

Laporan yang tersedia:
- **Revenue** harian, mingguan, dan bulanan.
- **Produk terlaris** (berdasarkan kuantitas terjual pada rentang waktu).
- **Transaksi per cashier** (jumlah transaksi & total nilai per kasir).

**Acceptance Criteria:**
- Owner dapat memfilter laporan berdasarkan rentang tanggal.
- Angka revenue hanya menghitung transaksi berstatus **paid/settlement** (transaksi pending/gagal tidak dihitung).
- Produk terlaris menampilkan minimal nama produk dan total kuantitas terjual, terurut menurun.
- Laporan per cashier menampilkan jumlah transaksi dan total nilai, dapat difilter per periode.
- Nilai pada dashboard konsisten dengan data transaksi mentah (dapat direkonsiliasi).

### 3.7 Manajemen Promo dan Diskon
CRUD promo dengan dua tipe diskon: **persentase** (mis. 10%) dan **nominal** (mis. Rp10.000).

Atribut promo minimal: nama, tipe (`percentage` / `fixed`), nilai, periode berlaku (mulai–selesai), status aktif, dan opsional minimum pembelian.

**Acceptance Criteria:**
- Owner dapat membuat promo bertipe persentase atau nominal.
- Promo memiliki periode berlaku; di luar periode atau saat nonaktif, promo tidak muncul sebagai promo aktif bagi kasir.
- Diskon persentase: nilai 0–100. Diskon nominal: nilai ≥ 0 dan tidak boleh membuat total transaksi menjadi negatif (total minimal Rp0).
- Promo dengan minimum pembelian hanya dapat diterapkan jika subtotal memenuhi syarat.

### 3.8 Export Laporan ke CSV
Owner dapat mengekspor laporan ke format CSV.

**Acceptance Criteria:**
- Owner dapat mengunduh file CSV untuk laporan transaksi pada rentang tanggal terpilih.
- File CSV memiliki header kolom yang jelas dan dapat dibuka di spreadsheet umum (encoding UTF-8).
- Data pada CSV konsisten dengan yang ditampilkan di dashboard untuk periode yang sama.

---

## 4. Fitur Cashier

### 4.1 Manajemen Shift
Cashier membuka shift dengan **modal kas (opening balance)** dan menutup shift dengan **rekap**.

**Acceptance Criteria:**
- Cashier membuka shift dengan memasukkan nominal modal kas awal; sistem mencatat waktu buka dan kasir.
- Seorang cashier hanya boleh memiliki **satu shift terbuka** pada satu waktu.
- Transaksi hanya dapat dibuat ketika cashier memiliki shift yang terbuka.
- Saat menutup shift, sistem menampilkan rekap: modal awal, total penjualan tunai/non-tunai selama shift, jumlah transaksi, dan ekspektasi kas akhir.
- Cashier memasukkan kas akhir aktual; selisih (over/short) dihitung dan disimpan.
- Setelah ditutup, shift tidak dapat dibuka kembali dan transaksinya bersifat final.

### 4.2 Buat Transaksi Baru
Cashier memilih **meja**, memilih **produk**, dan mengatur **quantity**.

**Acceptance Criteria:**
- Cashier dapat memilih satu meja yang berstatus tersedia/aktif.
- Cashier dapat menambahkan satu atau lebih item produk (hanya produk aktif & belum terhapus) dengan quantity ≥ 1.
- Subtotal per item dan total transaksi dihitung otomatis dan ter-update saat item/quantity berubah.
- Quantity yang melebihi stok tersedia ditolak dengan pesan jelas.
- Transaksi tanpa item tidak dapat diproses ke checkout.

### 4.3 Apply Promo Aktif
Cashier menerapkan promo yang sedang aktif ke transaksi.

**Acceptance Criteria:**
- Hanya promo yang aktif dan dalam periode berlaku yang dapat dipilih.
- Sistem menghitung ulang total setelah diskon (persentase atau nominal) dan menampilkan besaran potongan.
- Promo dengan syarat minimum pembelian ditolak bila subtotal belum memenuhi.
- Maksimal satu promo per transaksi (kecuali ditentukan lain), dan total akhir tidak pernah negatif.

### 4.4 Checkout dengan Midtrans Snap
Pembayaran diproses melalui **Midtrans Snap**.

**Acceptance Criteria:**
- Saat checkout, sistem membuat transaksi berstatus **pending** dan memperoleh Snap token/URL dari Midtrans.
- Cashier dapat menampilkan halaman/pop-up pembayaran Snap kepada pelanggan.
- Status transaksi diperbarui berdasarkan hasil pembayaran (settlement/paid, pending, atau gagal) — idealnya melalui notifikasi/webhook Midtrans, dengan verifikasi keabsahan notifikasi.
- Hanya transaksi yang **berhasil dibayar** yang mengurangi stok dan masuk ke perhitungan laporan.
- Transaksi gagal/dibatalkan tidak mengubah stok dan ditandai statusnya dengan jelas.

### 4.5 Riwayat Transaksi Shift Hari Ini
Cashier melihat transaksi pada shift berjalan hari ini.

**Acceptance Criteria:**
- Cashier hanya melihat transaksi miliknya pada shift hari ini (bukan milik kasir lain).
- Daftar menampilkan minimal: waktu, meja, total, status pembayaran.
- Detail transaksi dapat dibuka untuk melihat item, promo, dan total.

---

## 5. Business Rules

Aturan bisnis yang **wajib ditegakkan** oleh sistem:

1. **Otentikasi & otorisasi:** Semua akses memerlukan login. Hak akses ditegakkan per role; cashier tidak dapat mengakses fungsi Owner.
2. **Shift wajib untuk transaksi:** Cashier tidak dapat membuat transaksi tanpa shift terbuka. Satu cashier maksimal satu shift terbuka.
3. **Soft delete produk:** Produk yang dihapus tidak benar-benar dihilangkan dari database dan tidak dapat dipilih untuk transaksi baru, namun tetap terhubung ke transaksi historis (laporan lama tetap utuh).
4. **Hanya produk aktif yang dijual:** Produk nonaktif atau terhapus tidak boleh masuk transaksi baru.
5. **Stok tidak boleh negatif:** Transaksi yang melebihi stok tersedia ditolak. Pengurangan stok terjadi saat pembayaran berhasil, dan tercatat sebagai stock movement bertipe `out`.
6. **Validasi promo:** Promo hanya berlaku dalam periode aktif. Diskon persentase 0–100%. Diskon nominal tidak membuat total < 0. Syarat minimum pembelian harus terpenuhi.
7. **Konsistensi laporan:** Revenue dan laporan hanya menghitung transaksi berstatus paid/settlement. Transaksi pending atau gagal tidak masuk perhitungan.
8. **Integritas pembayaran:** Status transaksi mengikuti status resmi dari Midtrans; perubahan status hanya sah bila berasal dari notifikasi terverifikasi. Tidak ada transaksi yang ditandai lunas tanpa konfirmasi pembayaran.
9. **Finalitas shift:** Shift yang sudah ditutup bersifat final; transaksinya tidak dapat diubah. Selisih kas (over/short) tercatat.
10. **Keamanan kredensial:** Password disimpan sebagai hash. Akun nonaktif tidak dapat login.
11. **Akuntabilitas:** Setiap transaksi terikat pada cashier dan shift; setiap pergerakan stok terikat pada user dan waktu.
12. **Mata uang & pembulatan:** Seluruh nilai uang menggunakan satuan Rupiah (tanpa desimal) dengan aturan pembulatan yang konsisten di perhitungan subtotal, diskon, dan total.
13. **Stok berkurang setelah webhook confirmed:** Pengurangan stok hanya terjadi setelah sistem menerima **webhook/notifikasi Midtrans yang terkonfirmasi** (status settlement/capture) dan terverifikasi keabsahannya. Selama transaksi masih pending atau hanya dibuat di sisi kasir (sebelum webhook), stok belum berkurang.
14. **Buka shift sebelum transaksi:** Cashier harus membuka shift terlebih dahulu sebelum dapat membuat transaksi baru. Tanpa shift terbuka, pembuatan transaksi ditolak.
15. **Satu promo per transaksi:** Satu transaksi hanya boleh menggunakan **satu promo**. Promo tidak dapat ditumpuk (no stacking).
16. **Order final setelah checkout:** Order yang sudah masuk tahap checkout **tidak dapat diubah itemnya** (tambah/hapus item maupun perubahan quantity). Perubahan hanya mungkin dengan membatalkan order dan membuat order baru.

---

## 6. Out of Scope

Fitur berikut **tidak** dibangun pada versi ini (v1.0):

- **Multi-outlet / multi-cabang** — sistem hanya untuk satu outlet.
- **Manajemen inventori bahan baku (recipe/BOM)** — stok dikelola pada level produk jadi, bukan bahan baku.
- **Integrasi hardware** — printer struk, cash drawer, barcode scanner, dan customer display tidak termasuk.
- **Metode pembayaran selain Midtrans Snap** — tidak ada integrasi EDC bank lain atau pencatatan pembayaran tunai terotomasi di luar rekap shift.
- **Program loyalty / membership / poin pelanggan.**
- **Manajemen pelanggan (CRM)** dan database pelanggan.
- **Online ordering / aplikasi pelanggan / QR self-order.**
- **Manajemen pengiriman (delivery) dan integrasi ojek online.**
- **Akuntansi penuh** (jurnal, neraca, pajak otomatis) — sistem hanya menyediakan laporan penjualan dasar dan export CSV.
- **Aplikasi mobile native** — versi ini berbasis web.
- **Manajemen jadwal/absensi karyawan** di luar mekanisme shift kasir.
- **Refund / void transaksi yang sudah lunas** melalui sistem — ditangani manual di luar v1.0.
