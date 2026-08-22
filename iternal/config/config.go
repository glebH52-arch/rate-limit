package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDb       int
}

func Load() (*Config, error) {

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR  is required")
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))

	if err != nil {
		return nil, fmt.Errorf("REDIS_DB must be nubmers")
	}
	if redisDb < 0 {
		return nil, fmt.Errorf("REDIS_DB must be non-negative")
	}
	return &Config{
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDb:       redisDb,
	}, nil
}
