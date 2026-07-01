-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE public.user_role AS ENUM ('admin', 'user');
CREATE TYPE public.account_status AS ENUM ('pending', 'approved', 'banned');

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,
    username TEXT UNIQUE,
    password TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,

    description TEXT DEFAULT '',
    banner_image TEXT DEFAULT '',
    image TEXT DEFAULT '',
    batch TEXT DEFAULT '',

    social_links JSONB NOT NULL DEFAULT '{}'::jsonb,

    status public.account_status NOT NULL DEFAULT 'pending',
    role public.user_role NOT NULL DEFAULT 'user',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS public.account_status;
DROP TYPE IF EXISTS public.user_role;
