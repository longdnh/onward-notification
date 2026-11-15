package dedup

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "notification:processed:"

type Redis struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedis(addr string, ttl time.Duration) *Redis {
	return &Redis{client: redis.NewClient(
		&redis.Options{Addr: addr}),
		ttl: ttl,
	}
}

func (r *Redis) Claim(ctx context.Context, id string) (bool, error) {
	return r.client.SetNX(ctx, keyPrefix+id, 1, r.ttl).Result()
}

func (r *Redis) Release(ctx context.Context, id string) error {
	return r.client.Del(ctx, keyPrefix+id).Err()
}
