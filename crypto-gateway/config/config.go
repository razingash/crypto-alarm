package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	SecretKey           string
	Vapid_Public_Key    string
	Vapid_Private_Key   string
	Database_Url        string
	Internal_Server_Api string
)

func LoadConfig() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Application startup via docker")
	}

	SecretKey = os.Getenv("SECRET_KEY")
	Vapid_Public_Key = os.Getenv("VAPID_PUBLIC_KEY")
	Vapid_Private_Key = os.Getenv("VAPID_PRIVATE_KEY")
	var Database_Name = os.Getenv("DB_NAME")
	var Database_User = os.Getenv("DB_USER")
	var Database_Password = os.Getenv("DB_PASSWORD")
	var Database_Host = os.Getenv("DB_HOST")
	var Database_Port = os.Getenv("DB_PORT")
	var Internal_Server_Api = os.Getenv("INTERNAL_SERVER_API")

	// учитывая как тут работает кэш коллектор и глобальные окружения лучше в ручну проверять(по крайней мере на наличие)
	if Internal_Server_Api == "" {
		Internal_Server_Api = "http://127.0.0.1:8000/api/v1"
	}
	if SecretKey == "" {
		log.Fatal("SecretKey не установлена в окружении")
	}
	if Vapid_Public_Key == "" {
		log.Fatal("Vapid_Public_Key не установлена в окружении")
	}
	if Vapid_Private_Key == "" {
		log.Fatal("Vapid_Private_Key не установлена в окружении")
	}
	if Database_Name == "" {
		log.Fatal("Database_Name не установлена в окружении")
	}
	if Database_User == "" {
		log.Fatal("Database_User не установлена в окружении")
	}
	if Database_Password == "" {
		log.Fatal("Database_Password не установлена в окружении")
	}
	if Database_Host == "" {
		log.Fatal("Database_Host не установлена в окружении")
	}
	if Database_Port == "" {
		log.Fatal("Database_Port не установлена в окружении")
	}

	// сейчас без ORM, поэтому не будет возможности подключится асинхронно, в самом конце сделать асинхронное подключение
	// чтобы сверить скорости
	Database_Url = fmt.Sprintf(
		"postgresql://%v:%v@%v:%v/%v?sslmode=disable",
		Database_User, Database_Password, Database_Host, Database_Port, Database_Name,
	)
}
