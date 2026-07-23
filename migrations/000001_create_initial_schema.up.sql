-- EVG-40: Initial Database Schema Migration (Up)
-- Based on EVG-34 Data Entity & Relationship Draft Revision 3

-- 1. Role Table
CREATE TABLE IF NOT EXISTS roles (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

-- 2. User Table
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    role_id INT NOT NULL REFERENCES roles(role_id) ON DELETE RESTRICT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 3. Event Table
CREATE TABLE IF NOT EXISTS events (
    event_id SERIAL PRIMARY KEY,
    organizer_id INT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255) NOT NULL,
    banner_url VARCHAR(255),
    slug VARCHAR(255) NOT NULL UNIQUE,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_approval', 'approved', 'revision_requested', 'published', 'rejected', 'cancelled', 'completed')),
    published_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_by INT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    updated_by INT REFERENCES users(user_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 4. Event Approval Log Table
CREATE TABLE IF NOT EXISTS event_approval_logs (
    log_id SERIAL PRIMARY KEY,
    event_id INT NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    event_version INT NOT NULL DEFAULT 1,
    action VARCHAR(50) NOT NULL CHECK (action IN ('submitted', 'approved', 'rejected', 'revision_requested')),
    submitted_by INT REFERENCES users(user_id) ON DELETE SET NULL,
    submitted_at TIMESTAMPTZ,
    reviewed_by INT REFERENCES users(user_id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    notes TEXT
);

-- 5. Dynamic Question Table
CREATE TABLE IF NOT EXISTS dynamic_questions (
    question_id SERIAL PRIMARY KEY,
    event_id INT NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    question_type VARCHAR(50) NOT NULL CHECK (question_type IN ('text', 'textarea', 'number', 'date', 'select', 'radio', 'checkbox')),
    requirement_type VARCHAR(50) NOT NULL DEFAULT 'wajib' CHECK (requirement_type IN ('wajib', 'opsional', 'kondisional')),
    depends_on_question_id INT REFERENCES dynamic_questions(question_id) ON DELETE SET NULL,
    depends_on_value VARCHAR(255),
    validation_rule VARCHAR(255),
    placeholder VARCHAR(255),
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 6. Question Option Table
CREATE TABLE IF NOT EXISTS question_options (
    option_id SERIAL PRIMARY KEY,
    question_id INT NOT NULL REFERENCES dynamic_questions(question_id) ON DELETE CASCADE,
    option_label VARCHAR(255) NOT NULL,
    option_value VARCHAR(255) NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- 7. Ticket Type Table
CREATE TABLE IF NOT EXISTS ticket_types (
    ticket_type_id SERIAL PRIMARY KEY,
    event_id INT NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,
    price NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (price >= 0),
    max_capacity INT NOT NULL CHECK (max_capacity >= 0),
    sold_count INT NOT NULL DEFAULT 0 CHECK (sold_count >= 0 AND sold_count <= max_capacity),
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 8. Participant Table
CREATE TABLE IF NOT EXISTS participants (
    participant_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(100) NOT NULL,
    phone_number VARCHAR(50),
    student_id VARCHAR(50),
    class_name VARCHAR(50),
    guardian_name VARCHAR(255),
    institution_unit VARCHAR(255),
    CONSTRAINT unique_participant_email_name UNIQUE (email, name)
);

-- 9. Registration Table
CREATE TABLE IF NOT EXISTS registrations (
    registration_id SERIAL PRIMARY KEY,
    registration_code VARCHAR(100) NOT NULL UNIQUE,
    participant_id INT NOT NULL REFERENCES participants(participant_id) ON DELETE RESTRICT,
    event_id INT NOT NULL REFERENCES events(event_id) ON DELETE RESTRICT,
    ticket_type_id INT NOT NULL REFERENCES ticket_types(ticket_type_id) ON DELETE RESTRICT,
    registration_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'pending_payment', 'confirmed', 'cancelled')),
    confirmed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancelled_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 10. Form Response Table
CREATE TABLE IF NOT EXISTS form_responses (
    response_id SERIAL PRIMARY KEY,
    registration_id INT NOT NULL REFERENCES registrations(registration_id) ON DELETE CASCADE,
    question_id INT NOT NULL REFERENCES dynamic_questions(question_id) ON DELETE CASCADE,
    answer_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_form_response UNIQUE (registration_id, question_id)
);

-- 11. Payment Table
CREATE TABLE IF NOT EXISTS payments (
    payment_id SERIAL PRIMARY KEY,
    registration_id INT NOT NULL REFERENCES registrations(registration_id) ON DELETE CASCADE,
    attempt_number INT NOT NULL DEFAULT 1,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    payment_method VARCHAR(50) NOT NULL CHECK (payment_method IN ('manual_transfer', 'e_wallet', 'qris', 'virtual_account', 'gateway_other')),
    gateway_transaction_id VARCHAR(255),
    payment_proof_url VARCHAR(255),
    reference_number VARCHAR(255),
    confirmed_by INT REFERENCES users(user_id) ON DELETE SET NULL,
    confirmed_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    failure_reason TEXT,
    refund_reason TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'failed', 'refunded', 'expired')),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 12. Ticket Table
CREATE TABLE IF NOT EXISTS tickets (
    ticket_id SERIAL PRIMARY KEY,
    registration_id INT NOT NULL UNIQUE REFERENCES registrations(registration_id) ON DELETE CASCADE,
    ticket_type_id INT NOT NULL REFERENCES ticket_types(ticket_type_id) ON DELETE RESTRICT,
    ticket_code VARCHAR(100) NOT NULL UNIQUE,
    qr_code UUID NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'used', 'expired', 'cancelled')),
    issued_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 13. Staff Assignment Table
CREATE TABLE IF NOT EXISTS staff_assignments (
    assignment_id SERIAL PRIMARY KEY,
    event_id INT NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    assigned_by INT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'revoked')),
    revoked_at TIMESTAMPTZ,
    revoked_by INT REFERENCES users(user_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_staff_assignment UNIQUE (event_id, user_id)
);

-- 14. Check-in Log Table
CREATE TABLE IF NOT EXISTS check_in_logs (
    log_id SERIAL PRIMARY KEY,
    event_id INT NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    ticket_id INT REFERENCES tickets(ticket_id) ON DELETE SET NULL,
    scanned_code_hash VARCHAR(255) NOT NULL,
    assignment_id INT REFERENCES staff_assignments(assignment_id) ON DELETE SET NULL,
    checked_by INT REFERENCES users(user_id) ON DELETE SET NULL,
    check_in_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL CHECK (status IN ('success', 'failed')),
    failure_reason VARCHAR(100) CHECK (failure_reason IS NULL OR failure_reason IN ('invalid_ticket', 'already_used', 'expired', 'cancelled', 'event_mismatch', 'unauthorized_staff')),
    notes TEXT,
    device_info VARCHAR(255)
);

-- Index Definitions
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_events_organizer_id ON events(organizer_id);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_events_slug ON events(slug);
CREATE INDEX IF NOT EXISTS idx_event_approval_logs_event_id ON event_approval_logs(event_id);
CREATE INDEX IF NOT EXISTS idx_dynamic_questions_event_id ON dynamic_questions(event_id);
CREATE INDEX IF NOT EXISTS idx_question_options_question_id ON question_options(question_id);
CREATE INDEX IF NOT EXISTS idx_ticket_types_event_id ON ticket_types(event_id);
CREATE INDEX IF NOT EXISTS idx_registrations_participant_id ON registrations(participant_id);
CREATE INDEX IF NOT EXISTS idx_registrations_event_id ON registrations(event_id);
CREATE INDEX IF NOT EXISTS idx_registrations_ticket_type_id ON registrations(ticket_type_id);
CREATE INDEX IF NOT EXISTS idx_registrations_status ON registrations(status);
CREATE INDEX IF NOT EXISTS idx_form_responses_registration_id ON form_responses(registration_id);
CREATE INDEX IF NOT EXISTS idx_form_responses_question_id ON form_responses(question_id);
CREATE INDEX IF NOT EXISTS idx_payments_registration_id ON payments(registration_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_tickets_registration_id ON tickets(registration_id);
CREATE INDEX IF NOT EXISTS idx_tickets_qr_code ON tickets(qr_code);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_staff_assignments_event_user ON staff_assignments(event_id, user_id);
CREATE INDEX IF NOT EXISTS idx_check_in_logs_event_id ON check_in_logs(event_id);
CREATE INDEX IF NOT EXISTS idx_check_in_logs_ticket_id ON check_in_logs(ticket_id);
CREATE INDEX IF NOT EXISTS idx_check_in_logs_status ON check_in_logs(status);
