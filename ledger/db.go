package ledger

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func databaseURL() string {
	url := os.Getenv("DATABASE_URL")
	if url != "" {
		return url
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASS", "postgres")
	name := getEnv("DB_NAME", "cashapp")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
}

func InitDB() error {
	conn, err := sql.Open("pgx", databaseURL())
	if err != nil {
		return fmt.Errorf("не удалось открыть подключение к базе: %w", err)
	}

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)

	err = conn.Ping()
	if err != nil {
		conn.Close()
		return fmt.Errorf("база данных не отвечает: %w", err)
	}

	db = conn
	log.Println("Подключились к PostgreSQL")
	return nil
}
