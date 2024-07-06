#!/bin/bash
echo "Migrate DB: $DB_NAME on host: ${DB_HOST}:${DB_PORT}"
export DSN_QUEUE="user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME host=$DB_HOST port=$DB_PORT sslmode=disable"
goose -dir="food-diary-db/" postgres "$DSN_QUEUE" up