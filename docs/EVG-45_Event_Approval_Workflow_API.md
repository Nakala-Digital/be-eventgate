# Technical Documentation: EVG-45 [BE3] Event Approval Workflow API

---

## 1. Document Metadata

| Attribute         | Details                                                      |
| :---------------- | :----------------------------------------------------------- |
| **Project**       | EventGate — Event Management System                          |
| **Task Code**     | EVG-45                                                       |
| **Feature**       | Event Approval Workflow & State Machine                      |
| **Tech Stack**    | Golang (Go), PostgreSQL, GORM                                |
| **Author**        | Backend Development Team                                     |
| **Status**        | Completed / Ready for Review                                 |

---

## 2. Executive Summary

Task **EVG-45** memperkenalkan sistem alur persetujuan bertingkat (*Hierarchical Approval Workflow*) untuk pengelolaan acara (Event). Sebelumnya, sebuah event yang dibuat dapat langsung terlihat atau dikelola tanpa pengawasan. Dengan pembaruan ini, setiap acara baru yang dibuat secara otomatis akan mendapatkan status `draft` dan harus melalui siklus peninjauan (pengajuan, persetujuan, revisi, rilis) sebelum dapat dipublikasikan ke publik. Sistem ini juga mencakup audit trail (`EventApprovalLog`) yang mencatat secara historis semua transisi status acara.

---

## 3. State Machine & Business Rules

### A. Peta Transisi Status (State Machine)
Siklus hidup sebuah acara secara ketat mengikuti aturan perpindahan status berikut:

*   `draft` ➔ `pending_approval` (Penyelenggara mengajukan rilis)
*   `pending_approval` ➔ `approved` | `rejected` | `revision_requested` (Keputusan oleh Peninjau/Super Admin)
*   `revision_requested` ➔ `pending_approval` (Penyelenggara mengirim ulang setelah revisi)
*   `approved` ➔ `published` (Super Admin mempublikasikan acara secara resmi)
*   `published` ➔ `draft` (Acara ditarik kembali / dibatalkan tayang)
*   `rejected` ➔ *(Status final, tidak ada transisi lanjutan)*

### B. Aturan Bisnis Kunci
1.  **Validasi Tiket:** Acara (Event) tidak dapat diajukan (`submit`) jika belum memiliki minimal satu jenis tiket (`TicketType`).
2.  **No Self-Approval:** Penyelenggara acara (Admin Panitia) secara sistematis tidak diizinkan untuk menyetujui acaranya sendiri. Peninjauan harus dilakukan oleh pengguna lain dengan hak akses `super_admin` atau `school_reviewer`.
3.  **Kewajiban Catatan (Notes):** Jika peninjau menolak (`reject`) atau meminta revisi (`request-revision`), mereka wajib menyertakan alasan penolakan/revisi pada atribut `notes`.
4.  **Audit Trail (Versioning):** Versi event (`event_version`) akan dinaikkan nilainya (+1) setiap kali penyelenggara menekan tombol submit (pergeseran ke `pending_approval`). Segala aksi rilis/tolak direkam rapi di tabel `event_approval_logs`.

---

## 4. Frontend Integration Guide

Panduan bagi tim Frontend untuk berinteraksi dengan API persetujuan:

### A. Manajemen Tombol Aksi (Action Buttons)
Frontend dapat menggunakan atribut `status` dari objek Event dan `role_name` dari user (didapatkan dari `/api/auth/me`) untuk menampilkan tombol yang tepat:
*   Jika User = Pemilik Event (Admin Panitia) & Status = `draft` atau `revision_requested`: **Tampilkan tombol "Submit for Review"**.
*   Jika User = Super Admin / Reviewer & Status = `pending_approval`: **Tampilkan tombol "Approve", "Reject", "Request Revision"**.
*   Jika User = Super Admin & Status = `approved`: **Tampilkan tombol "Publish"**.

### B. Daftar Log Persetujuan
Riwayat persetujuan dapat diambil melalui *endpoint* Approval Logs. Rekomendasi UX: Tampilkan dalam bentuk garis waktu (*Timeline*) untuk memperlihatkan historis pergerakan event dari revisi hingga rilis.

