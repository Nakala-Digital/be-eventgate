# EventGate Backend (`be-eventgate`)

Backend service untuk **EventGate (Event Management System)** yang dibangun menggunakan **Golang (Go)**, **Chi Router**, **GORM**, dan **PostgreSQL**. Service ini mengelola autentikasi, manajemen pengguna (RBAC), siklus hidup event, pendaftaran, pembayaran, dan pemindaian tiket QR.

---

## 📚 Project Documentation Index

Dokumentasi teknis proyek tersimpan di folder [`docs/`](file:///c:/Users/Lenovo/Documents/JTK/Akademik/KuliahPraktik/Romulus/Dev/Event-Gate/be-eventgate/docs/):

- ⚙️ **[EVG-38: Backend Project Structure & API Foundation Setup](file:///c:/Users/Lenovo/Documents/JTK/Akademik/KuliahPraktik/Romulus/Dev/Event-Gate/be-eventgate/docs/EVG-38_Backend_Setup.md)**  
  Arsitektur backend, struktur direktori, konfigurasi `.env`, PostgreSQL connection pool, middleware, dan health check API.
- 🎨 **[EVG-39: Frontend Setup & Architecture Specification](file:///c:/Users/Lenovo/Documents/JTK/Akademik/KuliahPraktik/Romulus/Dev/Event-Gate/be-eventgate/docs/EVG-39_Frontend_Setup.md)**  
  Arsitektur frontend, integrasi API, dan standar desain UI/UX.
- 🗄️ **[EVG-40: Database Migration Setup](file:///c:/Users/Lenovo/Documents/JTK/Akademik/KuliahPraktik/Romulus/Dev/Event-Gate/be-eventgate/docs/EVG-40_Database_Migration_Setup.md)**  
  Spesifikasi DDL skema 14 tabel database PostgreSQL, foreign key, index, dan perkakas CLI migration runner.
- 🔐 **[EVG-41: Authentication & RBAC Setup](file:///c:/Users/Lenovo/Documents/JTK/Akademik/KuliahPraktik/Romulus/Dev/Event-Gate/be-eventgate/docs/EVG-41_Auth_RBAC_Setup.md)**  
  Logika autentikasi JWT, password hashing Bcrypt, matriks 5 role RBAC, dan middleware otorisasi.
- 📅 **[EVG-43: Event CRUD API Implementation](file:///c:/Users/Lenovo/Documents/JTK/Akademik/KuliahPraktik/Romulus/Dev/Event-Gate/be-eventgate/docs/EVG-43_Event_CRUD_API.md)**  
  Spesifikasi endpoint CRUD Event, validasi field minimum, aturan bisnis event berbayar/gratis, dan proteksi otorisasi berbasis peran.

---

## 🚀 Quickstart Guide

### 1. Prasyarat Sistem
- **Go**: Version 1.20+
- **PostgreSQL**: Terinstal lokal / Docker container

### 2. Konfigurasi Lingkungan (`.env`)
Salin file `.env.example` ke `.env` dan sesuaikan kredensial basis data Anda:
```powershell
Copy-Item .env.example .env
```

### 3. Eksekusi Migrasi Database
Jalankan migrasi skema awal dan seed data peran:
```powershell
go run cmd/migrate/main.go up
```

### 4. Jalankan Server Development
```powershell
go mod tidy
go run main.go
```

### 5. Verifikasi Server
Akses endpoint health check melalui browser atau API client:
- `GET http://localhost:8080/api/v1/health`