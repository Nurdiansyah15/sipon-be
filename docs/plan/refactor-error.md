Prompt general yang bisa dipakai di codebase DDD manapun:
Kami sedang menerapkan pola error handling konsisten di seluruh layer DDD. Ikuti langkah berikut secara runtut, dan jalankan build command (go build ./... / cargo build / etc.) setelah tiap step untuk verifikasi.
Step 1 — Domain Layer: bungkus error dengan kode + pesan

1. Cari file konstanta error di domain (mis. <domain>/constant/ atau <domain>/errors.go) yang berisi definisi kode error (string/enum).
2. Telusuri semua file di domain layer (entity, value object, aggregate) yang membuat error menggunakan constructor tanpa pesan (mis. kernel.New(code), errors.New(...), atau constructor app-error tanpa message).
3. Ganti menjadi constructor yang membawa kode + pesan + nil (mis. kernel.WrapMsg(code, "pesan dalam bahasa yang sesuai", nil)).
4. Pesan harus deskriptif dan end-user-friendly, disesuaikan dengan makna tiap kode error — bukan sekadar copy kode errornya.
   Contoh:
   Sebelum: return Email{}, kernel.New(constant.ErrCodeEmailEmpty)
   Sesudah: return Email{}, kernel.WrapMsg(constant.ErrCodeEmailEmpty, "Email tidak boleh kosong", nil)
   Step 2 — Application Layer: teruskan pesan domain ke app error
5. Cari semua file di application layer (use case / service) yang menangkap domain error via type assertion (errors.As) lalu membungkusnya menjadi application error.
6. Jika app error dibuat tanpa memasukkan pesan domain (mis. Wrap(appCode, err), New(appCode), atau WrapMsg(appCode, string(kode), err)):

- Ganti string(ke.Code) → ke.Message di panggilan WrapMsg
- Ganti Wrap(appCode, err) → WrapMsg(appCode, ke.Message, ke)
- Ganti New(appCode) → WrapMsg(appCode, ke.Message, ke)

3. Pastikan setiap mapping domain error → app error menyertakan pesan dari domain.
   Step 3 — Repository/Infrastructure Layer: definisikan error yg tepat
1. Tambah kode error baru di domain constant jika diperlukan:

- ErrCodeNotFound / ErrCodeResourceNotFound — khusus untuk repository "not found"
- ErrCodeInternal — khusus untuk infrastructure/internal errors

2. Di file repository implementation:

- Hapus import "fmt" (jika hanya dipakai untuk error wrapping)
- Ganti semua error wrapping infrastruktur (fmt.Errorf, errors.Wrap, dll.) menjadi WrapMsg(ErrCodeInternal, "pesan spesifik per operasi", err) — setiap operasi DB/network punya pesan unik (mis. "gagal mencari data", "gagal menyimpan", "gagal memulai transaksi").
- Untuk not-found (mis. sql.ErrNoRows): ganti menjadi WrapMsg(ErrCodeNotFound, "Resource tidak ditemukan", nil).
- Untuk tx.Commit(): wrap kegagalannya juga: if err := tx.Commit(); err != nil { return WrapMsg(ErrCodeInternal, "gagal menyimpan transaksi", err) }.

3. Hasil akhir: zero fmt.Errorf, zero bare kernel.New di repository.
   Step 4 — Mapping error repo → use case
1. Di semua use case yang memanggil repository find methods, pastikan error code yang ditangkap sama persis dengan yang dilempar repository:

- Jika repo lempar ErrCodeNotFound, use case harus menangkap ErrCodeNotFound (bukan kode lain seperti ErrCodeInvalidXxx).

2. Hati-hati: jika sebuah file use case menangkap error code yang sama dari dua sumber berbeda (mis. ErrCodeNotActive dari entity call DAN dari repo find), pisahkan berdasarkan konteks:

- Dari repo → gunakan ErrCodeNotFound
- Dari entity → gunakan ErrCodeNotActive (sesuai domain)

3. Jangan lakukan replaceAll buta pada file dengan konteks campuran.
   Aturan Penting

