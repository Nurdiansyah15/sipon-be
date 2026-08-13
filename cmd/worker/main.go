package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"

	akademikModule "sipon-be/internal/modules/akademik"
	articleModule "sipon-be/internal/modules/article"
	dokumenAsetModule "sipon-be/internal/modules/dokumen_aset"
	feedbackModule "sipon-be/internal/modules/feedback"
	fingerprintModule "sipon-be/internal/modules/fingerprint"
	identityModule "sipon-be/internal/modules/identity"
	kesantrianModule "sipon-be/internal/modules/kesantrian"
	keuanganModule "sipon-be/internal/modules/keuangan"
	psbModule "sipon-be/internal/modules/psb"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/database"
	"sipon-be/internal/shared/logger"
	"sipon-be/internal/shared/scheduler/application"
	"sipon-be/internal/shared/scheduler/infrastructure/persistence"
	"sipon-be/internal/shared/timeutil"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal load config: %v", err)
	}

	if err := timeutil.Init(cfg.App.Timezone); err != nil {
		log.Fatalf("gagal load timezone %s: %v", cfg.App.Timezone, err)
	}

	lg := logger.New(cfg.App.Env, cfg.App.LogFormat)
	slog.SetDefault(lg)

	db, err := database.Open(cfg.Database.DSN)
	if err != nil {
		lg.Error("gagal koneksi database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()
	lg.Info("worker: terhubung ke PostgreSQL")

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		lg.Warn("Redis tidak dapat dijangkau", slog.Any("error", err))
	} else {
		lg.Info("worker: terhubung ke Redis", slog.String("addr", cfg.Redis.Addr))
	}
	defer redisClient.Close()

	identity := identityModule.NewModule(db, redisClient, cfg)
	dokumenAset := dokumenAsetModule.NewModule(db, cfg,
		identity.AuthMiddleware(), identity.PrincipalMiddleware())
	kesantrian := kesantrianModule.NewModule(db, redisClient, cfg,
		identity, dokumenAset,
		identity.AuthMiddleware(), identity.PrincipalMiddleware())
	psb := psbModule.NewModule(db, cfg,
		identity, kesantrian,
		identity.AuthMiddleware(), identity.PrincipalMiddleware())
	article := articleModule.NewModule(db, cfg,
		identity.AuthMiddleware(), identity.PrincipalMiddleware())
	keuangan := keuanganModule.NewModule(db, cfg, kesantrian,
		identity.AuthMiddleware(), identity.PrincipalMiddleware())
	feedback := feedbackModule.NewModule(db, cfg, identity,
		identity.AuthMiddleware(), identity.PrincipalMiddleware())
	fingerprint := fingerprintModule.NewModule(db, cfg,
		identity.AuthMiddleware(), identity.PrincipalMiddleware())
	akademik := akademikModule.NewModule(db, cfg,
		kesantrian, fingerprint,
		identity.AuthMiddleware(), identity.PrincipalMiddleware())
	kesantrian.SetAkademikProvisioner(akademik)

	registry := application.NewRegistry()
	identity.RegisterSchedulerHandlers(registry)
	dokumenAset.RegisterSchedulerHandlers(registry)
	kesantrian.RegisterSchedulerHandlers(registry)
	psb.RegisterSchedulerHandlers(registry)
	article.RegisterSchedulerHandlers(registry)
	keuangan.RegisterSchedulerHandlers(registry)
	feedback.RegisterSchedulerHandlers(registry)
	fingerprint.RegisterSchedulerHandlers(registry)
	akademik.RegisterSchedulerHandlers(registry)

	scheduledJobRepo := persistence.NewPostgresScheduledJobRepository(db)
	worker := application.NewWorker(
		scheduledJobRepo,
		registry,
		time.Duration(cfg.Worker.TickSeconds)*time.Second,
		lg,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Run(ctx)

	lg.Info("worker started, waiting for scheduled jobs")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	lg.Info("shutdown signal received")
	cancel()

	time.Sleep(2 * time.Second)
	lg.Info("worker stopped")
}
