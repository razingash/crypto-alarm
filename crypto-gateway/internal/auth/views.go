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

func RegisterUser(username string, password string) (*User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	var userUUID string

	err2 := db.DB.QueryRow(context.Background(), `
		INSERT INTO user_user (uuid, username, password, registered_date, ispremium) 
		VALUES (gen_random_uuid(), $1, $2, $3, $4) 
		RETURNING uuid`, username, hashedPassword, time.Now(), false).Scan(&userUUID)
	if err2 != nil {
		return nil, err2
	}

	return &User{UUID: userUUID}, nil
}
