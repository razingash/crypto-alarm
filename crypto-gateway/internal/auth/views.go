package auth

import (
	"context"
	"crypto-gateway/crypto-gateway/internal/db"
	"time"
)

type User struct {
	UUID     string
	Password string
}

func RegisterUser(username, password string) (*User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	var userUUID string

	err2 := db.DB.QueryRow(context.Background(), `
		INSERT INTO user_user (uuid, password, created_at) 
		VALUES (gen_random_uuid(), $1, $2) 
		RETURNING uuid`, hashedPassword, time.Now()).Scan(&userUUID)
	if err2 != nil {
		return nil, err2
	}

	return &User{UUID: userUUID}, nil
}
