-- drop the table if it exists to apply changes
DROP TABLE IF EXISTS "User";

CREATE TABLE "User" (
    id BIGSERIAL PRIMARY KEY, -- BIGSERIAL instead of TEXT to ensure automatic id increment without race conditions
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    user_name TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--  dummy entry for testing
INSERT INTO "User" (first_name, last_name, user_name, email, password) VALUES ('dummy', 'user', 'testuser', 'test@example.com', 'password');