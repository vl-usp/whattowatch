#!/bin/bash
source .env

sleep 2 && goose -dir "${MIGRATION_DIR}" postgres "${POSTGRES_DSN}" up -v%