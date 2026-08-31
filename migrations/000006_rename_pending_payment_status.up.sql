-- EVG-49: task brief pakai istilah 'waiting_payment', tapi CHECK constraint
-- asli di migrations/000001 memakai 'pending_payment'. Rename nilai +
-- constraint-nya supaya konsisten dengan API.
--
-- CATATAN: nama constraint diasumsikan default Postgres
-- (registrations_status_check). Kalau beda di project Anda, cek dulu lewat
-- \d registrations di psql dan sesuaikan baris DROP CONSTRAINT.

UPDATE registrations SET status = 'waiting_payment' WHERE status = 'pending_payment';

ALTER TABLE registrations DROP CONSTRAINT IF EXISTS registrations_status_check;

ALTER TABLE registrations ADD CONSTRAINT registrations_status_check
    CHECK (status IN ('pending', 'waiting_payment', 'confirmed', 'cancelled'));
