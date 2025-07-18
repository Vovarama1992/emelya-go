CREATE TYPE reward_type AS ENUM ('deposit', 'referral');

CREATE TABLE rewards (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    deposit_id INT REFERENCES deposits(id),
    type reward_type NOT NULL,
    amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
    withdrawn NUMERIC(12, 2) NOT NULL DEFAULT 0,
    last_accrued_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);