.PHONY: wire swagger dev build

wire:
	go tool wire ./di

swagger:
	go tool swag init -g cmd/web/main.go -o ./docs --parseInternal --parseDependency

dev: wire swagger
	go tool air

build: wire swagger
	go build -o ./bin/main ./cmd/web
