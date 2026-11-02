package ledger

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitCache() error {
	dbNumber, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return fmt.Errorf("REDIS_DB должен быть числом: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbNumber,
	})

	err = client.Ping(context.Background()).Err()
	if err != nil {
		return fmt.Errorf("Redis не отвечает: %w", err)
	}

	rdb = client
	log.Println("Подключились к Redis")
	return nil
}
