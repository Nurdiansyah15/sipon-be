# Infra Reference — Design Reference

Referensi infrastruktur **generik** untuk backend modular monolith: acuan
`docker-compose.yml` untuk service-level infra (dev), dan konvensi folder
`infrastructure/` di dalam setiap module untuk code-level infra. Ini
adalah pola/template, bukan dokumentasi satu project tertentu — nama
service, nama module, dan nilai env var di bawah ini bersifat ilustratif.
Untuk desain module secara umum, lihat
[modular-monolith-design.md](./modular-monolith-design.md).

## Dua lapis "infra"

1. **Infra deployment** — service eksternal (database, cache, broker
   pesan, object storage) yang dijalankan sebagai container, didefinisikan
   di `docker-compose.yml` + `Dockerfile`. Ini yang dimaksud "infra" dalam
   arti DevOps.
2. **Infra kode per module** — folder `infrastructure/` di dalam setiap
   `internal/modules/<nama>/`, tempat implementasi konkret (query
   database, HTTP client pihak ketiga, gateway ke module lain) diletakkan,
   sebagai lawan dari `domain/` dan `application/` yang murni tanpa
   dependency eksternal. Ini yang dimaksud "infra" dalam arti hexagonal
   architecture.

Keduanya saling terhubung: container database di compose adalah tempat
`infrastructure/persistence/*` di semua module menyimpan data; container
broker pesan adalah tempat `infrastructure/rabbitmq/*` (atau setara)
publish/consume.

---

## 1. Infra deployment — `docker-compose.yml` (contoh)

### Inventaris service (pola umum)

| Service | Peran | Healthcheck | Profile |
|---|---|---|---|
| `postgres` (atau DB lain) | Database utama, satu instance untuk semua module (skema per module dibedakan lewat penamaan tabel, bukan database terpisah) | `pg_isready` | default |
| `redis` | Cache + revocation token (module identity), rate limiter | `redis-cli ping` | default |
| `rabbitmq` (atau broker lain) | Broker untuk pola outbox (module messaging) — kalau dipakai, wajib aktif karena worker bergantung padanya | `rabbitmq-diagnostics ping` | default |
| `minio` (atau S3-compatible lain) | Object storage untuk module penyimpanan dokumen/aset dan module lain yang upload file | `mc ready local` | default |
| `<storage>-init` | One-shot: bikin bucket, set policy | - | `tooling` |
| `app` | HTTP server (`cmd/api`) | - | default |
| `worker` | Job terjadwal + outbox relay + message consumer (`cmd/worker`) | expose `/healthz`, `/metrics` | default |
| `migrate` | One-shot migrasi DB (`cmd/migrate`) | - | `tooling` |
| `seeder` | One-shot seed data awal | - | `tooling` |
| `devtools` | Generate dokumentasi API, lint, tooling lain | - | `tooling` |

Semua service satu network bridge sendiri. Volume persisten untuk tiap
data store (`postgres_data`, `redis_data`, `rabbitmq_data`, object storage
data), plus cache dependency/build (mis. `go_mod_cache`/`go_build_cache`
untuk Go) yang dishare `app`/`worker`/`migrate`/`seeder`/`devtools` supaya
tidak download/build ulang tiap start.

Catatan desain:
- Di dev, `app` dan `worker` bisa dijalankan langsung dari source
  (mis. `go run ./cmd/...`) di atas image toolchain resmi dengan source
  code di-bind-mount, supaya hot-reload cukup dengan restart container,
  tanpa rebuild image tiap ganti kode. Untuk image production, build
  binary lewat multi-stage `Dockerfile` (lihat di bawah).
- `migrate` dan `seeder` sebaiknya diberi `profiles: [tooling]` — tidak
  ikut naik otomatis saat `docker compose up`, harus dipanggil eksplisit
  lewat `docker compose run --rm migrate ...`. Ini karena keduanya
  one-shot command, bukan long-running service.
- Kalau memakai pola outbox (lihat modular-monolith-design.md), `worker`
  harus `depends_on` broker pesan, bukan cuma database+cache.

### Alur command (dev)

```
make dev-up          → naikkan dependency minimal (db + cache), tunggu healthy
make dev-all-up       → naikkan semua service, termasuk app+worker di container
make run              → jalankan app di foreground (buat lihat log)
make migrate-up        → jalankan migrasi lewat container migrate
make seed-all           → jalankan seed data lewat container seeder
make lint / gen-docs     → jalankan tooling lewat container devtools
```

Sediakan target terpisah untuk "naikkan dependency saja" (kalau ingin
jalankan app/worker di host langsung, mis. untuk debugging dengan
debugger lokal) vs "naikkan semuanya di container".

### `Dockerfile` (untuk build image production)

Pola multi-stage yang umum dipakai:

```
builder   → compile semua entrypoint (cmd/api, cmd/migrate, dst.) jadi binary statis
devtools  → base + tooling dev (linter, generator dokumentasi API)
app       → base minimal + binary api + file migrasi (image final utk cmd/api)
migrate   → base minimal + binary migrate + file migrasi (image final utk migrasi di CI/CD)
```

