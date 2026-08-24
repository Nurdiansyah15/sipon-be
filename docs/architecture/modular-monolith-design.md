# Desain Modular Monolith — Design Reference

Dokumen ini adalah **referensi desain generik** untuk membangun backend
sebagai *modular monolith*: satu binary/deployable, tapi dipecah jadi
module-module yang terisolasi lewat batas kode (bukan lewat network/service
terpisah). Isinya pola dan aturan struktural — bukan dokumentasi satu
project tertentu — supaya bisa dipakai ulang sebagai acuan tiap kali
membangun/mengevaluasi backend dengan pendekatan yang sama.

Contoh nama module, folder, dan urutan wiring di bawah ini bersifat
**ilustratif** (`module-a`, `identity`, `billing`, dst.) — ganti dengan
domain module nyata di project masing-masing. Untuk aturan detail
komunikasi antar-module lihat [module-boundaries.md](./module-boundaries.md)
(pola `Contract`); untuk referensi infra lihat
[infra-reference.md](./infra-reference.md).

## Kenapa modular monolith

- Tim kecil–menengah, satu database, satu deploy unit — microservices
  menambah biaya operasional (network, observability, distributed
  transaction) yang belum sepadan di skala ini.
- Batas domain tetap dijaga tegas lewat `Contract` per module (lihat
  module-boundaries.md), supaya kalau suatu saat module tertentu memang
  perlu dipisah jadi service sendiri, batasnya sudah ada — bukan digali
  dari kode yang saling silang.
- Semua module compile jadi satu binary, jadi refactor lintas-module dan
  static typing bahasa pemrograman langsung menangkap kesalahan di compile
  time, bukan di runtime seperti pada microservices.

## Struktur folder

```
cmd/
  api/        entrypoint HTTP server — merakit semua module, mount route
  worker/     entrypoint background worker — job terjadwal + consumer pesan asinkron
  migrate/    entrypoint migrasi DB
  seeder/     entrypoint seed data awal
internal/
  modules/    satu folder per bounded context (lihat pola di bawah)
  seeders/    implementasi seeder, dipanggil dari cmd/seeder
  shared/     kode lintas-module yang benar-benar generik (lihat bawah)
migrations/   file migrasi, flat (tidak dipisah per module)
```

`cmd/` adalah **composition root** — satu-satunya tempat yang boleh
mengimpor banyak module sekaligus dan merakit dependency di antaranya
(lihat "Composition root" di bawah). Kode di dalam `internal/modules/*`
tidak pernah mengimpor `cmd/*`.

## Struktur satu module

Setiap module di `internal/modules/<nama-module>/` mengikuti layering
DDD/hexagonal yang sama:

```
<module>/
  module.go              composition root module: NewModule(...) merakit semuanya
  contract.go            (opsional) interface Contract — API lintas-module, lihat module-boundaries.md
  domain/
    <aggregate>/         satu subfolder per aggregate/entity (entity, value object, domain error)
  application/
    command/             use case yang mengubah state (write)
    query/                use case yang membaca state (read)
    ports/                interface yang dibutuhkan module ini dari luar (repo, gateway module lain, dst)
    dto/                  data transfer object untuk request/response use case
  infrastructure/
    persistence/          implementasi ports.Repository — akses database
    <x>gateway/            adapter yang mengimplementasikan ports milik module ini dengan
                           memanggil Contract module LAIN (lihat module-boundaries.md)
    external/              adapter ke pihak ketiga (email, SMS, storage, dst)
  interfaces/
    http/                  handler HTTP + DTO request/response + route registration
    mq/                     (opsional) consumer/handler pesan asinkron
```

Bukan semua module punya semua folder ini — module yang tidak menerima
request HTTP tidak punya `interfaces/http`; module dengan domain sederhana
tidak punya `application/ports` yang tebal. Konsistensi yang dijaga bukan
"semua folder harus ada", tapi **arah dependency**: `interfaces` →
`application` → `domain`, dan `infrastructure` mengimplementasikan
`application/ports` (dependency inversion) — tidak pernah sebaliknya.

## Contoh daftar module (ilustratif)

Tabel berikut contoh generik untuk menggambarkan bagaimana bounded context
dipecah dan bagaimana `Contract` dipakai — sesuaikan isinya dengan domain
project nyata:

| Module (contoh) | Tanggung jawab (generik) | Contract? |
|---|---|---|
| `identity` | Autentikasi, session/token, user management, role & permission | Ya |
| `catalog` | Data induk entitas utama domain (mis. produk/anggota/aset) | Ya |
| `ordering` / proses inti | Proses transaksional utama domain, mengorkestrasi module lain | Tidak (konsumer Contract module lain) |
| `billing` | Akuntansi/keuangan: invoice, payment, jurnal, periode | Ya |
| `media` / asset storage | Penyimpanan generik dokumen/file, dipakai module lain | Ya |
| `notification` | Notifikasi in-app & push, delivery tracking, preference | Ya |
| `messaging` | Infrastruktur outbox/queue internal untuk komunikasi asinkron antar-module | Ya |
| `scheduler` | Penjadwalan job (recurring/one-off) berbasis lease, dijalankan proses worker | Ya |

## Komunikasi antar-module

