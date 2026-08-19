.PHONY: help dev-up dev-down dev-all-up dev-all-down run minio-init migrate-up migrate-down migrate-fresh migrate-version migrate-force migrate-create seed seed-all seed-role seed-user build tidy test test-unit test-integration test-usecase lint swagger swagger-check

# ── Help ──────────────────────────────────────────────────────────────────────
help:
	@echo ""
	@echo "  sipon-be — perintah yang tersedia:"
	@echo ""
	@echo "  Development:"
	@echo "    make dev-up         Jalankan postgres + redis (tunggu healthy)"
	@echo "    make dev-all-up     Jalankan semua container dev (tunggu healthy)"
	@echo "    make dev-down       Hentikan semua container dev"
	@echo "    make dev-all-down   Hentikan semua container dev dan hapus volume"
	@echo "    make run            Jalankan HTTP server via container app"
	@echo "    make swagger        Generate Swagger dari dalam container devtools"
	@echo "    make swagger-check  Validasi konfigurasi docker compose dev"
	@echo "    make logs            Tampilkan log semua container dev (tail 100, follow)"
	@echo "    make minio-init     Inisialisasi bucket MinIO (via container minio-init)"
	@echo ""
	@echo "  Migrasi (via container migrate):"
	@echo "    make migrate-up     Jalankan semua migrasi yang belum dijalankan"
	@echo "    make migrate-down   Rollback satu migrasi terakhir"
	@echo "    make migrate-fresh  Reset DB (drop all) lalu jalankan semua migrasi dari awal"
	@echo "    make migrate-version Cek versi migrasi saat ini"
	@echo "    make migrate-force  Force ke versi tertentu"
	@echo "    make migrate-create NAME=nama_migrasi  Buat file migrasi baru (prefix nomor urut)"
	@echo ""
	@echo "  Seeder (via container seeder):"
	@echo "    make seed-all       Jalankan semua seeder"
	@echo "    make seed NAME=role Jalankan seeder tertentu"
	@echo "    make seed-role      Shortcut untuk role seeder"
	@echo "    make seed-user      Shortcut untuk user seeder"
	@echo ""
	@echo "  Build & Test:"
	@echo "    make build          Build image devtools via docker"
	@echo "    make tidy           go mod tidy via container devtools"
	@echo "    make test           Jalankan semua test"
	@echo "    make test-unit      Jalankan domain unit tests (tanpa DB)"
	@echo "    make test-integration Jalankan persistence integration tests (butuh Docker)"
	@echo "    make test-usecase   Jalankan application/use case tests (mock-based)"
	@echo ""

# ── Development ───────────────────────────────────────────────────────────────
dev-up:
	docker compose -f docker-compose.dev.yml up -d --wait postgres redis
	@echo "postgres + redis berjalan dan healthy"

dev-down:
	docker compose -f docker-compose.dev.yml down

dev-all-up:
	docker compose -f docker-compose.dev.yml down
	docker compose -f docker-compose.dev.yml up -d --wait

# logs
logs:
	docker compose -f docker-compose.dev.yml logs -f --tail=100
# 	make minio-init

dev-all-down:
	docker compose -f docker-compose.dev.yml down -v

minio-init:
	docker compose -f docker-compose.dev.yml run --rm minio-init

run:
	docker compose -f docker-compose.dev.yml up app

# ── Migrasi ───────────────────────────────────────────────────────────────────
migrate-up:
	docker compose -f docker-compose.dev.yml run --rm migrate up

migrate-down:
	docker compose -f docker-compose.dev.yml run --rm migrate down 1

migrate-fresh:
	docker compose -f docker-compose.dev.yml run --rm migrate fresh

migrate-version:
	docker compose -f docker-compose.dev.yml run --rm migrate version

migrate-force:
	@read -p "Force ke versi: " v; \
	docker compose -f docker-compose.dev.yml run --rm migrate force $$v

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=nama_migrasi"; \
		exit 1; \
	fi; \
	DIR=migrations; \
	TS=$$(date +%Y%m%d%H%M%S); \
	UP=$$DIR/$${TS}_$(NAME).up.sql; \
	DOWN=$$DIR/$${TS}_$(NAME).down.sql; \
	touch $$UP $$DOWN; \
	echo "Dibuat:"; \
	echo "  $$UP"; \
	echo "  $$DOWN"

# ── Seeder ────────────────────────────────────────────────────────────────────
seed-all:
	docker compose -f docker-compose.dev.yml run --rm seeder all

seed:
	@if [ -z "$(NAME)" ]; then \
		docker compose -f docker-compose.dev.yml run --rm seeder all; \
	else \
		docker compose -f docker-compose.dev.yml run --rm seeder $(NAME); \
	fi

seed-role:
	docker compose -f docker-compose.dev.yml run --rm seeder role

seed-user:
	docker compose -f docker-compose.dev.yml run --rm seeder user

# ── Build ─────────────────────────────────────────────────────────────────────
build:
	docker compose -f docker-compose.dev.yml build devtools
	@echo "Image devtools berhasil di-build"

# ── Lainnya ───────────────────────────────────────────────────────────────────
tidy:
	docker compose -f docker-compose.dev.yml run --rm --no-deps devtools go mod tidy

test:
	go test ./... -v -count=1 -timeout 180s

test-unit:
	go test ./internal/modules/.../domain/... -v -count=1 -short

test-integration:
	go test ./internal/modules/.../infrastructure/persistence/... -v -count=1 -timeout 120s

test-usecase:
	go test ./internal/modules/.../application/... -v -count=1

lint:
	docker compose -f docker-compose.dev.yml run --rm --no-deps devtools golangci-lint run ./...

swagger:
	docker compose -f docker-compose.dev.yml run --rm --no-deps -u $$(id -u):$$(id -g) devtools swag init -g cmd/api/main.go -o docs

swagger-check:
	docker compose -f docker-compose.dev.yml config >/dev/null
