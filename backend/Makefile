export DB_HOST=127.0.0.1
export DB_USER=postgres
export DB_PASSWORD=1234
export DB_NAME=bookstore
export DB_PORT=8989

dev:
	air

seed:
	go run ./cmd/seed

sqlc_generate:
	sqlc generate

goose_up:
	goose -dir ./sql/schema postgres "host=$$DB_HOST user=$$DB_USER password=$$DB_PASSWORD dbname=$$DB_NAME port=$$DB_PORT" up

test:
	go test ./... -v
