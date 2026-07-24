# Technical Documentation: EVG-41 [BE] Authentication & Role-Based Access Control Setup

---

## Document Metadata

| Attribute         | Details                                                      |
| :---------------- | :----------------------------------------------------------- |
| **Project**       | EventGate — Event Management System                          |
| **Task Code**     | EVG-41 (Sprint 2)                                            |
| **Feature**       | Authentication, JWT, and Role-Based Access Control (RBAC)    |
| **Tech Stack**    | Golang (Go), PostgreSQL, `golang-jwt/v5`, `bcrypt`           |
| **Author**        | Backend Development Team                                     |
| **Status**        | Completed / Ready for Review                                 |

---

## 1. Executive Summary

Task **EVG-41** mengimplementasikan fondasi keamanan inti untuk aplikasi EventGate. Implementasi ini mencakup pembuatan model pengguna (*User* dan *Role*), sistem *hashing* kata sandi menggunakan algoritma *Bcrypt*, penerbitan *JSON Web Token* (JWT) untuk manajemen sesi *stateless*, serta sekumpulan *Middleware* yang bertugas memvalidasi token dan memblokir rute berdasarkan *Role-Based Access Control* (RBAC).

---

## 2. Role & Permission Matrix (RBAC)

Sistem saat ini mendefinisikan lima tingkatan peran (*role*).

**Catatan Penting:** Peran `peserta` difungsikan khusus untuk proses pendaftaran *guest* dan secara sistematis **tidak diizinkan untuk memiliki akun login**.

| Role Name | Hak Akses dan Keterangan | Izin Login |
| :--- | :--- | :---: |
| `super_admin` | Pemilik sistem. Memiliki hak akses penuh untuk menyetujui, menolak, dan mempublikasikan *event*. | Ya |
| `admin_panitia` | Pembuat acara (*Event Organizer*). Terbatas hanya dapat mengelola *event* miliknya sendiri. | Ya |
| `staf_lapangan` | Petugas di lokasi. Terbatas pada akses memindai tiket QR untuk proses *check-in* peserta. | Ya |
| `school_reviewer` | Peran peninjau. Bertugas memverifikasi draf *event* sebelum publikasi. | Ya |
| `peserta` | Peserta acara. Terdaftar via pendaftaran tamu (tanpa kata sandi). | Tidak |

### Penjelasan Peran: `school_reviewer`
Peran `school_reviewer` dipersiapkan untuk mengakomodasi **Alur Persetujuan Bertingkat (*Hierarchical Approval Workflow*)** di lingkungan instansi pendidikan. Apabila panitia (`admin_panitia`) mengajukan proposal suatu acara, pihak peninjau (seperti bagian kemahasiswaan) dapat menggunakan peran ini untuk meninjau, menyetujui, atau menolak acara tersebut. Peran ini memisahkan hak peninjauan dari hak administratif penuh (`super_admin`), sehingga mencegah pihak peninjau memiliki kontrol akses berbahaya seperti modifikasi konfigurasi sistem atau penghapusan pengguna secara massal.

---

## 3. Frontend Integration Guide

Bagian ini disusun untuk memberikan panduan teknis kepada tim **Frontend** dalam melakukan integrasi fitur autentikasi.

### A. Alur Pemrosesan Token
1. Klien mengirimkan kredenisal (Email dan Password) ke *endpoint* `/api/auth/login`.
2. *Backend* memvalidasi kredensial dan membalas dengan **JWT Token**.
3. Klien **wajib menyimpan** JWT Token tersebut (direkomendasikan pada *Secure Cookie* atau *localStorage*).
4. Untuk setiap permintaan selanjutnya ke *endpoint* yang terproteksi (seperti profil atau dasbor), klien **wajib** menyisipkan token tersebut di HTTP Headers dengan format:
   `Authorization: Bearer <ISI_TOKEN>`

### B. Daftar Status Code HTTP
* **200 OK**: Permintaan berhasil diproses.
* **401 Unauthorized**: Gagal autentikasi (kredensial salah), Token kadaluarsa, atau Token tidak valid. *Aksi Klien yang Disarankan: Alihkan pengguna ke halaman Login.*
* **403 Forbidden**: Token valid, tetapi pengguna tidak memiliki hak otorisasi yang mencukupi untuk memanggil sumber daya terkait. *Aksi Klien yang Disarankan: Tampilkan pesan "Akses Ditolak".*

---

## 4. API Reference

Berikut adalah spesifikasi teknis *endpoint* yang tersedia.

### 4.1. Login Endpoint
Mengautentikasi pengguna dan mengembalikan JWT Token.

* **URL:** `/api/auth/login`
* **Method:** `POST`
* **Auth Required:** No

**Request Body (JSON):**
```json
{
  "email": "panitia@eventgate.com",
  "password": "Rahasia123!"
}
```

**Success Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 2,
    "username": "panitia",
    "email": "panitia@eventgate.com",
    "role_name": "admin_panitia",
    "is_active": true
  }
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "invalid email or password"
}
```
*(Catatan: Akun yang dinonaktifkan akan menerima pesan galat khusus).*

### 4.2. Get Current Profile
Mendapatkan profil pengguna yang sedang aktif (direkomendasikan untuk *refresh state* pada Frontend saat terjadi pemuatan ulang halaman).

* **URL:** `/api/auth/me`
* **Method:** `GET`
* **Auth Required:** Yes (Header `Authorization: Bearer <token>`)

**Success Response (200 OK):**
```json
{
  "id": 2,
  "username": "panitia",
  "email": "panitia@eventgate.com",
  "role_name": "admin_panitia",
  "is_active": true
}
```

---

## 5. Security Architecture

* **Hashing:** Sistem menggunakan pustaka `golang.org/x/crypto/bcrypt` untuk enkripsi asimetris. Kata sandi mentah tidak pernah disimpan dalam basis data.
* **Prepared Statements:** Menggunakan kapabilitas `gorm` untuk mencegah injeksi SQL pada seluruh kueri autentikasi.
* **JWT Signature:** JWT ditandatangani menggunakan algoritma *HS256*. Rahasia enkripsi disimpan di dalam variabel lingkungan (`JWT_SECRET`).
* **Context Propagation:** *User ID* dan *Role* yang tervalidasi akan diinjeksi ke dalam siklus hidup permintaan menggunakan `context.Context` dari utilitas HTTP baku Go.

---

## 6. Directory Layout & Module Structure

```text
be-eventgate/
├── config/
│   └── config.go                  # Mengelola JWT_SECRET dan JWT_EXPIRY
├── internal/
│   ├── auth/
│   │   ├── jwt.go                 # Logika Pembangkitan dan Validasi Token
│   │   ├── password.go            # Utilitas Hashing Bcrypt
│   │   └── permission.go          # Matriks Hak Akses Statis
│   ├── middleware/
│   │   ├── auth_middleware.go     # Verifikasi integritas JWT Token
│   │   ├── rbac_middleware.go     # Verifikasi otoritas pengguna (Role)
│   │   └── context.go             # Manajemen State dan Injeksi Context
│   ├── handlers/
│   │   ├── auth_handler.go        # Komponen Handler HTTP untuk Login
│   │   └── user_handler.go        # Komponen Handler HTTP untuk Profil
│   └── router/
│       └── router.go              # Registrasi dan Pengelompokan Rute API
```
