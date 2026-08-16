package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"

	articleModule "sipon-be/internal/modules/article"
	dokumenAsetModule "sipon-be/internal/modules/dokumen_aset"
	identityModule "sipon-be/internal/modules/identity"
	kesantrianModule "sipon-be/internal/modules/kesantrian"
	keuanganModule "sipon-be/internal/modules/keuangan"
	notification "sipon-be/internal/modules/notification"
	psbModule "sipon-be/internal/modules/psb"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/database"
	"sipon-be/internal/shared/logger"
	"sipon-be/internal/shared/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal load config: %v", err)
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
	kesantrian := kesantrianModule.NewModule(
		db, redisClient, cfg,
		identity, // *identity.Module satisfies identity.Contract
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	psb := psbModule.NewModule(
		db, cfg,
		identity,      // identity.Contract
		kesantrian,    // kesantrian.Contract (needs CreateSantriFromPendaftaran)
		identity.AuthMiddleware(),
		identity.PrincipalMiddleware(),
	)

	dokumenAset := dokumenAsetModule.NewModule(
		db, cfg,
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

	fcmService, err := notification.NewService(cfg.Firebase)
	if err != nil {
		lg.Warn("FCM service not available", slog.Any("error", err))
	} else {
		engine.POST("/api/v1/web/notifications/test", func(c *gin.Context) {
			var payload struct {
				Topic string            `json:"topic"`
				Title string            `json:"title"`
				Body  string            `json:"body"`
				Data  map[string]string `json:"data"`
			}

			if err := c.ShouldBindJSON(&payload); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid payload", "error": err.Error()})
				return
			}

			topic := strings.TrimSpace(payload.Topic)
			if topic == "" {
				topic = cfg.Firebase.DefaultTopic
			}
			if title := strings.TrimSpace(payload.Title); title != "" {
				payload.Title = title
			} else {
				payload.Title = "Sipon Notification"
			}
			if body := strings.TrimSpace(payload.Body); body != "" {
				payload.Body = body
			} else {
				payload.Body = "Tes notifikasi dari Sipon"
			}
			if payload.Data == nil {
				payload.Data = map[string]string{"type": "test", "route": "/dashboard"}
			}

			if err := fcmService.Send(c.Request.Context(), notification.SendRequest{
				Topic: topic,
				Title: payload.Title,
				Body:  payload.Body,
				Data:  payload.Data,
			}); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "failed to send FCM notification", "error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "notification sent",
				"topic": topic,
			})
		})
	}

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