Kalau ada entrypoint lain yang perlu di-deploy sebagai image terpisah
(mis. worker), tambahkan stage baru dengan pola yang sama seperti
`migrate`: compile binary-nya di stage `builder`, lalu stage runtime
minimal yang cuma copy binary tersebut.

### Environment variables (kelompok umum)

Semua konfigurasi lewat env var, di-load jadi satu struct config dari
`.env`. Kelompok yang biasanya muncul di backend modular monolith:

| Prefix (contoh) | Untuk | Catatan |
|---|---|---|
| `APP_*` | Port, env, format log, timezone platform | dipakai semua entrypoint |
| `DB_*` | Koneksi database | pakai nama service (mis. `postgres`) sebagai host di dalam docker network |
| `JWT_*` / `SESSION_*` | Secret + TTL token/session | module identity |
| `<mail/sms-provider>_*` | Email & SMS/WhatsApp gateway pihak ketiga | dipakai module notifikasi |
| `REDIS_*` | Cache/rate-limit/revocation | alamat internal beda dari port expose ke host |
| `<broker>_*` | Broker pesan untuk pola outbox | kalau dipakai, wajib `ENABLED=true` — worker sebaiknya exit kalau tidak |
| `<push-notification>_*` | Push notification pihak ketiga | kosongkan untuk nonaktifkan |
| `RATE_LIMIT_*` | Rate limit per-IP/per-user/per-endpoint sensitif | dipakai middleware rate-limit |
| `<object-storage>_*` | Object storage | endpoint internal (dipakai dari dalam container) vs endpoint publik (dipakai browser/klien) sebaiknya dipisah karena diakses dari sisi berbeda |
| `WORKER_*` | Tick/lease scheduler, health port proses worker | khusus `cmd/worker` |

Port di compose sebaiknya selalu punya dua angka: port **internal**
(dipakai antar container di network yang sama) dan port **expose** ke
host — supaya beberapa instance dev di mesin yang sama tidak rebutan port
host, tapi port internal tetap konsisten.

---

## 2. Infra kode per module — folder `infrastructure/`

Di dalam tiap `internal/modules/<nama>/infrastructure/`, isinya selalu
salah satu dari tiga kategori berikut (lihat juga layering lengkap di
[modular-monolith-design.md](./modular-monolith-design.md#struktur-satu-module)):

| Subfolder | Isi | Contoh pola penamaan |
|---|---|---|
| `persistence/` | Implementasi `application/ports.Repository` ke database | `identity/infrastructure/persistence` |
| `<nama>gateway/` | Adapter yang memanggil `Contract` module LAIN, mengimplementasikan port milik module ini sendiri (dependency inversion — lihat `module-boundaries.md`) | `ordering/infrastructure/catalogGateway`, `billing/infrastructure/identityGateway` |
| `external/` | Adapter ke pihak ketiga di luar sistem sendiri (bukan module lain) | `identity/infrastructure/external` (email/SMS/OAuth provider), `notification/infrastructure/external` (push provider) |
| lainnya (spesifik) | Adapter yang cukup spesifik sehingga tidak masuk 3 kategori di atas | `identity/infrastructure/cache` (Redis), `messaging/infrastructure/rabbitmq` |

Konvensi penamaan `<nama>gateway` (bukan generik `gateway/` untuk semua):
kalau satu module memanggil lebih dari satu module lain, masing-masing
punya foldernya sendiri — supaya jelas dari nama folder module mana saja
yang benar-benar dipanggil, tanpa perlu buka isi file.

### Contoh peta gateway antar-module (ilustratif)

```
ordering    → catalogGateway  → catalog.Contract
            → identityGateway → identity.Contract
billing     → catalogGateway  → catalog.Contract
media       → identityGateway → identity.Contract
```

Kalau menambah panggilan lintas-module baru, tambahkan folder
`infrastructure/<tujuan>gateway/` baru di module pemanggil — jangan
menaruh call ke `Contract` module lain langsung di `application/`
(harus lewat port + adapter, lihat aturan di `module-boundaries.md`).

### Migrasi database

Semua file migrasi flat di `migrations/` (bukan dipecah per module),
dengan prefix timestamp, dijalankan lewat entrypoint migrasi khusus. Nama
file biasanya diawali nama module (`create_<module>_tables`) supaya tetap
bisa ditelusuri module mana yang punya tabel apa, meski fisiknya tidak
dipisah folder.

### Menambah service infra baru (mis. cache khusus, broker lain)

1. Tambahkan service baru di `docker-compose.yml`, ikuti pola yang sudah
   ada: `healthcheck`, logging dengan rotasi ukuran, masuk network yang
   sama, env var lewat `${VAR:-default}` supaya bisa dioverride tanpa
   ubah file compose.
2. Tambahkan config struct baru di layer config + entry baru di
   `.env.example` dengan komentar penjelasan (jelaskan *kenapa*, bukan
   cuma *apa*).
3. Buat client/koneksinya di module yang butuh, ditaruh di
   `infrastructure/external/` (kalau pihak ketiga) — jangan taruh setup
   koneksi generik di `internal/shared/` kecuali benar-benar dipakai
   banyak module tanpa opini domain (lihat aturan `shared/` di
   `modular-monolith-design.md`).
