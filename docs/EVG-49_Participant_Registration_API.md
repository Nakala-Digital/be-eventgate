# Technical Documentation: EVG-49 [BE4] Participant Registration API

---

## 1. Document Metadata

| Attribute         | Details                                                      |
| :---------------- | :----------------------------------------------------------- |
| **Project**       | EventGate — Event Management System                          |
| **Task Code**     | EVG-49                                                       |
| **Feature**       | Participant Registration API                                 |
| **Tech Stack**    | Golang (Go), PostgreSQL, GORM                                |
| **Author**        | Hanif                                                        |
| **Status**        | Completed / Ready for Review                                 |

---

## 2. Executive Summary

Task **EVG-49** memperkenalkan fungsionalitas registrasi bagi peserta (Participant Registration API). Peserta dari luar sistem (tanpa perlu memiliki akun / *guest*) dapat melihat daftar tiket dan form pendaftaran dinamis untuk event yang sudah di-*publish*, melakukan pendaftaran (registrasi), serta melihat detail registrasinya menggunakan kode registrasi unik. Selain itu, ada perbaikan krusial pada integrasi GORM dan Database Migration untuk menghindari inkonsistensi skema.

---

## 3. API Reference

Berikut spesifikasi *endpoint* baru yang dapat dikonsumsi.
*Seluruh endpoint ini bersifat publik (di bawah subtree `/api/public/`), dan **Sengaja tanpa** HTTP Header `Authorization` (Tidak memerlukan `RequireAuth`).*

**HTTP Response Code yang Digunakan:**
* `200 OK`: Permintaan berhasil diproses.
* `201 Created`: Registrasi baru berhasil dibuat.
* `400 Bad Request`: Validasi input gagal, format permintaan tidak sesuai, atau form tidak lengkap.
* `404 Not Found`: Event tidak ditemukan / belum *publish*, atau kode registrasi tidak valid.
* `409 Conflict`: Tiket sudah habis, atau email sudah terdaftar pada event ini.
* `500 Internal Server Error`: Terjadi kesalahan pada server atau basis data.

### 3.1. Get Public Ticket Types
*   **URL:** `GET /api/public/events/{id}/ticket-types`
*   **Fungsi:** Mengambil daftar tiket yang berstatus aktif untuk sebuah event yang *Published*.
*   **Akses:** Publik (tanpa autentikasi).

### 3.2. Get Public Dynamic Questions
*   **URL:** `GET /api/public/events/{id}/questions`
*   **Fungsi:** Mengambil struktur form registrasi dinamis untuk sebuah event yang *Published*.
*   **Akses:** Publik (tanpa autentikasi).

### 3.3. Register to Event
*   **URL:** `POST /api/public/events/{id}/register`
*   **Fungsi:** Mendaftarkan peserta ke event tertentu dengan melampirkan pilihan tiket, data diri peserta, serta jawaban dari pertanyaan dinamis.
*   **Akses:** Publik (tanpa autentikasi).

### 3.4. Get Registration By Code
*   **URL:** `GET /api/public/registrations/{code}`
*   **Fungsi:** Mengambil detail registrasi peserta berdasarkan `registration_code`. Digunakan sebagai "kunci akses" bagi peserta karena peserta tidak memiliki akun sistem.
*   **Akses:** Publik (tanpa autentikasi).

---

## 4. Frontend Integration Guide

Panduan bagi tim Frontend untuk mengimplementasikan form registrasi dinamis:

### A. Pre-Registration Data Fetching
* Sebelum menampilkan form registrasi, sistem di Frontend perlu memanggil `GET /api/public/events/{id}/ticket-types` dan `GET /api/public/events/{id}/questions`.
* Form tidak boleh diizinkan untuk di-submit apabila ticket-types yang dipilih sudah *sold out*.

### B. Validasi Dynamic Form di Klien
* Backend telah menangani validasi *conditional requirement* dan *select validation*. Namun, disarankan agar aplikasi klien juga menaati aturan tersebut untuk User Experience (UX) yang lebih responsif.
* Jika `requirement_type = kondisional`, elemen form tersebut wajib diisi hanya jika pertanyaan `depends_on_question_id` dijawab sesuai dengan nilai `depends_on_value`.

### C. Post-Registration
* Ketika proses registrasi berhasil (HTTP `201 Created`), Backend akan mengembalikan sebuah `registration_code`.
* Frontend harus menyimpan kode ini (misalnya di URL, *LocalStorage*, atau diinstruksikan untuk dicatat oleh pengguna) dan bisa mengarahkan ke halaman ringkasan dengan memanggil endpoint `GET /api/public/registrations/{code}`.

---

## 5. Architecture & Security

* **Transactional Atomicity (ACID):** Pembuatan partisipan baru (atau update partisipan lama), pembaruan kuota tiket, dan penyisipan Form Responses dijalankan di dalam **satu transaksi database**. Hal ini mengamankan integritas data dari isu separuh data tersimpan apabila terjadi _failure_ di tengah proses.
* **Concurrency Handling untuk Kuota Tiket:** Pembaruan kapasitas tiket dilakukan menggunakan mekanisme *atomic in-place update* (`UpdateColumn("sold_count", gorm.Expr("sold_count + 1"))` dilengkapi pengecekan klausa WHERE untuk mematuhi *max_capacity*). Hal ini mencegah terjadinya *overselling* saat banyak permintaan registrasi masuk secara konkuren (Race Conditions).
* **Idempotency Peserta:** Kombinasi unik `(email, name)` digunakan untuk melakukan resolusi `Participant`. Jika sebuah email & nama sudah pernah terdaftar di database (misal dari event sebelumnya), sistem menggunakan kembali `participant_id` yang ada (*reuse*) daripada menciptakan *redundant data*.
* **Status Approval Registrasi:** Tiket berbayar secara otomatis mendapatkan status `pending_payment` sesuai ERD, dan tiket gratis langsung `confirmed`.
* **JSON Envelope:** Seluruh respons menggunakan field `success`, `message`, `data`, dan `errors`. Payload endpoint berada di dalam `data`.

