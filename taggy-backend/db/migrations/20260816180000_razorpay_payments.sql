-- +goose Up
CREATE TYPE payment_status AS ENUM (
    'CREATED',
    'PAID',
    'FAILED'
);

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    razorpay_order_id TEXT NOT NULL UNIQUE,
    razorpay_payment_id TEXT UNIQUE,
    amount INTEGER NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    status payment_status NOT NULL DEFAULT 'CREATED',
    receipt VARCHAR(40) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX payments_user_id_created_idx ON payments (user_id, created_at DESC);
CREATE INDEX payments_status_idx ON payments (status);

-- +goose Down
DROP TABLE IF EXISTS payments;
DROP TYPE IF EXISTS payment_status;
