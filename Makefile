.PHONY: wire swagger generate dev build create-migration apply-migration rollback-migration fmt format fmt-check test test-coverage

CONFIG_FILE_PATH = config.yaml
DB_MIGRATION_PATH = infra/db/migrations

DB_DRIVER := postgres
DB_USERNAME := $(shell yq '.database.username' $(CONFIG_FILE_PATH))
DB_PASSWORD := $(shell yq '.database.password' $(CONFIG_FILE_PATH))
DB_HOST := $(shell yq '.database.host' $(CONFIG_FILE_PATH))
DB_PORT := $(shell yq '.database.port' $(CONFIG_FILE_PATH))
DB_NAME := $(shell yq '.database.name' $(CONFIG_FILE_PATH))
MIGRATION_MSG ?= change_with_actual_message

MIGRATE = go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
GOLINES ?= go tool golines

fmt:
	go fmt ./...
	$(GOLINES) -m 80 -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Unformatted files (gofmt):" && gofmt -l . && exit 1)
	@test -z "$$($(GOLINES) -l .)" || (echo "Unformatted files (golines):" && $(GOLINES) -l . && exit 1)

wire:
	go tool wire ./di

swagger:
	go tool swag init -d ./cmd/web,./internal/delivery/http -g main.go -o ./docs --parseInternal --parseDependency --packagePrefix github.com/bryanfks-dev/logos --v3.1

generate: wire swagger

create-migration:
	$(MIGRATE) create -ext sql -dir $(DB_MIGRATION_PATH) $(MIGRATION_MSG)

apply-migration:
	$(MIGRATE) -database "$(DB_DRIVER)://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" -path $(DB_MIGRATION_PATH) up

rollback-migration:
	$(MIGRATE) -database "$(DB_DRIVER)://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" -path $(DB_MIGRATION_PATH) down 1

dev: wire swagger
	go tool air

build: wire swagger
	go build -o ./bin/main ./cmd/web
	go build -o ./bin/worker ./cmd/worker

test:
	go test -v -count=1 ./...

test-coverage:
	go test -cover ./...
