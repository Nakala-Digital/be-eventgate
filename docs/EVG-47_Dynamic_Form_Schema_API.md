# Technical Documentation: EVG-47 [BE3] Dynamic Registration Form API

---

## 1. Document Metadata

| Attribute         | Details                                                      |
| :---------------- | :----------------------------------------------------------- |
| **Project**       | EventGate — Event Management System                          |
| **Task Code**     | EVG-47                                                       |
| **Feature**       | Dynamic Registration Form API                                |
| **Tech Stack**    | Golang (Go), PostgreSQL, GORM                                |
| **Author**        | Hanif                                       |
| **Status**        | Completed / Ready for Review                                 |

---

## 2. Executive Summary

Task **EVG-47** memperkenalkan fitur form registrasi dinamis (`Dynamic Form Schema`) yang memungkinkan penyelenggara event (Admin Panitia) menyusun pertanyaan tambahan secara spesifik untuk setiap event. Pembaruan ini mencakup pengelolaan `question_type`, tingkat kewajiban (`requirement_type`), dukungan kustomisasi opsi (Question Options), mekanisme logika prasyarat (Conditional Question), serta penerapan Ownership Validation.

---

## 3. API Reference

Berikut spesifikasi *endpoint* baru yang dapat dikonsumsi.
*Semua endpoint memerlukan HTTP Header `Authorization: Bearer <TOKEN>`.*

**HTTP Response Code yang Digunakan:**
* `200 OK`: Permintaan berhasil diproses.
* `201 Created`: Entitas baru berhasil dibuat.
* `400 Bad Request`: Validasi input gagal atau format permintaan tidak sesuai.
* `401 Unauthorized`: Token tidak valid atau tidak ditemukan.
* `403 Forbidden`: Pengguna tidak memiliki akses ke entitas tersebut (contoh: bukan pemilik event).
* `404 Not Found`: Entitas event atau pertanyaan tidak ditemukan.
* `500 Internal Server Error`: Terjadi kesalahan pada server atau basis data.

### 3.1. Get Form Schema (List Questions)
*   **URL:** `GET /api/events/{id}/questions`
*   **Akses:** Terbatas untuk Super Admin, Reviewer, atau Pemilik Event. (Aturan visibilitas mengikuti aturan event approval).
*   **Query Params:** `include_inactive=true` (opsional) untuk melihat pertanyaan tersimpan yang sudah menggunakan status logical deletion.

### 3.2. Create Dynamic Question
*   **URL:** `POST /api/events/{id}/questions`
*   **Akses:** Hanya Pemilik Event (Admin Panitia) atau Super Admin.

### 3.3. Update Dynamic Question
*   **URL:** `PUT /api/events/{id}/questions/{questionID}`
*   **Akses:** Hanya Pemilik Event (Admin Panitia) atau Super Admin.

### 3.4. Delete Dynamic Question (Logical Deletion)
*   **URL:** `DELETE /api/events/{id}/questions/{questionID}`
*   **Akses:** Hanya Pemilik Event (Admin Panitia) atau Super Admin.

---

## 4. Frontend Integration Guide

Panduan bagi tim Frontend untuk mengimplementasikan form registrasi dinamis:

### A. Rendering Form & Conditional Logic
* Endpoint `GET /api/events/{id}/questions` mengembalikan array daftar pertanyaan yang telah terurut otomatis berdasarkan `display_order`.
* Untuk `requirement_type` bernilai `kondisional`, aplikasi di sisi klien (frontend) harus menyembunyikan pertanyaan tersebut secara *default*. Tampilkan input pertanyaan ini secara dinamis **hanya jika** pengguna telah menjawab pertanyaan pada ID `depends_on_question_id` dengan jawaban yang identik dengan nilai `depends_on_value`.

### B. Penggunaan `options`
* Jika `question_type = select`, `radio`, atau `checkbox` → Frontend **wajib** menggunakan array `options` untuk menyajikan pilihan jawaban.
* Selain tipe tersebut → field `options` dapat diabaikan.

### C. Pengelolaan Data Options (Replace-All Strategy)
* Saat proses pembaruan (`PUT`), integrasi harus menyertakan kembali seluruh array `options` (jika ada). Backend menggunakan pendekatan **Replace-All**, yaitu seluruh `QuestionOption` yang dimiliki suatu pertanyaan dihapus terlebih dahulu, kemudian seluruh option dibuat kembali berdasarkan payload terbaru dalam satu transaksi database.

---

## 5. Architecture & Security

* **Transactional Atomicity (ACID):** Pembuatan (`Create`) maupun pengubahan (`Update`) pertanyaan utama (`DynamicQuestion`) bersamaan dengan kumpulan Question Options diproses dalam satu transaksi basis data relasional (`DB.Transaction`). Hal ini memastikan tidak terjadi partial update maupun partial insert, sehingga mencegah munculnya pertanyaan yang tidak memiliki opsi jawaban utuh.
* **Logical Deletion (soft delete berbasis flag):** Delete diimplementasikan menggunakan logical deletion (`is_active = false`) sehingga data tidak dihapus secara fisik. Pendekatan ini membantu menjaga konsistensi data dengan mempertahankan record pertanyaan di basis data meskipun sudah tidak aktif, sehingga referensi dari entitas lain tetap dapat dipertahankan apabila diperlukan. Endpoint `GET` baku otomatis menyaring elemen nonaktif ini.

---

## 6. Directory Layout & Modules

