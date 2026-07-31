# Technical Documentation: EVG-40 [BE1] Database Migration Setup Based on ERD Draft

---

## Document Metadata

| Attribute         | Details                                                      |
| :---------------- | :----------------------------------------------------------- |
| **Project**       | EventGate — Event Management System                          |
| **Task Code**     | EVG-40 (Sprint 2)                                            |
| **Feature**       | Database Migration Setup Based on ERD Draft                  |
| **Tech Stack**    | Golang (Go), PostgreSQL, SQL DDL                             |
| **Author**        | Backend Development Team                                     |
| **Status**        | Completed / Ready for Review                                 |
| **Reference Doc** | `docs/EVG-34_Data_Entity_Relationship.md` & `docs/image.png` |

---

## 1. Executive Summary

Task **EVG-40** mengimplementasikan fondasi database migration awal untuk repositori backend `be-eventgate` berdasarkan spesifikasi ERD Draft (EVG-34 Revisi 3). Pengerjaan ini mencakup DDL migration script murni PostgreSQL, seed data peran pengguna sistem (_roles_), perkakas CLI migration Go runner (`cmd/migrate/main.go`), serta index dan foreign key constraints untuk mendukung 14 entitas utama Sprint 2 (_User_, _Role_, _Event_, _Approval Log_, _Dynamic Form_, _Ticket_, _Registration_, _Payment_, _Check-in_, dan _Staff Assignment_).

---

## 2. Directory Layout & Migration Structure

```text
be-eventgate/
├── cmd/
│   └── migrate/
│       └── main.go                             # Go Migration CLI Runner (up / down)
├── config/
│   ├── config.go                               # Reader environment variables (.env)
│   └── database.go                             # PostgreSQL DB connection pool
├── docs/
│   ├── EVG-34_Data_Entity_Relationship.md      # Referensi ERD Draft
│   ├── EVG-38_Backend_Setup.md                 # Dokumentasi backend foundation
│   ├── EVG-40_Database_Migration_Setup.md      # Dokumentasi teknis task ini
│   └── image.png                               # Visual diagram ERD dbdiagram.io
├── migrations/
│   ├── 000001_create_initial_schema.up.sql     # DDL DDL 14 tabel & indexes (Up)
│   ├── 000001_create_initial_schema.down.sql   # Drop 14 tabel CASCADE (Down)
│   ├── 000002_seed_initial_roles.up.sql        # Seed data 4 role awal (Up)
│   └── 000002_seed_initial_roles.down.sql      # Delete seed data role (Down)
├── .env.example
├── .env
├── main.go
└── README.md
```

---

## 3. Database Schema Overview (14 Tables)

### 3.1 `roles`

Mengelola lookup tabel peran pengguna sistem.

- **Primary Key**: `role_id` (SERIAL)
- **Columns**: `role_name` (VARCHAR(50), UNIQUE, NOT NULL), `description` (TEXT)
- **Initial Seed Data**:
  - `super_admin`: Kendali penuh platform & global reviewer event.
  - `admin_panitia`: Organizer/EO internal, mengelola Event milik sendiri.
  - `staf_lapangan`: Petugas operasional scan QR & check-in lokasi.
  - `school_reviewer`: Role eksplisit opsional pemisah wewenang approval.

### 3.2 `users`

Menyimpan pengguna terintegrasi akses login dan soft delete.

- **Primary Key**: `user_id` (SERIAL)
- **Foreign Keys**: `role_id` -> `roles(role_id)` ON DELETE RESTRICT
- **Unique Constraints**: `username`, `email`
- **Audit Columns**: `is_active`, `last_login_at`, `created_at`, `updated_at`, `deleted_at`

### 3.3 `events`

Informasi utama event, lifecycle audit, dan URL slug.

- **Primary Key**: `event_id` (SERIAL)
- **Foreign Keys**: `organizer_id` -> `users(user_id)`, `created_by` -> `users(user_id)`, `updated_by` -> `users(user_id)`
- **Unique Constraint**: `slug`
- **Status Enum**: `draft`, `pending_approval`, `approved`, `revision_requested`, `published`, `rejected`, `cancelled`, `completed`

### 3.4 `event_approval_logs`

Mencatat seluruh histori pengajuan dan pengulasan event per versi.

- **Primary Key**: `log_id` (SERIAL)
- **Foreign Keys**: `event_id` -> `events(event_id)` ON DELETE CASCADE, `submitted_by` -> `users(user_id)`, `reviewed_by` -> `users(user_id)`
- **Action Enum**: `submitted`, `approved`, `rejected`, `revision_requested`

### 3.5 `dynamic_questions`

Pertanyaan dinamis form pendaftaran per event (_Form Builder_).

- **Primary Key**: `question_id` (SERIAL)
- **Foreign Keys**: `event_id` -> `events(event_id)` ON DELETE CASCADE, `depends_on_question_id` -> `dynamic_questions(question_id)` ON DELETE SET NULL
- **Type Enum**: `text`, `textarea`, `number`, `date`, `select`, `radio`, `checkbox`
- **Requirement Enum**: `wajib`, `opsional`, `kondisional`