---

## 5. API Reference

Berikut spesifikasi *endpoint* baru yang dapat dikonsumsi. 
*Semua endpoint mewajibkan HTTP Header `Authorization: Bearer <TOKEN>`.*

### 5.1. Get Approval Logs
Mengambil histori log persetujuan dari suatu acara (diurutkan kronologis).
*   **URL:** `GET /api/events/{id}/approval-logs`
*   **Akses:** Terbatas untuk Super Admin, Reviewer, atau Pemilik Event. Staf Lapangan tidak memiliki akses (403 Forbidden).

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "event_id": 1,
    "event_version": 1,
    "action": "submitted",
    "submitted_by": { "id": 1, "username": "John Doe", "email": "john@example.com" },
    "submitted_at": "2026-07-25T10:00:00Z",
    "notes": "Mohon ditinjau proposal kami"
  },
  {
    "id": 2,
    "event_id": 1,
    "event_version": 1,
    "action": "approved",
    "reviewed_by": { "id": 2, "username": "Admin Kampus", "email": "admin@example.com" },
    "reviewed_at": "2026-07-25T11:00:00Z",
    "notes": "Sesuai kriteria"
  }
]
```

### 5.2. Submit Event For Approval
*   **URL:** `POST /api/events/{id}/submit`
*   **Akses:** Hanya Pemilik Event (Admin Panitia).
*   **Body (Opsional):** `{ "notes": "Pengajuan revisi ke-2" }`

### 5.3. Review Event (Approve, Reject, Request Revision)
*   **URL (Approve):** `POST /api/events/{id}/approve`
*   **URL (Reject):** `POST /api/events/{id}/reject`
*   **URL (Revision):** `POST /api/events/{id}/request-revision`
*   **Akses:** Super Admin, School Reviewer (Kecuali pemilik event itu sendiri).
*   **Body:** `{ "notes": "Catatan wajib untuk reject dan revisi" }`

### 5.4. Publish / Unpublish Event
*   **URL (Publish):** `POST /api/events/{id}/publish`
*   **URL (Unpublish):** `POST /api/events/{id}/unpublish`
*   **Akses:** Super Admin.
*   **Body (Opsional):** `{ "notes": "Catatan" }`

*(Respons dari seluruh endpoint aksi (5.2 - 5.4) akan mengembalikan objek JSON detail event dengan atribut `status` yang telah diperbarui).*

---

## 6. Architecture & Security

*   **Decoupled State Machine:** Logika perpindahan status tidak dicampur dengan logika HTTP / Database, melainkan dikelompokkan pada *package* murni `internal/eventapproval`. Hal ini diatur agar alur logika dapat diuji (Unit Test) sepenuhnya terisolasi dari basis data.
*   **ACID Transaction (Database):** Pembaruan `status` di tabel `events` dan penambahan *record* baru pada tabel `event_approval_logs` diikat dalam satu transaksi relasional (`tx.Transaction` GORM). Hal ini menjamin bahwa jika penulisan log gagal, maka perubahan status event akan dibatalkan (*rollback*) agar data tidak terpisah.

---

## 7. Directory Layout & Modules

```text
be-eventgate/
├── docs/
│   └── EVG-45_Event_Approval_Workflow_API.md      # Dokumen ini
├── migrations/
│   └── 000004_add_event_version_to_events.up.sql  # Modifikasi skema DB
├── internal/
│   ├── eventapproval/
│   │   ├── transitions.go                         # Mesin Statis Transisi Status
│   │   └── transitions_test.go                    # Unit test logika transisi
│   ├── handlers/
│   │   ├── event_dto.go                           # Tipe Request/Response baru
│   │   ├── event_approval_handler.go              # Komponen HTTP Handler persetujuan
│   │   └── event_approval_handler_test.go         # Test untuk Handler
│   └── models/
│       ├── event.go                               # Modifikasi atribut Event
│       ├── event_approval_log.go                  # Skema Audit Trail log
│       └── ticket_type.go                         # Skema Jenis Tiket minimal
```
