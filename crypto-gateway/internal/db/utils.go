package db

import "context"

func GetIdbyUuid(uuid string) (string, error) {
	var id string

	err := DB.QueryRow(context.Background(), `
		SELECT id 
		FROM user_user 
		WHERE uuid = $1
	`, uuid).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}
