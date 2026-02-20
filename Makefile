postgres:
	podman run --network inventium --name postgres  --hostname posrgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres:16-alpine
createdb:
	podman exec -it postgres createdb --username=root --owner=root inventory-service
dropdb:
	podman exec -it postgres dropdb --username=root inventory-service
migrateup:
	migrate -path ./models/migration -database "$(DB_SOURCE)" -verbose up
migratedown:
	migrate -path ./models/migration -database "$(DB_SOURCE)" -verbose down
sqlc:
	sqlc generate --no-remote
loaddata:
	PGPASSWORD=secret psql -h localhost -U root -d inventory-service -f data/sql/inventium.sql
runcontainer:
	podman run --network inventium --name inventory-service -p 13740:13740 -d -e DB_SOURCE="$(DB_SOURCE)" -e CLERK_KEY="$(CLERK_KEY)" -e SERVICE_NAME="$(SERVICE_NAME)" -e OTEL_EXPORTER_OTLP_ENDPOINT="$(OTEL_EXPORTER_OTLP_ENDPOINT) -e LOKI_URL="$(LOKI_URL)" -e LOG_FILE_PATH="$(LOG_FILE_PATH)" inventory-service:test
.PHONY: postgres createdb dropdb migrateup migratedown sqlc loaddata runcontainer

# Docker/Podman multi-platform build targets
build:
	podman build -t inventory-service:latest .

# Build for specific platform
build-amd64:
	podman build --platform linux/amd64 -t inventory-service:latest-amd64 .

build-arm64:
	podman build --platform linux/arm64 -t inventory-service:latest-arm64 .

# Build for multiple platforms using buildx
build-multi:
	podman build --platform linux/amd64,linux/arm64 -t inventory-service:latest .

# Build and push multi-platform image to registry
build-push-multi:
	podman build --platform linux/amd64,linux/arm64 \
		-t your-registry/inventory-service:latest \
		--push .

# Create buildx builder (run once)
buildx-setup:
	docker buildx create --name multiplatform --use
	docker buildx inspect --bootstrap

.PHONY: build build-amd64 build-arm64 build-multi build-push-multi buildx-setup