### 3.6 `question_options`

Pilihan opsi jawaban untuk pertanyaan bertipe `select`, `radio`, atau `checkbox`.

- **Primary Key**: `option_id` (SERIAL)
- **Foreign Keys**: `question_id` -> `dynamic_questions(question_id)` ON DELETE CASCADE

### 3.7 `ticket_types`

Kategori dan kuota tiket custom per event.

- **Primary Key**: `ticket_type_id` (SERIAL)
- **Foreign Keys**: `event_id` -> `events(event_id)` ON DELETE CASCADE
- **Check Constraints**: `price >= 0`, `max_capacity >= 0`, `sold_count >= 0 AND sold_count <= max_capacity`

### 3.8 `participants`

Data profil peserta tanpa login.

- **Primary Key**: `participant_id` (SERIAL)
- **Unique Constraint**: `UNIQUE(email, name)` (memungkinkan satu email orang tua untuk beberapa anak)

### 3.9 `registrations`

Transaksi pendaftaran peserta ke event tertentu.

- **Primary Key**: `registration_id` (SERIAL)
- **Foreign Keys**: `participant_id` -> `participants(participant_id)`, `event_id` -> `events(event_id)`, `ticket_type_id` -> `ticket_types(ticket_type_id)`
- **Unique Constraint**: `registration_code`
- **Status Enum**: `pending`, `pending_payment`, `confirmed`, `cancelled`

### 3.10 `form_responses`

Jawaban dinamis peserta yang disimpan dalam JSON string-array (`JSONB`).

- **Primary Key**: `response_id` (SERIAL)
- **Foreign Keys**: `registration_id` -> `registrations(registration_id)` ON DELETE CASCADE, `question_id` -> `dynamic_questions(question_id)` ON DELETE CASCADE
- **Unique Constraint**: `UNIQUE(registration_id, question_id)`

### 3.11 `payments`

Percobaan pembayaran peserta (1:N Payment Attempt).

- **Primary Key**: `payment_id` (SERIAL)
- **Foreign Keys**: `registration_id` -> `registrations(registration_id)` ON DELETE CASCADE, `confirmed_by` -> `users(user_id)`
- **Payment Method Enum**: `manual_transfer`, `e_wallet`, `qris`, `virtual_account`, `gateway_other`
- **Status Enum**: `pending`, `paid`, `failed`, `refunded`, `expired`

### 3.12 `tickets`

Bukti tiket digital peserta berisi token QR unik bertipe UUID v4.

- **Primary Key**: `ticket_id` (SERIAL)
- **Foreign Keys**: `registration_id` -> `registrations(registration_id)` ON DELETE CASCADE, `ticket_type_id` -> `ticket_types(ticket_type_id)`
- **Unique Constraints**: `ticket_code`, `qr_code` (UUID), `registration_id` (1:1 Registrasi ke Tiket)
- **Status Enum**: `active`, `used`, `expired`, `cancelled`

### 3.13 `staff_assignments`

Penugasan staf lapangan ke event tertentu untuk otorisasi scan QR.

- **Primary Key**: `assignment_id` (SERIAL)
- **Foreign Keys**: `event_id` -> `events(event_id)` ON DELETE CASCADE, `user_id` -> `users(user_id)` ON DELETE CASCADE, `assigned_by` -> `users(user_id)`
- **Unique Constraint**: `UNIQUE(event_id, user_id)`
- **Status Enum**: `active`, `inactive`, `revoked`

### 3.14 `check_in_logs`

Catatan audit riwayat pemindaian QR code di pintu masuk.

- **Primary Key**: `log_id` (SERIAL)
- **Foreign Keys**: `event_id` -> `events(event_id)`, `ticket_id` -> `tickets(ticket_id)` ON DELETE SET NULL (nullable jika tiket tidak valid), `assignment_id` -> `staff_assignments(assignment_id)`, `checked_by` -> `users(user_id)`
- **Status Enum**: `success`, `failed`
- **Failure Reason Enum**: `invalid_ticket`, `already_used`, `expired`, `cancelled`, `event_mismatch`, `unauthorized_staff`

---

## 4. Local Execution & Verification Guide

### Cara Menjalankan Migration:

1. **Pastikan PostgreSQL Berjalan dan Konfigurasi `.env` Sesuai**:

   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=eventgate_db
   DB_SSLMODE=disable
   ```

2. **Eksekusi Migration UP (Create Schema & Seed Roles)**:

   ```powershell
   go run cmd/migrate/main.go up
   ```

3. **Eksekusi Migration DOWN (Rollback Schema)**:

   ```powershell
   go run cmd/migrate/main.go down
   ```

4. **Verifikasi Hasil Migration**:
   Buka terminal psql atau PostgreSQL GUI client (DBeaver / PgAdmin / TablePlus):
   ```sql
   SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';
   SELECT * FROM roles;
   ```
