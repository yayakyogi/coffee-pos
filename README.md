# Coffee Shop POS

Backend API untuk sistem **Point of Sales coffee shop skala menengah**, dibangun dengan **Go** dan **MySQL**. Sistem ini menangani manajemen produk & stok, meja, shift kasir, transaksi penjualan, promo/diskon, hingga pembayaran digital via **Midtrans Snap**, serta dashboard pelaporan untuk Owner.

Dua role pengguna: **Owner** (master data + laporan) dan **Cashier** (operasional shift & transaksi).

## Dokumentasi

- [PRD](docs/PRD.md) — Product Requirements Document
- [Database Schema](docs/SCHEMA.md) — desain tabel MySQL
- [API Contract](docs/API_CONTRACT.md) — daftar endpoint, auth, format response
- [Postman Collection](docs/postman_collection.json) — skeleton request semua endpoint

## Tech Stack

- **Go** — REST API
- **MySQL** — penyimpanan utama
- **Redis** — caching / session
- **JWT** — autentikasi (role-based: prefix `/owner` & `/cashier`)
- **Midtrans Snap** — payment gateway

## Struktur Project

```
cmd/api/            # entry point aplikasi (main.go)
internal/
  entity/           # domain struct
  repository/       # repository interfaces
    mysql/          # implementasi MySQL
  service/          # business logic
  handler/          # HTTP handlers
  middleware/       # middleware (auth, role, dll)
  dto/              # request/response struct
pkg/
  database/         # koneksi database
  redis/            # koneksi Redis
  jwt/              # JWT helper
  response/         # response helper
  validator/        # input validator
migrations/         # file SQL migration
config/             # konfigurasi aplikasi
docs/               # PRD, schema, API contract, Postman
```

## Getting Started

```bash
# 1. Salin environment template lalu sesuaikan
cp .env.example .env

# 2. Jalankan aplikasi
go run ./cmd/api
```

Output yang diharapkan: `Coffee Shop POS starting...`

## Konfigurasi

Semua konfigurasi dibaca dari environment variables — lihat [.env.example](.env.example) untuk daftar lengkap (app, MySQL, Redis, JWT, Midtrans).