```text
be-eventgate/
├── docs/
│   └── EVG-47_Dynamic_Form_Schema_API.md          # Dokumen ini
├── migrations/
│   └── 000001_create_initial_schema.up.sql         # ERD enum/check constraints
├── internal/
│   ├── handlers/
│   │   ├── question_dto.go                        # Tipe Request/Response baru
│   │   ├── question_handler.go                    # Komponen HTTP Handler dinamis
│   │   └── question_handler_test.go               # Test untuk Handler dinamis
│   ├── models/
│   │   ├── dynamic_question.go                    # Skema pertanyaan
│   │   └── question_option.go                     # Skema opsi pilihan ganda
│   └── router/
│       ├── router.go                              # Pendaftaran rute API baru
│       └── router_test.go                         # Integrasi end-to-end routing
```

---

## 7. Catatan Implementasi Dynamic Registration Form

### A. Penyesuaian `question_type`

**Referensi Dokumen**
* **EVG-30 – Dynamic Registration Form** mendefinisikan tipe field minimum sebagai: Text, Textarea, Dropdown, Checkbox, Radio Button, Upload File, Number, Date.
* **EVG-26 – URD** kemudian merevisi kebutuhan Sprint 2 dengan menyatakan bahwa: Upload File **tidak termasuk kebutuhan minimum Sprint 2** dan menjadi **future enhancement**.
* **EVG-34 – ERD** menggunakan enum database: `text, textarea, number, date, select, radio, checkbox` serta menyatakan `file_upload` sebagai **future enhancement/out of scope**.

**Implementasi**
ERD menjadi sumber kebenaran untuk kontrak backend. Nilai `question_type` yang diterima adalah `text`, `textarea`, `number`, `date`, `select`, `radio`, dan `checkbox`. Istilah `dropdown` pada dokumen kebutuhan dipetakan ke `select`; `file_upload` tetap berada di luar scope karena tidak tersedia pada ERD.

Migration tambahan yang sebelumnya mengubah enum telah dihapus sehingga CHECK constraint dari migration awal tetap menjadi acuan.

Seluruh respons endpoint mengikuti JSON envelope standar: `success`, `message`, `data`, dan `errors`.

### B. Dukungan Conditional Question

**Referensi Dokumen**
* **EVG-30** hanya menjelaskan atribut `is_required` (required/optional) tanpa membahas Conditional Question.
* **EVG-26 – URD** juga hanya menyebut atribut minimum (label, tipe jawaban, required/optional, urutan tampil, daftar pilihan jawaban) tanpa menjelaskan mekanisme Conditional Question.
* **EVG-34 – ERD** sudah menyediakan struktur database berupa: `requirement_type = wajib | opsional | kondisional`, `depends_on_question_id`, dan `depends_on_value`.

**Implementasi**
ERD telah menyediakan atribut `requirement_type`, `depends_on_question_id`, dan `depends_on_value` sebagai bagian dari skema Dynamic Question. Oleh karena itu implementasi API turut mendukung fitur tersebut meskipun Task Brief belum menjelaskan alur bisnis (business flow) Conditional Question secara rinci. Implementasi backend menyediakan dukungan pada level skema dan API. Mekanisme penampilan maupun penyembunyian pertanyaan berdasarkan kondisi dijalankan oleh aplikasi frontend menggunakan atribut `depends_on_question_id` dan `depends_on_value` sebagaimana dijelaskan pada bagian *Frontend Integration Guide*.

### C. Ownership Dynamic Question

**Referensi Dokumen**
* **EVG-30** menyatakan bahwa Admin Panitia hanya dapat membuat, mengubah, dan menghapus pertanyaan pada **form miliknya sendiri**.
* **Role & Permission Matrix (EVG-27)** menetapkan bahwa penyusunan dynamic form hanya berlaku untuk **Event Sendiri**, sedangkan Super Admin memiliki hak **override**.
* **URD (EVG-26)** juga menyatakan Admin Panitia mengelola event miliknya sendiri serta menyusun dynamic form untuk event tersebut.

**Implementasi**
API Create, Update, dan Delete Question mengunci hak akses menggunakan validasi ownership:
* **Admin Panitia** hanya dapat mengelola pertanyaan pada event yang dimilikinya.
* **Super Admin** memiliki hak khusus (override) untuk mengelola seluruh dynamic question milik siapapun.

Ownership divalidasi berdasarkan `organizer_id` dari event sebelum operasi Create, Update, maupun Delete dijalankan.

### D. Catatan Sinkronisasi Dokumen

Selama implementasi ditemukan beberapa perbedaan antara Task Brief, URD, dan ERD, khususnya terkait tipe `question_type` dan dukungan Conditional Question. Implementasi mengikuti keputusan yang digunakan pada task pengembangan saat ini. Dokumentasi sebaiknya disinkronkan pada revisi berikutnya agar requirement, ERD, dan implementasi menggunakan definisi yang konsisten.

---

## 8. Design Decisions

* Menggunakan database transaction pada Create dan Update untuk menjaga konsistensi antara `DynamicQuestion` dan `QuestionOption`.
* Menggunakan Replace-All Strategy pada Update `QuestionOption` untuk menyederhanakan sinkronisasi perubahan daftar opsi dan menghindari inkonsistensi antara payload terbaru dengan data yang tersimpan.
* Menggunakan Logical Deletion (soft delete berbasis flag) (`is_active = false`) agar data historis tetap tersedia.
* Menggunakan ownership validation agar Admin Panitia hanya dapat mengelola event miliknya sesuai Role Permission Matrix.
