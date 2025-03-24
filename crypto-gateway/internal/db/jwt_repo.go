package db

import (
	"context"
	"time"
)

/*
All data should be validated before calling this functions
*/

func CreateUser(username string, password string) (string, error) {
	var userUUID string
	err := DB.QueryRow(context.Background(), `
        INSERT INTO user_user (uuid, username, password, registered_date, ispremium) 
        VALUES (gen_random_uuid(), $1, $2, $3, $4) 
        RETURNING uuid`,
		username, password, time.Now(), false,
	).Scan(&userUUID)

	if err != nil {
		return "", err
	}
	return userUUID, nil
}

func SaveAccessToken(uuid string, accessToken string) error {
	_, err := DB.Exec(context.Background(), `
		INSERT INTO access_tokens (user_uuid, token, expires_at, created_at) 
		VALUES ($1, $2, $3, $4)`,
		uuid, accessToken, time.Now().Add(15*time.Minute), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func SaveRefreshToken(uuid string, refreshToken string) error {
	_, err := DB.Exec(context.Background(), `
		INSERT INTO refresh_tokens (user_uuid, token, expires_at, created_at, revoked) 
		VALUES ($1, $2, $3, $4, false)`,
		uuid, refreshToken, time.Now().Add(24*time.Hour), time.Now())
	if err != nil {
		return err
	}
	return nil
}
