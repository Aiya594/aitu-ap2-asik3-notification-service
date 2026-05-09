package jobqueue

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const idempotencyTTL = 24 * time.Hour

// IsProcessed checks whether a job with the given key has already been processed.
func IsProcessed(ctx context.Context, client *redis.Client, key string) (bool, error) {
	val, err := client.Get(ctx, "idempotency:"+key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return val == "done", nil
}

// MarkProcessed stores the idempotency key in Redis with a 24h TTL.
func MarkProcessed(ctx context.Context, client *redis.Client, key string) error {
	return client.Set(ctx, "idempotency:"+key, "done", idempotencyTTL).Err()
}
