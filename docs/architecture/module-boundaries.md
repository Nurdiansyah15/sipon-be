# Batas Antar-Module (Modular Monolith DDD)

## Aturan

Module harus terisolasi. Satu module **hanya boleh** memanggil module lain
lewat `Contract` yang di-expose module tersebut — tidak pernah lewat
domain entity, repository interface, atau `application/ports`-nya secara
langsung.

Alurnya selalu satu arah, dari sudut pandang module **pemanggil**:

```
Contract module tujuan
        │  (di-import)
        ▼
application/ports milik module pemanggil (port, dalam kosakata module pemanggil)
        │  (diimplementasikan oleh)
        ▼
infrastructure/<nama-gateway> milik module pemanggil (adapter)
        │  (dipakai oleh)
        ▼
usecase (application/command atau application/query) milik module pemanggil
```

Module pemanggil tidak pernah tahu module tujuan itu pakai Postgres, Redis,
atau struktur domain seperti apa — yang ia tahu cuma `Contract`-nya.

## Pola `contract.go`

Setiap module yang ingin bisa dipanggil module lain menyediakan satu file
`contract.go` sejajar dengan `module.go`-nya (package sama, bukan
subpackage — subpackage cuma nambah satu hop import tanpa menambah isolasi
nyata, karena yang benar-benar menegakkan batas adalah *apa yang diexport*,
bukan *seberapa dalam foldernya*).

Isi `contract.go`:
1. Interface `Contract` — satu-satunya tipe yang boleh di-import module lain
   dari module ini. Semua parameter dan return value-nya harus DTO milik
   module ini sendiri (dideklarasikan di file yang sama), **tidak pernah**
   entity domain atau tipe dari `application/ports`.
2. `var _ Contract = (*Module)(nil)` — assertion compile-time supaya kalau
   `Module` lupa mengimplementasikan salah satu method Contract, build
   langsung gagal, bukan ketahuan saat runtime.
3. Method-method `Contract` diimplementasikan langsung di `*Module`,
   biasanya cuma mendelegasikan ke usecase yang sudah ada (atau usecase baru
   yang sengaja dibuat minimal khusus untuk contract ini — lihat contoh
   `identity`).

Contoh nyata: `internal/modules/identity/contract.go`.
```go
type Contract interface {
	GetUserSummary(ctx context.Context, userID string) (*UserSummary, error)
	GetPrincipal(ctx context.Context, userID string) (*Principal, error)
}

var _ Contract = (*Module)(nil)
```

Catatan penting soal DTO kontrak: **jangan** pakai ulang DTO yang sudah ada
untuk kebutuhan lain (mis. `dto.UserManagementResponse` yang bentuknya untuk
admin UI). Kalau bentuk data admin UI berubah, kontrak lintas-module ikut
berubah padahal tidak ada hubungannya — itu sebabnya `identity.UserSummary`
adalah tipe sendiri, dan usecase di baliknya (`GetUserSummaryUseCase`) juga
usecase baru yang sengaja diminimalkan, bukan pakai ulang `GetUserUseCase`.

Jangan mendesain `Contract` untuk kebutuhan hipotetis. Mulai dari method
yang benar-benar dibutuhkan module lain HARI INI; tambah method baru begitu
ada module nyata yang butuh kapabilitas baru (YAGNI). Contract yang gemuk
dari awal cuma jadi permukaan yang harus dijaga kompatibel tanpa ada yang
memakainya.

## `Module` itu sendiri: zero exported field, permukaan public murni method

`internal/modules/identity/module.go` sengaja dibuat **tanpa satupun field
exported**. Semua yang dirakit `NewModule` (repo, hasher, token generator,
usecase, dst.) disimpan sebagai field privat di `*Module`, atau kalau cuma
dipakai sekali untuk construct usecase lain, tidak disimpan sama sekali.

Permukaan public `*Module` hanya 4 method:
- `RegisterRoutes(router gin.IRouter)` — dipakai `cmd/api/main.go` buat
  mounting HTTP.
- `RateLimiter() ports.RateLimiter` — dipakai `cmd/api/main.go` buat
  merakit rate-limit middleware global. Return type-nya interface
  (`ports.RateLimiter`), bukan tipe konkret `*cache.RedisRateLimiter` —
  pemanggil tidak perlu tahu implementasinya Redis.
- `GetUserSummary(...)` / `GetPrincipal(...)` — dua method `Contract`,
  untuk module lain (lihat contract.go).

