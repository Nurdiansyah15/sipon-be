package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	SMTP        SMTPConfig
	Fonnte      FonnteConfig
	Redis       RedisConfig
	RateLimit   RateLimitConfig
	Minio       MinioConfig
	Fingerprint FingerprintConfig
	Worker      WorkerConfig
	Google      GoogleConfig
}

type GoogleConfig struct {
	ClientIDs []string
}

type WorkerConfig struct {
	Enabled     bool
	TickSeconds int
}

// FingerprintConfig mengatur integrasi absensi mesin fingerprint.
type FingerprintConfig struct {
	// SandboxEnabled menghidupkan endpoint sandbox (simulasi mesin fingerprint
	// palsu). Default false — di production endpoint ini tidak didaftarkan.
	SandboxEnabled bool
}

type AppConfig struct {
	Port      string
	Env       string
	LogFormat string
	Timezone  string
}

type DatabaseConfig struct {
	DSN        string
	DB_HOST    string
	DB_PORT    string
	DB_USER    string
	DB_PASS    string
	DB_NAME    string
	DB_SSLMODE string
}

type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type FonnteConfig struct {
	Token string
	URL   string
}

type RedisConfig struct {
	Addr string
}

type RateLimitConfig struct {
	Enabled           bool
	IPLimit           int
	IPWindowSeconds   int
	UserLimit         int
	UserWindowSeconds int
	AuthLimit         int
	AuthWindowSeconds int
}

type MinioConfig struct {
	// Endpoint dipakai oleh server untuk operasi yang butuh koneksi nyata ke
	// MinIO (Stat/Remove object) — harus alamat yang bisa dijangkau DARI DALAM
	// container, contoh: nama service docker compose "minio:9000".
	Endpoint string
	// PublicEndpoint dipakai HANYA untuk membuat presigned URL & public URL —
	// harus alamat yang bisa dijangkau BROWSER/FE (mis. "localhost:9004" atau
	// IP LAN mesin host). Jangan disamakan dengan Endpoint: nama service
	// docker ("minio") tidak bisa di-resolve dari luar container.
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	Bucket         string
	PrivateBucket  string
	// UseSSL dipakai untuk endpoint publik (presign/PublicURL) — biasanya
	// HTTPS di production (reverse proxy di depan MinIO).
	UseSSL bool
	// InternalUseSSL dipakai untuk client internal (Stat/Remove object). Di
	// docker network, minio biasanya plain HTTP (minio:9000) meski endpoint
	// publiknya HTTPS — jadi pisahkan agar tidak "HTTPS client ke server HTTP".
	InternalUseSSL bool
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	cfg := &Config{
		App: AppConfig{
			Port:      getEnv("APP_PORT", "8080"),
			Env:       getEnv("APP_ENV", "development"),
			LogFormat: getEnv("LOG_FORMAT", ""),
			Timezone:  getEnv("APP_TIMEZONE", "Asia/Jakarta"),
		},
		Database: DatabaseConfig{
			DB_HOST:    requireEnv("DB_HOST"),
			DB_PORT:    requireEnv("DB_PORT"),
			DB_USER:    requireEnv("DB_USER"),
			DB_PASS:    requireEnv("DB_PASS"),
			DB_NAME:    requireEnv("DB_NAME"),
			DB_SSLMODE: requireEnv("DB_SSLMODE"),
			DSN: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				getEnv("DB_USER", "sipon"),
				getEnv("DB_PASS", "secret"),
				getEnv("DB_HOST", "localhost"),
				getEnv("DB_PORT", "5432"),
				getEnv("DB_NAME", "sipon"),
				getEnv("DB_SSLMODE", "disable"),
			),
		},
		JWT: JWTConfig{
			SecretKey:       requireEnv("JWT_SECRET_KEY"),
			AccessTokenTTL:  parseDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: parseDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@sipon.dev"),
		},
		Fonnte: FonnteConfig{
			Token: getEnv("FONNTE_TOKEN", ""),
			URL:   getEnv("FONNTE_URL", "https://api.fonnte.com/send"),
		},
		Redis: RedisConfig{
			Addr: getEnv("REDIS_ADDR", "redis:6379"),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnv("RATE_LIMIT_ENABLED", "true") != "false",
			IPLimit:           parseInt("RATE_LIMIT_IP_MAX", 120),
			IPWindowSeconds:   parseInt("RATE_LIMIT_IP_WINDOW_SECONDS", 60),
			UserLimit:         parseInt("RATE_LIMIT_USER_MAX", 300),
			UserWindowSeconds: parseInt("RATE_LIMIT_USER_WINDOW_SECONDS", 60),
			AuthLimit:         parseInt("RATE_LIMIT_AUTH_MAX", 10),
			AuthWindowSeconds: parseInt("RATE_LIMIT_AUTH_WINDOW_SECONDS", 60),
		},
		Minio: MinioConfig{
			Endpoint:       getEnv("MINIO_ENDPOINT", ""),
			PublicEndpoint: getEnv("MINIO_PUBLIC_ENDPOINT", getEnv("MINIO_ENDPOINT", "")),
			AccessKey:      getEnv("MINIO_ACCESS_KEY", ""),
			SecretKey:      getEnv("MINIO_SECRET_KEY", ""),
			Bucket:         getEnv("MINIO_BUCKET", "sipon-public"),
			PrivateBucket:  getEnv("MINIO_PRIVATE_BUCKET", "sipon-private"),
			UseSSL:         getEnv("MINIO_USE_SSL", "false") == "true",
			InternalUseSSL: getEnv("MINIO_INTERNAL_USE_SSL", "false") == "true",
		},
		Fingerprint: FingerprintConfig{
			SandboxEnabled: getEnv("FINGERPRINT_SANDBOX_ENABLED", "false") == "true",
		},
		Worker: WorkerConfig{
			Enabled:     getEnv("WORKER_ENABLED", "true") != "false",
			TickSeconds: parseInt("WORKER_TICK_SECONDS", 10),
		},
		Google: GoogleConfig{
			ClientIDs: parseCSV("GOOGLE_CLIENT_IDS", nil),
		},
	}
	return cfg, nil
}

func parseCSV(key string, defaultVal []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
	return v
}

func parseInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func parseDuration(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	minutes, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return time.Duration(minutes) * time.Minute
}
