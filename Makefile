.PHONY: wire swagger dev build

DB_MIGRATION_PATH = resources/db/migrations

DB_DRIVER = postgres
DB_USERNAME = root
DB_PASSWORD = vcVBncknc5FBmxrJ
DB_HOST = postgres-db.pve.lab
DB_PORT = 5432
DB_NAME = selah_dev
MIGRATION_MSG ?= change_with_actual_message

wire:
	go tool wire ./di

swagger:
	go tool swag init -g cmd/web/main.go -o ./docs --parseInternal --parseDependency

create-migration:
	go tool migrate create -ext sql -dir $(DB_MIGRATION_PATH) $(MIGRATION_MSG)

apply-migration:
	go tool migrate -database "$(DB_DRIVER)://$(DB_USERNAME):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?charset=utf8mb4&parseTime=True&loc=Local" -path $(DB_MIGRATION_PATH) up

dev: wire swagger
	go tool air

build: wire swagger
	go build -o ./bin/main ./cmd/web