Kenapa **tidak** sekalian expose usecase (mis. `getUserSummaryUC`) jadi
field public, atau bikin semua dependency jadi field public seperti
`Repositories`/`Services` versi awal dulu:
1. **Consumer jadi terikat ke tipe konkret internal.** Kalau usecase
   direfactor/dipecah/ganti nama, siapa pun yang pegang field itu langsung
   ikut rusak — padahal tidak ada perubahan "kontrak" secara konsep.
2. **Titik penegakan aksesnya hilang.** Kalau module lain nanti
   constructor-nya menerima `*identity.Module` (bukan `identity.Contract`),
   dia otomatis BISA manggil `RegisterRoutes`/`RateLimiter()` juga — padahal
   itu cuma buat `main.go`. `Contract` (2 method) itu yang bikin "akses
   sesempit mungkin" BENERAN ditegakkan Go: selama constructor module lain
   mendeklarasikan parameternya bertipe `identity.Contract`, bukan
   `*identity.Module`, static typing Go membatasi dia cuma bisa manggil 2
   method itu — walau value konkret yang dioper (`*Module`) sebenarnya
   "punya" 4 method. `main.go` boleh pegang `*identity.Module` penuh
   (karena dia yang menyambungkan semuanya), tapi module lain **harus**
   menerima tipe `Contract`.

Kalau suatu saat menulis kode yang meng-import field dari `someModule.Module`
langsung (bukan lewat `Contract`-nya) dari module LAIN (bukan dari
`cmd/api/main.go`), itu tandanya arsitekturnya sudah menyimpang — perbaiki
dengan menambah method ke `Contract` module tujuan, bukan menjangkau
`Module`-nya langsung. Karena semua field di `Module` privat, kesalahan ini
seharusnya tidak mungkin lolos compiler — kalaupun ada yang mencoba, Go akan
menolaknya duluan.

## Contoh lengkap: module kedua (`content`, hipotetis)

Belum ada module kedua di repo ini — bagian ini murni template untuk siapa
pun yang menambah module berikutnya.

1. **`internal/modules/content/application/ports/identity_reader.go`** — port
   milik `content` sendiri, dalam kosakata `content`:
   ```go
   package ports

   type IdentityReader interface {
       GetUserSummary(ctx context.Context, userID string) (*identity.UserSummary, error)
       GetPrincipal(ctx context.Context, userID string) (*identity.Principal, error)
   }
   ```
   Boleh langsung pakai `identity.UserSummary`/`identity.Principal` sebagai
   tipe balik — itu sudah DTO milik batas kontrak, tidak perlu translasi
   kedua yang tidak menambah nilai.

2. **`internal/modules/content/infrastructure/identitygateway/gateway.go`** —
   adapter yang mengimplementasikan port di atas:
   ```go
   package identitygateway

   type Gateway struct {
       contract identity.Contract
   }

   func New(contract identity.Contract) *Gateway {
       return &Gateway{contract: contract}
   }

   func (g *Gateway) GetUserSummary(ctx context.Context, userID string) (*identity.UserSummary, error) {
       return g.contract.GetUserSummary(ctx, userID)
   }
   // GetPrincipal serupa.
   ```

3. Usecase `content` menerima `ports.IdentityReader` lewat constructor
   (dependency inversion) — tidak pernah `identity.Contract` langsung,
   apalagi `identity/domain/*` atau `identity/infrastructure/*`.

4. **`cmd/api/main.go`** merakit keduanya, identity duluan:
   ```go
   identity := identityModule.NewModule(db, redisClient, cfg)
   content := contentModule.NewModule(db, redisClient, cfg, identity)
   ```
   `content.NewModule` menerima `identityContract identity.Contract`
   sebagai parameter — `*identity.Module` otomatis memenuhi interface itu,
   jadi `identity` (concrete type) bisa langsung dioper.

## Yang di luar cakupan dokumen ini

`internal/shared/eventbus/eventbus.go` sudah ada di repo tapi belum dipakai
module mana pun. Kalau nanti kebutuhan komunikasi antar-module sifatnya
fire-and-forget/async (bukan request-response seperti `Contract`), itu
tempatnya dilirik — bukan pengganti `Contract` untuk kebutuhan sinkron.

Jangan bikin abstraksi DI container atau module-registry generik. Constructor
injection biasa (seperti contoh `content.NewModule(..., identity)` di atas)
sudah cukup untuk berapa pun jumlah module yang akan ada di repo ini.
