package db

import (
	"crypto-gateway/crypto-gateway/config"
	"database/sql"
	"fmt"
	"log"
)

var (
	DB *sql.DB
)

func InitDB() {
	var err error
	DB, err = sql.Open("postgres", config.Database_Url)
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Ошибка пинга базы данных: %v", err)
	}

	fmt.Println("Успешное подключение к PostgreSQL!")
}