- **Sinkron (request-response dalam proses yang sama)**: lewat `Contract`
  milik module tujuan — lihat aturan lengkap di
  [module-boundaries.md](./module-boundaries.md). Ini jalur utama untuk
  kebutuhan seperti "ambil ringkasan user dari module identity" atau
  "buat entitas X berdasarkan hasil proses di module Y".
- **Asinkron (fire-and-forget, event-driven)**: lewat module `messaging` —
  pola **outbox**: module penulis event menyimpan baris ke tabel outbox
  dalam transaksi yang sama dengan perubahan data utamanya (konsistensi
  terjamin tanpa 2-phase commit), lalu proses worker menjalankan **Outbox
  Relay** yang mempublikasikan baris outbox ke broker pesan, dan **Message
  Consumer** yang menerima pesan dari broker dan memanggil handler module
  tujuan (diregistrasikan lewat mekanisme semacam
  `RegisterMessageHandlers`).
- **Scheduled job**: module `scheduler` menerima job (recurring/one-off)
  lewat Contract-nya dari module lain, lalu proses worker menjalankan
  dispatcher yang mengklaim job jatuh tempo (lease-based, aman untuk
  multi-instance worker) dan menulisnya ke outbox — jadi eksekusi job pun
  tetap lewat jalur outbox/broker yang sama, bukan jalur direct-call.

## Composition root (`cmd/api`, `cmd/worker`)

`main.go` di `cmd/api` dan `cmd/worker` adalah satu-satunya tempat semua
module diimpor dan dirakit lewat constructor injection biasa (`NewModule
(...)`), **dalam urutan dependency**: module yang tidak butuh module lain
dibuat duluan, lalu module yang menerima `Contract`-nya sebagai parameter
constructor. Contoh urutan ilustratif:

```
identity   (tidak depend module lain)
  → notification   (butuh identity middleware + identity sebagai user provider)
  → media           (butuh identity middleware)
    → catalog        (butuh identity.Contract + media.Contract)
      → ordering      (butuh identity.Contract + catalog.Contract)
      → billing        (butuh catalog.Contract)
  → messaging, scheduler   (infrastruktur asinkron, dipakai module lain lewat Contract)
```

Catatan penting:
- **Kalau dua module butuh Contract satu sama lain (siklus dependency)**,
  pecahkan dengan late-binding: module A dibuat dulu tanpa referensi ke
  module B, module B dibuat sesudahnya menerima `A.Contract`, baru
  kemudian dipanggil setter semacam `a.SetBProvisioner(b)`. Jangan
  gabungkan dua module jadi satu supaya siklusnya "hilang" — itu
  menghancurkan batas domain yang justru ingin dijaga.
- `cmd/api` dan `cmd/worker` merakit **module yang sama** tapi mendaftarkan
  permukaan yang berbeda: `cmd/api` memanggil `RegisterRoutes` (HTTP),
  `cmd/worker` memanggil `RegisterMessageHandlers` (consumer asinkron) dan
  menjalankan dispatcher scheduler. Module itu sendiri tidak tahu proses
  mana yang memanggilnya.
- Jangan bikin container DI generik atau module-registry. Constructor
  injection manual seperti di atas sudah cukup dan tetap type-safe — lihat
  alasan lengkap di penutup `module-boundaries.md`.

## `internal/shared` — apa yang boleh masuk sini

`shared/` **bukan** tempat untuk kode domain atau logika bisnis apa pun —
kalau ada dua module butuh logika yang sama secara domain, itu adalah
sinyal untuk komunikasi lewat `Contract`, bukan untuk diekstrak ke
`shared/`. Yang boleh ada di `shared/` hanya infrastruktur teknis generik
yang tidak punya opini domain:

| Folder | Isi |
|---|---|
| `kernel/` | Tipe generik lintas-module: error sentinel generik (mis. `ErrNotFound`), ID generator |
| `config/` | Load environment variable jadi satu struct `Config` |
| `database/` | Koneksi database, transactor (`WithTx`) |
| `logger/` | Setup structured logger |
| `middleware/` | Middleware HTTP generik: CORS, request-id, request-logger, error-handler, rate-limit |
| `httperror/` | Konversi error domain/kernel → HTTP response |
| `respond/` | Helper format response sukses/error yang konsisten antar-module |
| `timeutil/` | Timezone platform untuk parse/format tanggal-waktu |

## Aturan menambah module baru

1. Ikuti struktur folder di atas (`domain/`, `application/{command,query,
   ports,dto}/`, `infrastructure/`, `interfaces/`).
2. Kalau module ini perlu dipanggil module lain, tambahkan `contract.go`
   sesuai pola di `module-boundaries.md` — jangan expose tipe module
   konkret atau tipe dari `application/ports`/`domain` ke luar.
3. Kalau module ini perlu memanggil module lain, module tujuan diterima
   lewat parameter `NewModule(...)` bertipe `Contract`-nya (bukan tipe
   konkret), dan diakses lewat adapter di
   `infrastructure/<nama>gateway/` yang mengimplementasikan port di
   `application/ports/` milik module ini sendiri.
4. Daftarkan wiring-nya di `cmd/api/main.go` (route) dan, kalau module
   punya handler asinkron atau job terjadwal, juga di `cmd/worker/main.go`.
5. Tambah migrasi di `migrations/` (flat, prefix timestamp).
6. Jangan desain `Contract` atau `ports` untuk kebutuhan hipotetis (YAGNI)
   — lihat penjelasan lengkap di `module-boundaries.md`.
