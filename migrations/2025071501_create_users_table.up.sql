CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    patronymic TEXT,
    email TEXT NOT NULL UNIQUE,
    phone TEXT NOT NULL UNIQUE,
    is_email_verified BOOLEAN DEFAULT FALSE,
    is_phone_verified BOOLEAN DEFAULT FALSE,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    referrer_id INT REFERENCES users(id) ON DELETE SET NULL,
    card_number TEXT,
    balance NUMERIC(12, 2) DEFAULT 0,
    role user_role NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT now()
);