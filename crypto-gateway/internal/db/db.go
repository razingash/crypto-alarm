package db

import (
	"context"
	"crypto-gateway/config"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	var err error

	config, err := pgxpool.ParseConfig(config.Database_Url)
	if err != nil {
		log.Fatalf("Ошибка при разборе строки подключения: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 30 * time.Minute

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = DB.Ping(ctx)
	if err != nil {
		log.Fatalf("Ошибка пинга базы данных: %v", err)
	}

	fmt.Println("Успешное подключение к PostgreSQL!")
}