- Setiap error yang dibuat di domain dan repository WAJIB menggunakan constructor dengan pesan (WrapMsg).
- Application layer hanya melakukan mapping kode (domain → HTTP/semantic) sambil meneruskan pesan domain.
- Jangan ubah error entity yang berasal dari business rule (mis. EnsureCanLogin, Activate, Reactivate) — hanya mapping dari panggilan repository yang disesuaikan.
- Setiap langkah diakhiri dengan build command untuk verifikasi.

# 2

Berikut prompt yang bisa langsung diberikan ke AI:
Instruksi:
Kerjakan secara runtut langkah demi langkah berikut pada codebase Go dengan DDD structure di /home/nurdiansyah/Desktop/sipon/sipon-be. Scope: hanya internal/modules/identity/domain/user/ dan internal/modules/identity/application/ serta internal/modules/identity/infrastructure/persistence/postgres_user_repo.go.
Step 1 — Domain: ganti kernel.New menjadi kernel.WrapMsg dengan pesan Indonesia
Baca file internal/modules/identity/domain/user/constant/user_constant.go (berisi daftar kernel.Code konstanta error). Lalu cari SEMUA tempat di dalam domain/user/ (entity, value object, login_identity) dan infrastructure/persistence/postgres_user_repo.go yang membuat error dengan kernel.New(constant.ErrCodeXxx) — ganti menjadi kernel.WrapMsg(constant.ErrCodeXxx, "pesan indonesia", nil). Pesan disesuaikan dengan makna error code, contoh:

- ErrCodeEmailEmpty → "Email tidak boleh kosong"
- ErrCodePlainPasswordTooShort → "Kata sandi terlalu pendek (minimal 8 karakter)"
- ErrCodeUserBanned → "Pengguna telah diblokir"
- ErrCodeUserNotActive → "Pengguna tidak aktif"
- ErrCodeUserLockedOut → "Pengguna terkunci sementara"
- ErrCodeIdentityNotVerified → "Identitas belum diverifikasi"
- ErrCodeCredentialNotLocal → "Kredensial bukan tipe lokal"
- ErrCodeUsernameEmpty → "Username tidak boleh kosong"
- ErrCodeUsernameTooShort → "Username terlalu pendek (minimal 3 karakter)"
- ErrCodeUsernameTooLong → "Username terlalu panjang (maksimal 30 karakter)"
- ErrCodeUsernameInvalidChar → "Username hanya boleh mengandung huruf, angka, dan underscore"
- ErrCodeLoginIdentifierEmpty → "Identitas login tidak boleh kosong"
- ErrCodeLoginIdentifierUnknownKind → "Jenis identitas login tidak dikenali"
- ErrCodeInvalidLoginIdentityValue → "Nilai identitas login tidak valid"
- ErrCodeHashedPasswordTooShort → "Hashed password terlalu pendek"
- ErrCodePlainPasswordEmpty → "Kata sandi tidak boleh kosong"
- ErrCodePlainPasswordNoUppercase → "Kata sandi harus mengandung huruf kapital"
- ErrCodePlainPasswordNoDigit → "Kata sandi harus mengandung angka"
- ErrCodeOTPCodeEmpty → "Kode OTP tidak boleh kosong"
- ErrCodeOTPCodeInvalidLength → "Panjang kode OTP harus 6 digit"
- ErrCodeOTPCodeNotDigit → "Kode OTP harus berupa angka"
- ErrCodePhoneNumberEmpty → "Nomor telepon tidak boleh kosong"
- ErrCodePhoneNumberInvalidFormat → "Format nomor telepon tidak valid"
- ErrCodeUserIDRequired → "ID pengguna wajib diisi"
- ErrCodeUserEmailRequired → "Email pengguna wajib diisi"
- ErrCodeUserPhoneNumberInvalid → "Nomor telepon pengguna tidak valid"
- ErrCodeUserAlreadyActive → "Pengguna sudah aktif"
- ErrCodeUserAlreadyBanned → "Pengguna sudah diblokir"
- ErrCodeUserAlreadyDeleted → "Pengguna sudah dihapus"
  File yang perlu diedit di domain:
- internal/modules/identity/domain/user/valueobject/user_vo.go
- internal/modules/identity/domain/user/entity/user.go
- internal/modules/identity/domain/user/entity/login_identity.go
  Step 2 — Application: propagasi ke.Message di semua kernel.WrapMsg
  Di SEMUA file di bawah internal/modules/identity/application/command/ dan internal/modules/identity/application/query/ yang menangkap error domain via errors.As(err, &ke):
