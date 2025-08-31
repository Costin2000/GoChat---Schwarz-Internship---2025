#!/bin/bash
set -e

# 1. Set DB vars
DB_USER="postgres"
DB_PASSWORD="postgres"
DB_NAME="gochat"
DB_PORT="5432"

# 2. Wait for DB
echo "Waiting for PostgreSQL..."
until pg_isready -h localhost -p "$DB_PORT" -U "$DB_USER"; do
  sleep 1
done

# 3. Create Alice
ALICE_JSON=$(cat <<EOF
{
  "user": {
    "first_name": "Alice",
    "last_name": "Example",
    "user_name": "alice",
    "email": "alice@example.com",
    "password": "password123"
  }
}
EOF
)

ALICE_ID=$(curl -s -X POST http://localhost:8080/v1/user \
  -H "Content-Type: application/json" \
  -d "$ALICE_JSON" | jq -r '.user.id')

echo "Alice created with ID: $ALICE_ID"

# 4. Create Bob
BOB_JSON=$(cat <<EOF
{
  "user": {
    "first_name": "Bob",
    "last_name": "Example",
    "user_name": "bob",
    "email": "bob@example.com",
    "password": "password123"
  }
}
EOF
)

BOB_ID=$(curl -s -X POST http://localhost:8080/v1/user \
  -H "Content-Type: application/json" \
  -d "$BOB_JSON" | jq -r '.user.id')

echo "Bob created with ID: $BOB_ID"

# 5. Create Friend Request (Alice -> Bob)
FR_JSON=$(cat <<EOF
{
  "sender_id": "$ALICE_ID",
  "receiver_id": "$BOB_ID"
}
EOF
)

FR_ID=$(curl -s -X POST http://localhost:8080/v1/friend-request \
  -H "Content-Type: application/json" \
  -d "$FR_JSON" | jq -r '.request.id')

echo "Friend request created with ID: $FR_ID"

# 6. Update Friend Request (accept)
UPDATE_JSON=$(cat <<EOF
{
  "friend_request": {
    "id": "$FR_ID",
    "status": "STATUS_ACCEPTED"
  },
  "field_mask": {
    "paths": ["status"]
  }
}
EOF
)

UPDATED_FR=$(grpcurl -plaintext -d "$UPDATE_JSON" localhost:50052 friendrequest.FriendRequestService/UpdateFriendRequest)

echo "Updated Friend Request:"
echo "$UPDATED_FR"