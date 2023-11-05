package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"os"
	"time"
)

type Redis struct {
	client *redis.Client
}

func NewRedis() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	address := host + ":" + fmt.Sprint(port)

	client := redis.NewClient(&redis.Options{
		Addr: address,
	})

	return client
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

	err = json.Unmarshal([]byte(result), value)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (m *Redis) Set(ctx context.Context, key string, value any) error {
	rawValue, err := json.Marshal(value)

	if err != nil {
		return errors.WithStack(err)
	}

	err = m.client.Set(ctx, key, rawValue, 0).Err()

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (m *Redis) SetEx(ctx context.Context, key string, value any, duration time.Duration) error {
	rawValue, err := json.Marshal(value)

	if err != nil {
		return errors.WithStack(err)
	}

	err = m.client.Set(ctx, key, rawValue, duration).Err()

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
