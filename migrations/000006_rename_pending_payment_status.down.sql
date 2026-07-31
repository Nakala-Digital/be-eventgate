UPDATE registrations SET status = 'pending_payment' WHERE status = 'waiting_payment';

ALTER TABLE registrations DROP CONSTRAINT IF EXISTS registrations_status_check;

ALTER TABLE registrations ADD CONSTRAINT registrations_status_check
    CHECK (status IN ('pending', 'pending_payment', 'confirmed', 'cancelled'));
