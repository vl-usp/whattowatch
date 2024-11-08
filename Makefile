include .env.default
LOCAL_BIN=$(CURDIR)/bin
NAME=default
MIGRATION_DIR=$(CURDIR)/migration

.PHONY: build install-deps bot load migration-create migration-status migration-up migration-down docker-up 

build:
	go build -o $(LOCAL_BIN)/loader cmd/loader/main.go
	go build -o $(LOCAL_BIN)/bot cmd/bot/main.go

bot: build
	$(LOCAL_BIN)/bot

load: build
	$(LOCAL_BIN)/loader

install-deps:
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.20.0

migration-create:
	$(LOCAL_BIN)/goose -dir $(MIGRATION_DIR) create $(NAME) sql

migration-status:
	$(LOCAL_BIN)/goose -dir $(MIGRATION_DIR) postgres $(POSTGRES_DSN) status -v

migration-up:
	$(LOCAL_BIN)/goose -dir $(MIGRATION_DIR) postgres $(POSTGRES_DSN) up -v

migration-down:
	$(LOCAL_BIN)/goose -dir $(MIGRATION_DIR) postgres $(POSTGRES_DSN) down -v
	
migration-down-all:
	$(LOCAL_BIN)/goose -dir $(MIGRATION_DIR) postgres $(POSTGRES_DSN) down-to 0 -v

migration-reset: migration-down-all migration-up

docker-up:
	mkdir -p .tmp
	docker compose build --progress plain &> .tmp/build/docker.log
	docker compose up -d
