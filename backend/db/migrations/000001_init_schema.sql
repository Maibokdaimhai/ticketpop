-- =============================================================
-- TicketPop Database Schema
-- PostgreSQL
-- =============================================================

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================
-- LOOKUP TABLES
-- =============================================================

CREATE TABLE account_status (
    id     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL UNIQUE  -- 'active', 'suspended', 'banned'
);

CREATE TABLE event_categories (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE  -- e.g. 'concert', 'sport', 'theater'
);

-- =============================================================
-- USERS & AUTH
-- =============================================================

CREATE TABLE users (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email             VARCHAR(100) NOT NULL UNIQUE,
    username          VARCHAR(50)  NOT NULL UNIQUE,
    display_name      TEXT,
    pfp_url           TEXT,                          -- profile picture URL, NULL = use default
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ,                   -- soft delete: NULL = active
    account_status_id UUID        NOT NULL REFERENCES account_status(id)
);

CREATE TABLE user_credentials (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    password_hash   VARCHAR(255) NOT NULL,           -- bcrypt / argon2 hash เท่านั้น ห้าม plain text
    last_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id         UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE
);

-- =============================================================
-- ORGANIZERS
-- =============================================================

CREATE TABLE organizers (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT        NOT NULL,
    bio           TEXT,
    contact_email TEXT,                              -- อีเมลติดต่อผู้จัด อาจต่างจาก users.email
    logo_url      TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id       UUID        NOT NULL REFERENCES users(id)
);

-- =============================================================
-- EVENTS
-- =============================================================

CREATE TABLE events (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title        VARCHAR(255) NOT NULL,
    description  TEXT,
    place        TEXT,
    show_time    TIMESTAMPTZ  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,                        -- soft delete
    img_url      TEXT,
    organizer_id UUID         NOT NULL REFERENCES organizers(id),
    category_id  UUID         REFERENCES event_categories(id)
);

-- =============================================================
-- TICKET POOLS (โซน/ประเภทตั๋วของแต่ละ event)
-- =============================================================

CREATE TABLE ticket_pools (
    id                 UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_name          VARCHAR(50)   NOT NULL,        -- e.g. 'VIP', 'Gold', 'Silver'
    price              NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    total_capacity     INT           NOT NULL CHECK (total_capacity > 0),
    available_capacity INT           NOT NULL CHECK (available_capacity >= 0),
    display_name       TEXT,                          -- ชื่อแสดงผล e.g. 'VIP Front Row'
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ   NOT NULL DEFAULT now(),
    version            INT           NOT NULL DEFAULT 1, -- optimistic locking สำหรับ sync Redis → DB
    event_id           UUID          NOT NULL REFERENCES events(id),

    CONSTRAINT available_lte_total CHECK (available_capacity <= total_capacity)
);

-- =============================================================
-- BOOKINGS
-- =============================================================

CREATE TABLE bookings (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    status      VARCHAR(20)   NOT NULL DEFAULT 'pending', -- 'pending', 'confirmed', 'cancelled', 'expired'
    total_price NUMERIC(10,2) NOT NULL CHECK (total_price >= 0),
    expires_at  TIMESTAMPTZ   NOT NULL,              -- booking หมดอายุถ้ายังไม่ชำระ เช่น +15 นาที
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    user_id     UUID          NOT NULL REFERENCES users(id),
    event_id    UUID          NOT NULL REFERENCES events(id)
);

-- =============================================================
-- TICKETS (ตั๋วแต่ละใบ)
-- =============================================================

CREATE TABLE tickets (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    seat_number    VARCHAR(30),                      -- NULL ได้ถ้าเป็น general admission
    status         VARCHAR(20) NOT NULL DEFAULT 'reserved', -- 'reserved', 'confirmed', 'used', 'cancelled', 'refunded'
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    ticket_pool_id UUID        NOT NULL REFERENCES ticket_pools(id),
    booking_id     UUID        NOT NULL REFERENCES bookings(id)
);

-- =============================================================
-- TICKET STATUS LOG (audit trail)
-- =============================================================

CREATE TABLE ticket_status_log (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    old_status VARCHAR(50) NOT NULL,
    new_status VARCHAR(50) NOT NULL,
    reason     TEXT,                                 -- e.g. 'payment confirmed', 'booking expired'
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ticket_id  UUID        NOT NULL REFERENCES tickets(id),
    changed_by UUID        REFERENCES users(id)      -- NULL = system process
);

-- =============================================================
-- PAYMENT LEDGERS
-- =============================================================

CREATE TABLE payment_ledgers (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_type    VARCHAR(20)   NOT NULL,      -- 'charge', 'refund', 'partial_refund'
    amount              NUMERIC(10,2) NOT NULL,
    payment_gateway_ref VARCHAR(100),                -- reference ID จาก Omise/Stripe เช่น 'ch_3OtXY2...'
    status              VARCHAR(20)   NOT NULL DEFAULT 'pending', -- 'pending', 'success', 'failed'
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT now(),
    booking_id          UUID          NOT NULL REFERENCES bookings(id)
);

-- =============================================================
-- REFUNDS
-- =============================================================

CREATE TABLE refunds (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    amount            NUMERIC(10,2) NOT NULL CHECK (amount > 0),
    status            VARCHAR(20)   NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'rejected', 'completed'
    reason            TEXT,
    gateway_ref       VARCHAR(100),                 -- refund reference จาก gateway
    requested_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),
    resolved_at       TIMESTAMPTZ,                  -- NULL = ยังไม่ได้ resolve
    payment_ledger_id UUID          NOT NULL REFERENCES payment_ledgers(id),
    booking_id        UUID          NOT NULL REFERENCES bookings(id)
);

-- =============================================================
-- INDEXES
-- =============================================================

-- users
CREATE INDEX idx_users_email          ON users(email);
CREATE INDEX idx_users_deleted_at     ON users(deleted_at) WHERE deleted_at IS NULL;

-- events
CREATE INDEX idx_events_show_time     ON events(show_time);
CREATE INDEX idx_events_organizer_id  ON events(organizer_id);
CREATE INDEX idx_events_category_id   ON events(category_id);
CREATE INDEX idx_events_deleted_at    ON events(deleted_at) WHERE deleted_at IS NULL;

-- ticket_pools
CREATE INDEX idx_ticket_pools_event_id ON ticket_pools(event_id);

-- bookings
CREATE INDEX idx_bookings_user_id     ON bookings(user_id);
CREATE INDEX idx_bookings_event_id    ON bookings(event_id);
CREATE INDEX idx_bookings_status      ON bookings(status);
CREATE INDEX idx_bookings_expires_at  ON bookings(expires_at) WHERE status = 'pending';

-- tickets
CREATE INDEX idx_tickets_booking_id      ON tickets(booking_id);
CREATE INDEX idx_tickets_ticket_pool_id  ON tickets(ticket_pool_id);

-- payment_ledgers
CREATE INDEX idx_payment_ledgers_booking_id ON payment_ledgers(booking_id);

-- ticket_status_log
CREATE INDEX idx_ticket_status_log_ticket_id ON ticket_status_log(ticket_id);

-- =============================================================
-- SEED: account_status
-- =============================================================

INSERT INTO account_status (id, status) VALUES
    (gen_random_uuid(), 'active'),
    (gen_random_uuid(), 'suspended'),
    (gen_random_uuid(), 'banned');
