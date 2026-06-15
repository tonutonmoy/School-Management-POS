.PHONY: run build migrate sqlc templ tailwind docker-up docker-down test tidy

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

migrate:
	goose -dir migrations postgres "$(DATABASE_URL)" up

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

templ:
	go run github.com/a-h/templ/cmd/templ@latest generate

tailwind:
	npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/app.css --minify

generate: templ sqlc tailwind

tidy:
	go mod tidy

test:
	go test ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
