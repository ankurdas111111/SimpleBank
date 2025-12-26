DB_NAME ?= simple_bank
DB_USER ?= root
DB_PASSWORD ?= secret
DB_HOST ?= localhost
DB_PORT ?= 5400
DB_SSLMODE ?= disable
DB_URL ?= postgresql://root:secret@localhost:5400/simple_bank?sslmode=disable

APP_IMAGE ?= simplebank:latest
APP_CONTAINER ?= simplebank
APP_PORT ?= 8080
APP_GIN_MODE ?= release

postgres:  
	@if docker ps --format '{{.Names}}' | grep -q '^postgres12$$'; then \
		echo "postgres12 already running"; \
	elif docker ps -a --format '{{.Names}}' | grep -q '^postgres12$$'; then \
		docker start postgres12; \
	else \
		docker run --name postgres12 -p 5400:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine; \
	fi

createdb:
	@docker exec postgres12 psql -U $(DB_USER) -tAc "SELECT 1 FROM pg_database WHERE datname='$(DB_NAME)'" | grep -q 1 || \
		docker exec postgres12 createdb --username=$(DB_USER) --owner=$(DB_USER) $(DB_NAME)

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1


migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

dropdb:
	docker exec postgres12 dropdb --username=$(DB_USER) $(DB_NAME)

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	 go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go -build_flags="-mod=mod" github.com/ankurdas111111/simplebank/db/sqlc Store

docker-build:
	docker build -t $(APP_IMAGE) .

docker-rm:
	-docker rm -f $(APP_CONTAINER)

docker-run:
	docker run -d --name $(APP_CONTAINER) \
		-p $(APP_PORT):8080 \
		-e GIN_MODE=$(APP_GIN_MODE) \
		-e DB_SOURCE="postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" \
		$(APP_IMAGE)

docker-logs:
	docker logs -f $(APP_CONTAINER)

.PHONY: createdb dropdb postgres migrateup migratedown migrateup1 migratedown1 sqlc test server mock docker-build docker-rm docker-run docker-logs

