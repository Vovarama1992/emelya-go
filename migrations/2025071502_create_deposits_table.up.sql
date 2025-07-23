CREATE TYPE deposit_status AS ENUM (
    'pending',
    'approved',
    'closed'
);

CREATE TABLE deposits (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    amount NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_at TIMESTAMPTZ,
    block_until TIMESTAMPTZ,
    daily_reward NUMERIC(12, 2),
    status deposit_status NOT NULL DEFAULT 'pending'
);