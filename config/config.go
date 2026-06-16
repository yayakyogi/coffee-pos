package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from the environment.
type Config struct {
	// App
	AppPort string
	AppEnv  string

	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// Midtrans
	MidtransServerKey string
	MidtransClientKey string
	MidtransEnv       string
}

// Load reads configuration from a .env file (if present) and the process
// environment, validates required fields, and returns the populated Config.
//
// A missing .env file is not an error: in production the environment is
// typically injected by the orchestrator rather than a file.
func Load() (*Config, error) {
	// Ignore the error: .env is optional. Required fields are validated below.
	_ = godotenv.Load()

	cfg := &Config{
		AppPort: os.Getenv("APP_PORT"),
		AppEnv:  os.Getenv("APP_ENV"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),

		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),

		JWTSecret: os.Getenv("JWT_SECRET"),

		MidtransServerKey: os.Getenv("MIDTRANS_SERVER_KEY"),
		MidtransClientKey: os.Getenv("MIDTRANS_CLIENT_KEY"),
		MidtransEnv:       os.Getenv("MIDTRANS_ENV"),
	}

	// Validate required fields.
	required := map[string]string{
		"DB_HOST":     cfg.DBHost,
		"DB_NAME":     cfg.DBName,
		"DB_USER":     cfg.DBUser,
		"DB_PASSWORD": cfg.DBPassword,
		"JWT_SECRET":  cfg.JWTSecret,
	}
	for name, value := range required {
		if value == "" {
			return nil, fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	// JWTExpiryHours defaults to 24 when unset.
	cfg.JWTExpiryHours = 24
	if raw := os.Getenv("JWT_EXPIRY_HOURS"); raw != "" {
		hours, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS %q: %w", raw, err)
		}
		cfg.JWTExpiryHours = hours
	}

	return cfg, nil
}

// MysqlDSN returns the MySQL connection string for the configured database.
func (c *Config) MysqlDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// RedisAddr returns the Redis address in host:port form.
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}
