BEGIN;

-- Testing database - every run is a clean run to ensure safety when introducing or removing dummy data

-- drop tables and types in order of dependency to avoid foreign key constraint errors

DROP TABLE IF EXISTS "Message";
DROP TABLE IF EXISTS "Friend Requests";
DROP TABLE IF EXISTS "Conversation";
DROP TABLE IF EXISTS "User";

DROP TYPE IF EXISTS FRIEND_REQUEST_STATUS;


CREATE TABLE IF NOT EXISTS "User" (
    id BIGSERIAL PRIMARY KEY, -- BIGSERIAL instead of TEXT to ensure automatic id increment without race conditions
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    user_name TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE FRIEND_REQUEST_STATUS AS ENUM ('pending', 'accepted', 'rejected', 'blocked');

CREATE TABLE IF NOT EXISTS "Friend Requests" (
    id BIGSERIAL PRIMARY KEY,
    sender_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    receiver_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    status FRIEND_REQUEST_STATUS NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (sender_id <> receiver_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS FRIEND_REQUEST_ORDER_IDX 
ON "Friend Requests" (LEAST(sender_id, receiver_id), GREATEST(sender_id, receiver_id));

CREATE TABLE IF NOT EXISTS "Conversation" (
    id BIGSERIAL PRIMARY KEY,
    user1_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    user2_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK (user1_id <> user2_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS CONVERSATION_USER_ORDER_IDX 
ON "Conversation" (LEAST(user1_id, user2_id), GREATEST(user1_id, user2_id));

CREATE TABLE IF NOT EXISTS "Message" (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES "Conversation"(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS message_conv_created_idx 
ON "Message"(conversation_id, created_at);

COMMIT;