---

## 6. Directory Layout & Modules

```text
be-eventgate/
├── docs/
│   └── EVG-49_Participant_Registration_API.md     # Dokumen ini
├── internal/
│   ├── handlers/
│   │   ├── registration_dto.go                    # Request/Response struktur untuk registrasi
│   │   ├── registration_handler.go                # Komponen HTTP Handler fungsi pendaftaran dan pencarian kode
│   │   └── registration_handler_test.go           # Unit & end-to-end routing Test
│   ├── models/
│   │   ├── participant.go                         # Definisi Skema Participant
│   │   ├── registration.go                        # Definisi Skema Registration (diperbaiki tag GORM-nya)
│   │   └── form_response.go                       # Definisi Skema JSON Response Form Dinamis
│   └── router/
│       └── router.go                              # Pendaftaran rute API baru (public subtree)
├── cmd/
│   └── server/
│       └── main.go                                # Dihapusnya fitur intdb.AutoMigrate
```

---

## 7. Catatan Implementasi & Penyelesaian Bug

### A. Penghapusan AutoMigrate di Lingkungan Production
Terdapat insiden pencemaran skema Database yang diakibatkan oleh `database.Migrate(gormDB)` (GORM AutoMigrate) yang dieksekusi dari `main.go`. GORM memodifikasi PostgreSQL dengan cara yang tidak sesuai dengan raw SQL migration (menghasilkan hilangnya constraint *NOT NULL*, menimpa nama kolom, dsb).
**Penyelesaian:** Panggilan `AutoMigrate` pada environment production dihapus permanen. Manajemen skema production murni bergantung secara sequential menggunakan `cmd/migrate/main.go`. `AutoMigrate` tetap dipertahankan, namun **hanya untuk mode unit-testing SQLite**.

### B. Bug: GORM Has-One vs Belongs-To Misinterpretation
Pada saat *fetching* registrasi dengan `Preload("Participant")`, GORM sempat mengalami bug fatal dengan memuat profil peserta yang salah. 
*   **Analisis:** Penempatan tag `gorm:"foreignKey:ParticipantID"` (tanpa referensi eksplisit) dipadukan dengan pemetaan Primary Key yang ambigu (field `ID` tetapi kolom `participant_id`), menyebabkan GORM mengambil keputusan default yang salah. GORM menebak relasinya sebagai **Has-One** (mencari `participant_id` di tabel Participant menggunakan Registration `ID`) alih-alih **Belongs-To**.
*   **Penyelesaian:** Tag `foreignKey` dihapus secara total untuk menggunakan *Convention over Configuration* milik GORM. GORM akhirnya cerdas menggunakan default relasi `Belongs-To` dan membaca data participant ID secara benar.

### C. Penyesuaian Requirement dan Penanganan Edge Cases
Terdapat beberapa keputusan teknis tambahan yang diimplementasikan untuk melengkapi *task brief* awal:
1. **Sinkronisasi Status Pembayaran:** ERD dan migration awal menetapkan `pending_payment`. Migration tambahan yang sebelumnya mengubahnya menjadi `waiting_payment` dihapus agar backend, frontend, dan database menggunakan nilai yang sama.
2. **Ketersediaan Endpoint Prasyarat:** Karena *payload* registrasi mewajibkan peserta untuk mengirimkan `ticket_type_id` dan jawaban dinamis, maka ditambahkan dua endpoint publik minimal sebagai prasyarat: `GET /api/public/events/{id}/ticket-types` dan `GET /api/public/events/{id}/questions`. Tanpa kedua endpoint ini, aplikasi frontend klien (yang *notabene* tidak memiliki akses *auth*) tidak akan memiliki cara untuk mengetahui tiket apa saja yang tersedia untuk dipilih.
3. **Pembatasan Registrasi Ganda (Duplikasi Entri):** Untuk mencegah satu peserta (berdasarkan validasi email) mendaftar berkali-kali pada event yang sama secara tidak wajar, diterapkan sebuah mekanisme *blocking*. Sistem memvalidasi bahwa 1 email hanya diperbolehkan memiliki 1 registrasi aktif per event (dengan status selain *cancelled*).

---

## 8. Design Decisions

* **Public Subtree:** API registrasi dikelompokkan dalam sub-rute `/api/public` agar *Middleware RequireAuth* yang global bisa dengan rapi mengecualikan (bypass) *authorization token*.
* **Registration Code vs ID:** Peserta yang tidak memiliki autentikasi tidak menggunakan ID Integer untuk memeriksa kembali pendaftarannya (mencegah *Insecure Direct Object Reference (IDOR)* atau enumeration). Sistem merespon menggunakan `RegistrationCode` yang diacak secara acak dan kompleks.
* **Database State Lockdown:** Menjaga kebersihan skema dengan mengunci struktur tabel agar murni berasal dari file migration (`migrations/*.sql`).
