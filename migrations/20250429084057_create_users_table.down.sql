CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name TEXT,
    last_name TEXT,
    patronymic TEXT,
    email TEXT,
    phone TEXT NOT NULL UNIQUE,
    is_email_verified BOOLEAN DEFAULT FALSE,
    is_phone_verified BOOLEAN DEFAULT FALSE,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);