package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// NewRedis membuat client Redis ke addr yang diberikan,
// lalu memverifikasi koneksi dengan Ping.
func NewRedis(addr, password string) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("gagal melakukan ping ke Redis: %w", err)
	}

	return client, nil
}
