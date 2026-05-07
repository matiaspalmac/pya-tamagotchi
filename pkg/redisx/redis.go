package redisx

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func MustOpen() *redis.Client {
	addr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	c := redis.NewClient(&redis.Options{Addr: addr})
	if err := c.Ping(context.Background()).Err(); err != nil {
		panic(fmt.Errorf("redis ping: %w", err))
	}
	return c
}
