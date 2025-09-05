#!/bin/bash
set -e

# stop the PostgreSQL service
echo "-> Stopping PostgreSQL service..."
brew services stop postgresql
echo "PostgreSQL service stopped."