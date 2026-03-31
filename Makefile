include .env
export

# Database URL untuk migrasi
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: run migrate-up migrate-down seed db-shell test clean

# 1. Jalankan Aplikasi
run:
	@echo "🚀 Starting Server..."
	go run ./cmd/api/main.go

# 2. Migrasi Database Manual (Jika dibutuhkan)
migrate-up:
	@echo "⬆️ Running Migrations..."
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	@echo "⬇️ Rolling back Migrations..."
	migrate -path ./migrations -database "$(DB_URL)" down 1

# 3. Seeding Data Awal (Data Dummy untuk Testing)
# Pastikan Anda sudah buat file scripts/seed.sql
seed:
	@echo "🌱 Seeding initial data..."
	psql -d $(DB_NAME) -f scripts/seed.sql

# 4. Buka Database CLI
db-shell:
	psql -d $(DB_NAME)

# 5. Testing
test:
	@echo "🧪 Running Tests..."
	go test -v ./...

# 6. Bersihkan Binary & Cache
clean:
	@echo "🧹 Cleaning up..."
	go clean
	rm -f main