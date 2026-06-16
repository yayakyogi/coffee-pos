package main

import (
	"fmt"
	"log"

	"github.com/yayakyogi/coffee-pos/config"
	"github.com/yayakyogi/coffee-pos/internal/handler"
	"github.com/yayakyogi/coffee-pos/pkg/database"
	redispkg "github.com/yayakyogi/coffee-pos/pkg/redis"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run wires up dependencies and starts the application. It is separated from
// main so that deferred cleanup (db.Close, rdb.Close) actually runs: log.Fatal
// in main calls os.Exit, which would otherwise skip any deferred calls.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.NewMySQL(cfg.MysqlDSN())
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer db.Close()

	rdb, err := redispkg.NewRedis(cfg.RedisAddr(), cfg.RedisPassword)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer rdb.Close()

	fmt.Println("MySQL connected.")
	fmt.Println("Redis connected.")
	fmt.Println("Starting server on port :" + cfg.AppPort)

	r := handler.NewRouter(cfg.AppEnv)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
