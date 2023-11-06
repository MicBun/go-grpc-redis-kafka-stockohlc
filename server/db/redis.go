package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type Redis struct {
	client *redis.Client
}

type RedisManager interface {
	Get(ctx context.Context, key string, value any) error
	Set(ctx context.Context, key string, value any) error
	SetEx(ctx context.Context, key string, value any, duration time.Duration) error
}

func NewRedis() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	address := host + ":" + fmt.Sprint(port)
	return redis.NewClient(&redis.Options{
		Addr: address,
	})
}

func NewRedisManager(client *redis.Client) *Redis {
	return &Redis{client}
}

func (m *Redis) Get(ctx context.Context, key string, value any) error {
	result, err := m.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return errors.WithStack(errors.New("data not found"))
	} else if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(json.Unmarshal([]byte(result), value))
}

func (m *Redis) Set(ctx context.Context, key string, value any) error {
	rawValue, err := json.Marshal(value)
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(m.client.Set(ctx, key, rawValue, 0).Err())
}

func (m *Redis) SetEx(ctx context.Context, key string, value any, duration time.Duration) error {
	rawValue, err := json.Marshal(value)
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(m.client.Set(ctx, key, rawValue, duration).Err())
}
