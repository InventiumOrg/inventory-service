postgres:
	podman run --network inventium --name postgres-1  --hostname posrgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres:16-alpine
createdb:
	podman exec -it postgres-1 createdb --username=root --owner=root inventory-service
dropdb:
	podman exec -it postgres-1 dropdb --username=root inventory-service
migrateup:
	migrate -path ./models/migration -database "postgresql://root:secret@localhost:5432/inventory-service?sslmode=disable" -verbose up
migratedown:
	migrate -path ./models/migration -database "postgresql://root:secret@localhost:5432/inventory-service?sslmode=disable" -verbose down
sqlc:
	sqlc generate --no-remote
loaddata:
	PGPASSWORD=secret psql -h localhost -U root -d inventory-service -f data/sql/inventium.sql
runcontainer:
	podman run --network inventium --name inventory-service -p 13740:13740 -d -e DB_SOURCE="postgresql://root:secret@postgres-1:5432/inventory-service?sslmode=disable" -e CLERK_KEY="sk_test_XhHg2KNAIqm9I65JwOgQbLajZj6UqeeLTnpjx1p4oa" inventory-service:1.0.0
.PHONY: postgres createdb dropdb migrateup migratedown sqlc loaddata runcontainer