package db

import (
	"crypto-gateway/crypto-gateway/config"
	"fmt"
	"log"

	"context"

	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

func InitDB() {
	var err error

	connConfig, err := pgx.ParseConfig(config.Database_Url)
	if err != nil {
		log.Fatalf("Ошибка при парсинге строки подключения: %v", err)
	}

	DB, err = pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}

	err = DB.Ping(context.Background())
	if err != nil {
		log.Fatalf("Ошибка пинга базы данных: %v", err)
	}

	fmt.Println("Успешное подключение к PostgreSQL!")
}
