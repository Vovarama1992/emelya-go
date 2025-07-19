CREATE TYPE deposit_status AS ENUM (
    'pending',
    'approved',
    'closed'
);

CREATE TYPE tariftype AS ENUM ('Легкий старт', 'Триумф', 'Максимум');

CREATE TABLE deposits (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    amount NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_at TIMESTAMPTZ,
    block_until TIMESTAMPTZ,
    tarif tariftype,
    daily_reward NUMERIC(12, 2),
    status deposit_status NOT NULL DEFAULT 'pending'
);