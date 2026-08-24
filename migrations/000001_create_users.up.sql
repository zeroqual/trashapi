CREATE EXTENSION IF NOT EXISTS citext;
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(25) NOT NULL,
    surname VARCHAR(25) NOT NULL,
    role VARCHAR(25) NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);