package auth

import (
	"context"
	"crypto-gateway/crypto-gateway/internal/db"
	"database/sql"
	"time"
)

const (
	ErrCodeUserNotFound    = 1
	ErrCodeDBError         = 2
	ErrCodeInvalidPassword = 3
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

func LoginUser(username string, password string) (int, *User) {
	var userUUID string
	var userPassword string

	err := db.DB.QueryRow(context.Background(), `
		SELECT uuid, password
		FROM user_user 
		WHERE username = $1
	`, username).Scan(&userUUID, &userPassword)

	if err == sql.ErrNoRows {
		return ErrCodeUserNotFound, nil
	} else if err != nil { // не найден пользователь с таким логином
		return ErrCodeDBError, nil
	}

	if !CheckPassword(password, userPassword) {
		return ErrCodeInvalidPassword, nil
	}

	return 0, &User{UUID: userUUID}
}
