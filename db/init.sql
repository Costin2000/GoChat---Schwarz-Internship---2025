CREATE TABLE User (
    id TEXT
    first_name TEXT NOT NULL
    last_name TEXT NOT NULL
    user_name TEXT NOT NULL
    email TEXT NOT NULL
    password TEXT NOT NULL
    created_at created_at TIMESTAMPTZ WITH TIME ZONE NOT NULL DEFAULT NOW()
)