- Ganti string(ke.Code) → ke.Message (replaceAll) di semua panggilan kernel.WrapMsg(application.Xxx, string(ke.Code), ke).
- Ganti kernel.Wrap(appCode, err) → kernel.WrapMsg(appCode, ke.Message, ke) untuk setiap case yang menangkap domain error code (seperti ErrCodeInvalidLoginIdentityValue/ErrCodeUserNotActive/ErrCodeUserBanned/ErrCodeUserAlreadyBanned/ErrCodeCredentialNotLocal).
- Ganti kernel.New(appCode) → kernel.WrapMsg(appCode, ke.Message, ke) untuk case yang menangkap domain error code lalu membuat application error baru.
  File command yang diubah: login.go, register.go, reset_password.go, change_password.go, set_password.go, refresh_token.go, verify_otp.go, create_account_with_nis.go, change_username.go, change_identity.go, add_nis_login_identity.go, manage_user.go, update_profile.go, update_fullname.go, request_otp.go, avatar.go.
  File query yang diubah: check_username.go, get_user.go, get_user_summary.go, get_profile.go.
  Step 3 — Repository: definisikan error domain yang benar
  Tambah 2 error code baru di user_constant.go:
  ErrCodeUserNotFound kernel.Code = "USER_NOT_FOUND"
  ErrCodeInternal kernel.Code = "USER_INTERNAL_ERROR"
  Di postgres_user_repo.go:
- Hapus import "fmt"
- Ganti SEMUA fmt.Errorf("...", err) → kernel.WrapMsg(userconstant.ErrCodeInternal, "pesan indonesia", err) — setiap error DB diberi pesan spesifik:
- "gagal memulai transaksi database" (begin tx)
- "gagal memperbarui pengguna" (update user)
- "gagal menyimpan kredensial" (insert/upsert credential)
- "gagal menyimpan identitas login" (insert/upsert login identity)
- "gagal menyimpan transaksi" (tx.Commit)
- "gagal mencari pengguna berdasarkan ID" (find by id DB error)
- "gagal mencari pengguna berdasarkan identitas login" (find by login identifier DB error)
- "gagal mencari pengguna berdasarkan username" (find by username DB error)
- "gagal memeriksa ketersediaan username" (exists by username)
- "gagal memeriksa ketersediaan identitas login" (exists by login identity)
- "gagal memperbarui username" (update username)
- "gagal menyimpan pengguna" (insert user)
- "gagal memuat kredensial" (load credentials)
- "gagal membaca data kredensial" (scan credential)
- "gagal memuat identitas login" (load login identities)
- "gagal membaca data identitas login" (scan login identity)
- Ganti ErrCodeUserNotActive → ErrCodeUserNotFound di ketiga find method (FindByID, FindByLoginIdentifier, FindByUsername) pada blok sql.ErrNoRows, gunakan pesan "Pengguna tidak ditemukan".
- Wrap tx.Commit() error di Save dan Update dengan kernel.WrapMsg(ErrCodeInternal, "gagal menyimpan transaksi", err).
  Step 4 — Application: mapping error code repo → use case
  Di SEMUA file use case (command + query) yang menangkap error dari panggilan userRepo.FindByID, userRepo.FindByIdentity, userRepo.FindByLoginIdentifier, userRepo.FindByUsername, userRepo.Save, userRepo.Update:
- Ganti userconstant.ErrCodeInvalidLoginIdentityValue → userconstant.ErrCodeUserNotFound (replaceAll di file yang hanya punya repo-find context).
- Untuk file dengan MIXED context (refresh_token.go, manage_user.go): hanya ganti yang dari panggilan repo (FindByID/Save/Update), JANGAN ganti yang dari panggilan entity (EnsureCanLogin, Reactivate) — biarkan tetap ErrCodeUserNotActive.
- Pertahankan ErrCodeUserNotActive di file yang hanya punya entity context: login.go dan register.go (dari EnsureCanLogin).
  Setiap step selesai, jalankan go build ./... untuk verifikasi kompilasi.
