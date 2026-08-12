package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"

	"sipon-be/internal/shared/timeutil"
)

func main() {
	_ = godotenv.Load()

	if err := timeutil.Init(getEnv("APP_TIMEZONE", "Asia/Jakarta")); err != nil {
		log.Fatalf("gagal load timezone: %v", err)
	}

	dsn := buildDSN()
	migrationsDir := getEnv("MIGRATIONS_DIR", "migrations")

	m, err := migrate.New("file://"+migrationsDir, dsn)
	if err != nil {
		log.Fatalf("gagal inisialisasi migrasi: %v", err)
	}
	defer m.Close()

	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("perintah diperlukan: up, down, fresh, version, force")
	}

	switch args[0] {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrasi up gagal: %v", err)
		}
		log.Println("migrasi up berhasil")
	case "down":
		n := 1
		if len(args) >= 2 {
			var err error
			n, err = strconv.Atoi(args[1])
			if err != nil || n < 1 {
				log.Fatalf("jumlah langkah down tidak valid: %s", args[1])
			}
		}
		if err := m.Steps(-n); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrasi down gagal: %v", err)
		}
		log.Printf("migrasi down %d langkah berhasil\n", n)
	case "fresh":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrasi down gagal: %v", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrasi up gagal: %v", err)
		}
		log.Println("migrasi fresh berhasil")
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("gagal mendapatkan versi: %v", err)
		}
		log.Printf("versi: %d, dirty: %v\n", version, dirty)
	case "force":
		if len(args) < 2 {
			log.Fatal("perintah force memerlukan nomor versi")
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("versi tidak valid: %v", err)
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("force gagal: %v", err)
		}
		log.Printf("force ke versi %d berhasil\n", v)
	default:
		log.Fatalf("perintah tidak dikenal: %s", args[0])
	}
}

func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "sipon")
	pass := getEnv("DB_PASS", "secret")
	name := getEnv("DB_NAME", "sipon")
	sslmode := getEnv("DB_SSLMODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, sslmode)
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
