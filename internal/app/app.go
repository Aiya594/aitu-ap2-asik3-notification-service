package app

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/broker"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/config"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/jobqueue"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/logger"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/subscrider"
	"github.com/redis/go-redis/v9"
)

type App struct {
	sub    *subscrider.Subscriber
	broker *broker.Client
	pool   *jobqueue.WorkerPool
	redis  *redis.Client
}

func New() *App {
	cfg := config.LoadCfg()

	slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := broker.New(cfg.NatsURL)
	if err != nil {
		log.Fatal("cannot connect to NATS:", err)
	}

	logg := logger.New()

	// Redis — best-effort
	var redisClient *redis.Client
	redisURL := cfg.RedisURL
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	opt, parseErr := redis.ParseURL(redisURL)
	if parseErr == nil {
		rc := redis.NewClient(opt)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if pingErr := rc.Ping(ctx).Err(); pingErr == nil {
			redisClient = rc
			slogger.Info("notification-service: redis connected")
		} else {
			slogger.Warn("notification-service: redis unreachable, idempotency disabled", "error", pingErr)
			rc.Close()
		}
	} else {
		slogger.Warn("notification-service: invalid REDIS_URL, idempotency disabled", "error", parseErr)
	}

	pool := jobqueue.NewWorkerPool(redisClient, slogger)

	sub := subscrider.New(client.Conn, logg, pool)

	return &App{
		sub:    sub,
		broker: client,
		pool:   pool,
		redis:  redisClient,
	}
}

func (a *App) Run() error {
	ctx := context.Background()
	a.pool.Start(ctx)
	return a.sub.Start()
}

func (a *App) Close() {
	a.pool.Stop()
	if a.redis != nil {
		a.redis.Close()
	}
}
