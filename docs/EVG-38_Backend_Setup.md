# Technical Documentation: EVG-38 [BE1] Backend Project Structure & API Foundation Setup

---

## Document Metadata

| Attribute | Details |
| :--- | :--- |
| **Project** | EventGate - Event Management System |
| **Task Code** | EVG-38 (Sprint 2) |
| **Feature** | Backend Project Structure & API Foundation Setup |
| **Tech Stack** | Golang (Go), PostgreSQL, `github.com/go-chi/chi/v5` |
| **Author** | Backend Development Team |
| **Status** | Completed / Ready for Review |

---

## 1. Executive Summary

Task **EVG-38** bertujuan untuk membangun fondasi arsitektur awal pada repositori backend `be-eventgate`. Pengerjaan ini menyediakan struktur folder modular berbasis *Clean Architecture*, konfigurasi *environment*, koneksi database PostgreSQL, middleware dasar, standar format respon API (JSON), serta *health check endpoint* sebagai pijakan pengembangan fitur-fitur Sprint 2 berikutnya (seperti *Event Management*, *Dynamic Form*, *QR Ticket*, dan *Payment Gateway*).

---

## 2. Directory Layout & Architecture

Proyek ini menerapkan struktur direktori terisolasi (*Clean Architecture*) agar kode mudah dipelihara, mempermudah kolaborasi antar tim *full-stack*, dan mencegah *merge conflict*:

```text
be-eventgate/
├── config/
│   ├── config.go         # Environment variables reader & struct (.env)
│   └── database.go       # PostgreSQL database connection pool & health ping
├── docs/                 # Dokumentasi proyek & requirement Sprint 2
│   └── EVG-38_Backend_Setup.md
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── middleware/
│   │       │   └── middleware.go  # Custom HTTP middlewares (CORS)
│   │       └── router.go          # Base Chi Router & API routing definitions
│   ├── domain/           # Entities, Contracts, and Data Models (Layer 1)
│   ├── repository/       # Data Access Layer & Database Queries (Layer 2)
│   └── usecase/          # Business Logic Layer (Layer 3)
├── pkg/
│   └── utils/
│       └── response/
│           └── response.go        # Standardized JSON response helper (Success/Error)
├── .env.example          # Template environment variable
├── .env                  # Environment file lokal
├── go.mod                # Go module file
├── go.sum                # Go checksum database
├── main.go               # Application entry point
└── README.md             # Project quickstart guide
```

---

## 3. Core Foundation Components

### A. Environment Configuration (`config/config.go`)
Membaca konfigurasi aplikasi dari file `.env` menggunakan library `github.com/joho/godotenv` dengan dukungan fallback default value:
- **Server Port**: `8080` (default)
- **Environment**: `development` (default)
- **Database Credentials**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`.

### B. Database Connection Setup (`config/database.go`)
Menginisialisasi driver PostgreSQL (`github.com/lib/pq`) dengan konfigurasi Connection Pool teroptimasi:
- **Max Open Connections**: 25
- **Max Idle Connections**: 5
- **Connection Max Lifetime**: 15 Menit

### C. Standardized Response Helper (`pkg/utils/response/response.go`)
Memastikan seluruh endpoint menghasilkan format respon JSON yang konsisten di seluruh ekosistem API EventGate:

```json
{
  "success": true,
  "message": "Pesan deskripsi status",
  "data": { ... },
  "errors": null
}
```

### D. Router & Middleware Engine (`internal/delivery/http/`)
- **Router Engine**: Menggunakan `github.com/go-chi/chi/v5` yang ringan dan kompatibel dengan `net/http`.
- **Global Middlewares**:
  - `middleware.Logger` (pencatatan request/response log).
  - `middleware.Recoverer` (penanganan panic recovery agar server tidak runtuh).
  - `customMiddleware.CORS` (penanganan Cross-Origin Resource Sharing untuk integrasi frontend).

---

## 4. Base API Endpoints

| HTTP Method | Endpoint Path | Description | Expected Status Code | Response Body |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/` | Welcoming Root Endpoint | `200 OK` | `{"success": true, "message": "Welcome to EventGate API"}` |
| `GET` | `/api/v1/health` | System Health Check | `200 OK` | `{"success": true, "message": "API is running healthy", "data": {"status": "UP"}}` |

---

## 5. Local Setup & Execution Guide

### Prasyarat System
- **Go Engine**: Minimal versi 1.20+
- **PostgreSQL**: Terinstal lokal atau via Docker container

### Quickstart Execution

1. **Inisialisasi Environment File**:
   ```powershell
   Copy-Item .env.example .env
   ```
   Atur kredensial `DB_USER`, `DB_PASSWORD`, dan `DB_NAME` pada file `.env` sesuai konfigurasi PostgreSQL lokal Anda.

2. **Unduh Dependensi**:
   ```powershell
   go mod tidy
   ```

3. **Jalankan Aplikasi**:
   ```powershell
   go run main.go
   ```

4. **Verifikasi Endpoint**:
   Buka browser atau API Client (Postman/Insomnia) dan akses:
   ```text
   http://localhost:8080/api/v1/health
   ```

---

## 6. Version Control & Commit Standard

### Recommended Commit Message

```bash
git add .
git commit -m "feat(setup): initialize project structure and base API foundation (EVG-38)"
git push origin <nama-branch>
```
