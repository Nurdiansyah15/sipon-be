package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
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
	messagingModule "sipon-be/internal/modules/messaging"
	msgApp "sipon-be/internal/modules/messaging/application"
	outboxEntity "sipon-be/internal/modules/messaging/domain/event_outbox/entity"
	outboxRepo "sipon-be/internal/modules/messaging/domain/event_outbox/repository"
	outboxPersistence "sipon-be/internal/modules/messaging/infrastructure/persistence"
	"sipon-be/internal/modules/messaging/interfaces/rabbitmq"
	psbModule "sipon-be/internal/modules/psb"
	schedulerModule "sipon-be/internal/modules/scheduler"
	"sipon-be/internal/modules/scheduler/application"
	schedulerPorts "sipon-be/internal/modules/scheduler/application/ports"
	"sipon-be/internal/shared/config"
	"sipon-be/internal/shared/database"
	"sipon-be/internal/shared/logger"
	"sipon-be/internal/shared/timeutil"
)

// outboxWriterAdapter membalik arah dependensi scheduler->messaging: scheduler
// hanya tahu port OutboxWriter, adapter ini menghubungkannya ke repository outbox
// milik messaging (composition root).
type outboxWriterAdapter struct {
	repo outboxRepo.Repository
}

func (a *outboxWriterAdapter) Save(ctx context.Context, routingKey string, payload json.RawMessage) error {
	return a.repo.Save(ctx, outboxEntity.NewOutboxEntry(routingKey, payload, ""))
}

var _ schedulerPorts.OutboxWriter = (*outboxWriterAdapter)(nil)

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
	scheduler := schedulerModule.NewModule(db,
		time.Duration(cfg.Worker.TickSeconds)*time.Second,
		time.Duration(cfg.Worker.LeaseSeconds)*time.Second,
		lg)
	akademik := akademikModule.NewModule(db, cfg,
		kesantrian, fingerprint, scheduler,
		identity.AuthMiddleware(), identity.PrincipalMiddleware())
	kesantrian.SetAkademikProvisioner(akademik)

	// Registrasi handler asynchronous via messaging.Contract, sama seperti modul
	// lain diintegrasikan lewat Contract-nya masing-masing (mis. scheduler.Contract).
	// Setiap module memanggil RegisterMessageHandlers; cmd/worker hanya composition
	// + kumpulkan binding queue.
	msgModule := messagingModule.NewModule(5)
	var bindings []messagingModule.Binding

	register := func(name string, reg func(messagingModule.Contract) ([]messagingModule.Binding, error)) {
		b, err := reg(msgModule)
		if err != nil {
			lg.Error("gagal registrasi message handlers", slog.String("module", name), slog.Any("error", err))
			os.Exit(1)
		}
		bindings = append(bindings, b...)
	}

	register("identity", identity.RegisterMessageHandlers)
	register("dokumen_aset", dokumenAset.RegisterMessageHandlers)
	register("kesantrian", kesantrian.RegisterMessageHandlers)
	register("psb", psb.RegisterMessageHandlers)
	register("article", article.RegisterMessageHandlers)
	register("keuangan", keuangan.RegisterMessageHandlers)
	register("feedback", feedback.RegisterMessageHandlers)
	register("fingerprint", fingerprint.RegisterMessageHandlers)
	register("akademik", akademik.RegisterMessageHandlers)

	dispatcher := scheduler.Dispatcher()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outboxRepo := outboxPersistence.NewPostgresOutboxRepository(db)
	transactor := database.NewTransactor(db)
	messageJobRepo := outboxPersistence.NewPostgresMessageJobRepository(db)

	metrics := &msgApp.Metrics{}
	useOutbox := cfg.Worker.Mode == "outbox"
	if useOutbox && !cfg.RabbitMQ.Enabled {
		lg.Error("WORKER_MODE=outbox membutuhkan RABBITMQ_ENABLED=true")
		os.Exit(1)
	}

	var wg sync.WaitGroup

	// Deklarasi RabbitMQ topology (exchange, queue per role, retry, DLQ) bila
	// RABBITMQ_ENABLED=true. Idempotent dan durable terhadap restart broker.
	if cfg.RabbitMQ.Enabled {
		if err := declareRabbitMQTopology(cfg, bindings, lg); err != nil {
			lg.Warn("gagal declare RabbitMQ topology", slog.Any("error", err))
		}
	}

	if useOutbox {
		// Mode outbox: Scheduler Dispatcher hanya menulis event_outbox; eksekusi
		// handler dilakukan oleh Message Consumer melalui RabbitMQ.
		dispatcher.WithOutboxMode(&outboxWriterAdapter{repo: outboxRepo}, transactor)

		publisher, err := rabbitmq.NewPublisher(cfg.RabbitMQ.DSN, cfg.RabbitMQ.Exchange, cfg.RabbitMQ.PublishTimeout)
		if err != nil {
			lg.Error("gagal init RabbitMQ publisher", slog.Any("error", err))
			os.Exit(1)
		}
		defer publisher.Close()
		msgModule.WithPublisher(publisher)

		// Outbox Relay: event_outbox -> RabbitMQ (dengan publisher confirm).
		outboxRelay := msgApp.NewOutboxRelay(outboxRepo, publisher, msgApp.OutboxRelayOptions{
			Interval:  2 * time.Second,
			Lease:     30 * time.Second,
			BaseDelay: 30 * time.Second,
			MaxDelay:  30 * time.Minute,
		}, lg).WithMetrics(metrics)
		wg.Add(1)
		go func() {
			defer wg.Done()
			outboxRelay.Start(ctx)
		}()

		// Message Consumer: RabbitMQ -> message_jobs -> module handler.
		consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQ.DSN, cfg.RabbitMQ.Prefetch)
		if err != nil {
			lg.Error("gagal init RabbitMQ consumer", slog.Any("error", err))
			os.Exit(1)
		}
		defer consumer.Close()

		msgConsumer := msgApp.NewMessageConsumer(
			consumer,
			messageJobRepo,
			msgModule.Registry(),
			msgModule.RetryPolicy(),
			publisher,
			msgApp.MessageConsumerOptions{
				Lease:       5 * time.Minute,
				BaseDelay:   30 * time.Second,
				MaxDelay:    30 * time.Minute,
				RetryDelays: cfg.RabbitMQ.RetryDelays,
			},
			lg,
		).WithMetrics(metrics)

		queues := map[string]struct{}{}
		for _, b := range bindings {
			queues[b.Queue] = struct{}{}
		}
		for q := range queues {
			queue := q
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := msgConsumer.Start(ctx, queue); err != nil {
					lg.Error("consumer stopped", slog.String("queue", queue), slog.Any("error", err))
				}
			}()
		}

		startHealthServer(cfg, db, publisher, metrics, lg)
	} else {
		// Mode direct: compatibility bridge sampai outbox/RabbitMQ consumer aktif.
		dispatcher.WithDirectMode(func(ctx context.Context, jobType string, payload json.RawMessage) error {
			msg, err := messagingModule.NewMessage(jobType, payload)
			if err != nil {
				return &application.FatalError{Err: err}
			}
			if err := msgModule.Dispatch(ctx, msg); err != nil {
				if messagingModule.IsFatal(err) {
					return &application.FatalError{Err: err}
				}
				return err
			}
			return nil
		})
		startHealthServer(cfg, db, nil, metrics, lg)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		dispatcher.Run(ctx)
	}()

	lg.Info("worker started, waiting for scheduled jobs")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	lg.Info("shutdown signal received")
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		lg.Info("all goroutines stopped gracefully")
	case <-time.After(10 * time.Second):
		lg.Warn("shutdown timeout; force exit")
	}
	lg.Info("worker stopped")
}

