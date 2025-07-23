CREATE TABLE tariffs (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    block_until TIMESTAMPTZ,
    daily_reward NUMERIC(12, 2),
    created_at TIMESTAMPTZ DEFAULT now()
);