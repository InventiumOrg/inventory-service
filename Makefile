postgres:
	podman run --name postgres-1 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres:16-alpine
createdb:
	podman exec -it postgres-1 createdb --username=root --owner=root simple_bank
dropdb:
	podman exec -it postgres-1 dropdb --username=root simple_bank
migrateup:
	migrate -path ./models/migration -database "$DB_SOURCE" -verbose up
migratedown:
	migrate -path ./models/migration -database "$DB_SOURCE" -verbose down
sqlc:
	sqlc generate --no-remote
.PHONY: postgres createdb dropdb migrateup migratedown sqlc