// startHealthServer menyediakan endpoint /healthz (DB + RabbitMQ) dan /metrics
// pada port terpisah dari proses worker.
func startHealthServer(
	cfg *config.Config,
	db *sql.DB,
	publisher *rabbitmq.RabbitMQPublisher,
	metrics *msgApp.Metrics,
	lg *slog.Logger,
) {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		dbErr := db.PingContext(r.Context())
		status := "ok"
		dbStatus := "ok"
		if dbErr != nil {
			dbStatus = dbErr.Error()
			status = "degraded"
		}

		rabbitStatus := "disabled"
		if publisher != nil {
			if err := publisher.Ping(r.Context()); err != nil {
				rabbitStatus = err.Error()
				status = "degraded"
			} else {
				rabbitStatus = "ok"
			}
		}

		if status != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		writeJSON(w, map[string]any{
			"status":   status,
			"database": dbStatus,
			"rabbitmq": rabbitStatus,
		})
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, metrics.Snapshot())
	})

	srv := &http.Server{Addr: ":" + strconv.Itoa(cfg.Worker.HealthPort), Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Warn("health server error", slog.Any("error", err))
		}
	}()
	lg.Info("health server started", slog.Int("port", cfg.Worker.HealthPort))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// declareRabbitMQTopology menghubungkan ke broker dan mendeklarasikan topology
// durable berdasarkan binding yang dikumpulkan dari module.
func declareRabbitMQTopology(cfg *config.Config, bindings []messagingModule.Binding, lg *slog.Logger) error {
	topo, err := rabbitmq.NewTopology(rabbitmq.Options{
		DSN:         cfg.RabbitMQ.DSN,
		Exchange:    cfg.RabbitMQ.Exchange,
		DLXExchange: cfg.RabbitMQ.DLXExchange,
		RetryDelays: cfg.RabbitMQ.RetryDelays,
	})
	if err != nil {
		return err
	}
	defer topo.Close()

	if err := topo.Declare(bindings); err != nil {
		return err
	}
	lg.Info("RabbitMQ topology declared",
		slog.String("exchange", cfg.RabbitMQ.Exchange),
		slog.Int("bindings", len(bindings)),
	)
	return nil
}
