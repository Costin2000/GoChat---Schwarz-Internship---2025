CREATE TABLE IF NOT EXISTS "User" (
    id BIGSERIAL PRIMARY KEY, -- BIGSERIAL instead of TEXT to ensure automatic id increment without race conditions
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    user_name TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);