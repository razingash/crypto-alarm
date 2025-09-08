package repo

import (
	"context"
	"crypto-gateway/internal/web/db"
	"fmt"
	"strings"
)

type TelegramNotificationInfo struct {
	Token     string `json:"token"`
	ChatId    string `json:"chat_id"`
	Cooldown  int    `json:"cooldown"`
	Condition bool   `json:"condition"`
	Message   string `json:"message"`
}

type TelegramNotificationPatch struct {
	Token     *string `json:"token,omitempty"`
	ChatId    *string `json:"chat_id,omitempty"`
	Cooldown  *int    `json:"cooldown,omitempty"`
	Condition *bool   `json:"condition,omitempty"`
	Message   *string `json:"message,omitempty"`
}

func GetTelegramNotificationInfo(telegramId int) (TelegramNotificationInfo, error) {
	var telegramInfo TelegramNotificationInfo
	err := db.DB.QueryRow(context.Background(), `
        SELECT token, chat_id, cooldown, condition, messages FROM module_notification WHERE id = $1
    `, telegramId).Scan(&telegramInfo.Token, &telegramInfo.ChatId, &telegramInfo.Cooldown, &telegramInfo.Condition,
		&telegramInfo.Message)
	if err != nil {
		return telegramInfo, err
	}
	return telegramInfo, nil
}

func SaveTelegramNotificationInfo(info TelegramNotificationInfo) error {
	// при сохранении проверять что будет доступ ко всем переннымы
	_, err := db.DB.Exec(context.Background(), `
		INSERT INTO module_notification (token, chat_id, cooldown, message, condition) 
		VALUES ($1, $2, $3, $4, $5)
	`, info.Token, info.ChatId, info.Cooldown, info.Condition, info.Message)
	if err != nil {
		return err
	}
	return nil
}

func UpdateTelegramNotificationInfo(id int, patch TelegramNotificationPatch) error {
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if patch.Token != nil {
		setClauses = append(setClauses, fmt.Sprintf("token = $%d", argIndex))
		args = append(args, *patch.Token)
		argIndex++
	}

	if patch.ChatId != nil {
		setClauses = append(setClauses, fmt.Sprintf("chat_id = $%d", argIndex))
		args = append(args, *patch.ChatId)
		argIndex++
	}

	if patch.Cooldown != nil {
		setClauses = append(setClauses, fmt.Sprintf("cooldown = $%d", argIndex))
		args = append(args, *patch.Cooldown)
		argIndex++
	}

	if patch.Condition != nil {
		setClauses = append(setClauses, fmt.Sprintf("condition = $%d", argIndex))
		args = append(args, *patch.Condition)
		argIndex++
	}

	if patch.Message != nil {
		setClauses = append(setClauses, fmt.Sprintf("message = $%d", argIndex))
		args = append(args, *patch.Message)
		argIndex++
	}

	if len(setClauses) == 0 {
		return nil
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE module_notification
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	_, err := db.DB.Exec(context.Background(), query, args...)
	return err
}
