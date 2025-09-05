#!/bin/bash
set -e

echo "Starting local DB setup for devtesting..."

# load Environment Variables
# ------------------------------------
if [ -f ./../../../devtest-db/.env ]; then
  export $(grep -v '^#' ./../../../devtest-db/.env | xargs)
  echo "Loaded configuration from .env file."
else
  echo "Error: .env file not found. Please review your workspace."
  exit 1
fi

# check for the init.sql file
if [ ! -f ./../../../devtest-db/init.sql ]; then
    echo "Error: init.sql file not found in the current directory."
    exit 1
fi

# start PostgreSQL server
# ------------------------------------
echo "Checking PostgreSQL server status..."
if ! brew services list | grep -E "postgresql(@\d+)?\s+started" -q; then
    echo "-> Postgres server is not running. Starting it now..."
    brew services start postgresql
    
    echo "-> Waiting for PostgreSQL to accept connections..."
    # Use pg_isready to wait until the server is available
    until pg_isready -q; do
        sleep 1
    done
    echo "PostgreSQL is up and running."
else
    echo "PostgreSQL server is already running."
fi

# Create user role if it does not exist
# ------------------------------------
# This connects using the default OS user, which is a superuser by default on Homebrew installs.
echo "Checking for role '$POSTGRES_USER'..."
ROLE_EXISTS=$(psql -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$POSTGRES_USER'")

if [ "$ROLE_EXISTS" = "1" ]; then
    echo "-> Role '$POSTGRES_USER' already exists."
else
    echo "-> Role '$POSTGRES_USER' not found. Creating it now..."
    createuser --superuser "$POSTGRES_USER"
    echo "-> Role created successfully."
fi


# recreate the Database and Run Init Script
# ------------------------------------
echo "Checking for database '$POSTGRES_DB' existence..."
DB_EXISTS=$(psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$POSTGRES_DB'")

if [ "$DB_EXISTS" != "1" ]; then
    echo "-> Database '$POSTGRES_DB' not found. Creating it..."
    createdb -h "$POSTGRES_HOST" -U "$POSTGRES_USER" "$POSTGRES_DB"
    echo "-> New empty database created."
else
    echo "-> Database '$POSTGRES_DB' already exists."
fi

# execute the init.sql script against the new database
psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f ./../../../devtest-db/init.sql
echo "Successfully executed init.sql to set up schema and data."


echo "DB setup complete!"
