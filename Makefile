PATH = cmd/api

SRC = $(PATH)/main.go $(PATH)/config.go

all:
	go run $(SRC)

test:
	go test ./... -cover

startdb:
	docker-compose up

stopdb:
	docker-compose down

.PHONY: all test startdb stopdb
