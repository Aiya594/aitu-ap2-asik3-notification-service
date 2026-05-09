package jobqueue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	maxRetries = 3
)

type GatewayRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Channel        string `json:"channel"`
	Recipient      string `json:"recipient"`
	Message        string `json:"message"`
}

type WorkerPool struct {
	jobs        chan Job
	workerCount int
	redisClient *redis.Client
	gatewayURL  string
	logger      *slog.Logger
	wg          sync.WaitGroup
}

func NewWorkerPool(redisClient *redis.Client, logger *slog.Logger) *WorkerPool {
	workerCount := 3
	if v := os.Getenv("WORKER_POOL_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			workerCount = n
		}
	}

	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080"
	}

	return &WorkerPool{
		jobs:        make(chan Job, 100),
		workerCount: workerCount,
		redisClient: redisClient,
		gatewayURL:  gatewayURL,
		logger:      logger,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.runWorker(ctx, i)
	}
	wp.logger.Info("worker pool started", "workers", wp.workerCount)
}

func (wp *WorkerPool) Stop() {
	close(wp.jobs)
	wp.wg.Wait()
}

func (wp *WorkerPool) Enqueue(ctx context.Context, job Job) {
	// Check idempotency before enqueuing
	if wp.redisClient != nil {
		processed, err := IsProcessed(ctx, wp.redisClient, job.IdempotencyKey)
		if err != nil {
			wp.logger.Error("idempotency check failed", "error", err, "job_id", job.IdempotencyKey)
		}
		if processed {
			wp.logger.Info("duplicate job dropped",
				"time", time.Now().UTC().Format(time.RFC3339),
				"level", "info",
				"job_id", job.IdempotencyKey,
				"status", "dropped_duplicate",
			)
			return
		}
	}

	wp.logTransition(job.IdempotencyKey, 0, "enqueued", "")

	select {
	case wp.jobs <- job:
	default:
		wp.logger.Warn("job queue full, dropping job", "job_id", job.IdempotencyKey)
	}
}

func (wp *WorkerPool) runWorker(ctx context.Context, workerID int) {
	defer wp.wg.Done()
	wp.logger.Info("worker started", "worker_id", workerID)

	for job := range wp.jobs {
		wp.processWithRetry(ctx, job)
	}

	wp.logger.Info("worker stopped", "worker_id", workerID)
}

func (wp *WorkerPool) processWithRetry(ctx context.Context, job Job) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		wp.logTransition(job.IdempotencyKey, attempt, "processing", "")

		err := wp.callGateway(ctx, job)
		if err == nil {
			// Success — mark idempotency key
			if wp.redisClient != nil {
				if merr := MarkProcessed(ctx, wp.redisClient, job.IdempotencyKey); merr != nil {
					wp.logger.Error("failed to mark job processed", "error", merr, "job_id", job.IdempotencyKey)
				}
			}
			wp.logTransition(job.IdempotencyKey, attempt, "success", "")
			return
		}

		lastErr = err
		if attempt < maxRetries {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			wp.logTransition(job.IdempotencyKey, attempt, "retry", err.Error())
			wp.logger.Warn("job retry",
				"job_id", job.IdempotencyKey,
				"attempt", attempt,
				"backoff_seconds", backoff.Seconds(),
				"error", err.Error(),
			)
			time.Sleep(backoff)
		}
	}

	// Dead-letter
	wp.writeDeadLetter(job, lastErr)
}

func (wp *WorkerPool) callGateway(ctx context.Context, job Job) error {
	payload := GatewayRequest{
		IdempotencyKey: job.IdempotencyKey,
		Channel:        job.Channel,
		Recipient:      job.Recipient,
		Message:        job.Message,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, wp.gatewayURL+"/notify", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("request creation error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("gateway unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("gateway returned 503")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected gateway status: %d", resp.StatusCode)
	}

	return nil
}

type jobLog struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	JobID   string `json:"job_id"`
	Attempt int    `json:"attempt"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

func (wp *WorkerPool) logTransition(jobID string, attempt int, status, errMsg string) {
	entry := jobLog{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Level:   "info",
		JobID:   jobID,
		Attempt: attempt,
		Status:  status,
		Error:   errMsg,
	}
	if errMsg != "" {
		entry.Level = "warn"
	}
	data, _ := json.Marshal(entry)
	fmt.Println(string(data))
}

func (wp *WorkerPool) writeDeadLetter(job Job, err error) {
	entry := jobLog{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Level:   "error",
		JobID:   job.IdempotencyKey,
		Attempt: maxRetries,
		Status:  "dead_letter",
		Error:   err.Error(),
	}
	data, _ := json.Marshal(entry)
	fmt.Fprintln(os.Stderr, string(data))
	wp.logger.Error("job moved to dead-letter", "job_id", job.IdempotencyKey, "error", err.Error())
}
