package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"sipon-be/internal/seeders"
	"sipon-be/internal/shared/timeutil"
)

func main() {
	_ = godotenv.Load()

	if err := timeutil.Init(getEnv("APP_TIMEZONE", "Asia/Jakarta")); err != nil {
		log.Fatalf("gagal load timezone: %v", err)
	}

	dsn := buildDSN()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("gagal koneksi database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	args := os.Args[1:]

	seederRegistry := seeders.NewRegistry()

	if len(args) == 0 || args[0] == "all" {
		if err := seederRegistry.RunAll(ctx, db); err != nil {
			log.Fatalf("seeder gagal: %v", err)
		}
		log.Println("semua seeder berhasil dijalankan")
		return
	}

	if err := seederRegistry.Run(ctx, db, args[0]); err != nil {
		log.Fatalf("seeder %s gagal: %v", args[0], err)
	}
	log.Printf("seeder %s berhasil dijalankan\n", args[0])
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
