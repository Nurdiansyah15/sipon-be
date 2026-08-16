package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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
	"sipon-be/internal/shared/middleware"
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
	lg.Info("terhubung ke PostgreSQL")

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		lg.Warn("Redis tidak dapat dijangkau, akan berjalan tanpa cache", slog.Any("error", err))
	} else {
		lg.Info("terhubung ke Redis", slog.String("addr", cfg.Redis.Addr))
	}
	defer redisClient.Close()

	identity := identityModule.NewModule(db, redisClient, cfg)

	dokumenAset := dokumenAsetModule.NewModule(
		db, cfg,
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	kesantrian := kesantrianModule.NewModule(
		db, redisClient, cfg,
		identity,    // identity.Contract (termasuk scope access resolution)
		dokumenAset, // dokumen_aset.Contract
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	psb := psbModule.NewModule(
		db, cfg,
		identity,   // identity.Contract
		kesantrian, // kesantrian.Contract (needs CreateSantriFromPendaftaran)
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	article := articleModule.NewModule(
		db, cfg,
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	keuangan := keuanganModule.NewModule(
		db, cfg,
		kesantrian, // kesantrian.Contract
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	feedback := feedbackModule.NewModule(
		db, cfg,
		identity, // identity.Contract
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	fingerprint := fingerprintModule.NewModule(
		db, cfg,
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	akademik := akademikModule.NewModule(
		db, cfg,
		kesantrian,  // kesantrian.Contract
		fingerprint, // fingerprint.Contract
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	// Late-binding: akademik bergantung pada kesantrian di konstruktor, sehingga
	// kesantrian menerima kontrak akademik setelah akademik terbentuk. Lihat
	// docs/plan/santri-program-mapping.md.
	kesantrian.SetAkademikProvisioner(akademik)

	const pendingUploadExpireDays = 1

	if err := identity.EnsurePendingUploadLifecycle(context.Background(), pendingUploadExpireDays); err != nil {
		lg.Warn("gagal set lifecycle pending upload (identity)", slog.Any("error", err))
	}
	if err := kesantrian.EnsurePendingUploadLifecycle(context.Background(), pendingUploadExpireDays); err != nil {
		lg.Warn("gagal set lifecycle pending upload (kesantrian)", slog.Any("error", err))
	}
	if err := psb.EnsurePendingUploadLifecycle(context.Background(), pendingUploadExpireDays); err != nil {
		lg.Warn("gagal set lifecycle pending upload (psb)", slog.Any("error", err))
	}
	if err := feedback.EnsurePendingUploadLifecycle(context.Background(), pendingUploadExpireDays); err != nil {
		lg.Warn("gagal set lifecycle pending upload (feedback)", slog.Any("error", err))
	}
	if err := dokumenAset.EnsurePendingUploadLifecycle(context.Background(), pendingUploadExpireDays); err != nil {
		lg.Warn("gagal set lifecycle pending upload (dokumen_aset)", slog.Any("error", err))
	}

	engine := gin.New()

	engine.Use(middleware.RequestID())
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.RequestLogger(lg))
	engine.Use(middleware.ErrorHandler(lg))

	if cfg.RateLimit.Enabled && identity.RateLimiter() != nil {
		engine.Use(middleware.RateLimitByIP(identity.RateLimiter(), cfg.RateLimit))
		lg.Info("rate limiting diaktifkan")
	}

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if cfg.App.Env == "development" {
		lg.Info("mode development aktif")
	}

	identity.RegisterRoutes(engine)
	kesantrian.RegisterRoutes(engine)
	psb.RegisterRoutes(engine)
	dokumenAset.RegisterRoutes(engine)
	article.RegisterRoutes(engine)
	keuangan.RegisterRoutes(engine)
	feedback.RegisterRoutes(engine)
	fingerprint.RegisterRoutes(engine)
	akademik.RegisterRoutes(engine)

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		lg.Info("server started",
			slog.String("addr", "http://0.0.0.0:"+cfg.App.Port),
			slog.String("env", cfg.App.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	lg.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		lg.Error("shutdown error", slog.Any("error", err))
		os.Exit(1)
	}
	lg.Info("server stopped")
}
