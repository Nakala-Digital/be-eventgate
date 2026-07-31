-- EVG-45: kolom event_version untuk tracking versi pengajuan approval.
-- Semua kolom lain yang dibutuhkan EVG-45 (published_at, cancelled_at,
-- completed_at, created_by, updated_by di tabel events; seluruh tabel
-- event_approval_logs dan ticket_types) SUDAH ADA sejak
-- 000001_create_initial_schema.up.sql — TIDAK perlu migration tambahan
-- untuk itu.
ALTER TABLE events ADD COLUMN IF NOT EXISTS event_version INT NOT NULL DEFAULT 0;
