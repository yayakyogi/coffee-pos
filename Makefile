DB_URL=mysql://root:secret@tcp(localhost:3306)/coffee_pos
MIGRATIONS_PATH=migrations

.PHONY: migrate-up migrate-down migrate-down-all migrate-version migrate-create run tidy

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

migrate-down-all:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down -all

migrate-version:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

run:
	go run cmd/api/main.go

tidy:
	go mod tidy
