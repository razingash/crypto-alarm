package repo

import (
	"context"
	"crypto-gateway/internal/web/db"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx"
)

type TelegramNotificationInfo struct {
	Bots     []string `json:"bots"`
	UniqueId string   `json:"element_id"`
	Id       int      `json:"id"`
	Name     string   `json:"name"`
	Token    string   `json:"token"`
	ChatId   string   `json:"chat_id"`
	Signal   bool     `json:"signal"`
	Message  string   `json:"message"`
}

type TelegramBot struct {
	ID     int64  `json:"id,omitempty"`
	Name   string `json:"name"`
	Token  string `json:"token"`
	ChatId string `json:"chat_id"`
}

type TelegramMessage struct {
	ID        int64  `json:"id,omitempty"`
	ElementId string `json:"element_id"`
	BotID     int64  `json:"bot_id,omitempty"`
	Message   string `json:"message"`
	Signal    bool   `json:"signal"`
}

type TelegramNotificationCreate struct {
	BotID   *int            `json:"bot_id,omitempty"`
	Bot     *TelegramBot    `json:"bot,omitempty"`
	Message TelegramMessage `json:"message"`
}

type TelegramNotificationPatch struct {
	BotName   *string `json:"bot_name,omitempty"`
	ElementId *string `json:"element_id,omitempty"`
	Message   *string `json:"message,omitempty"`
	Signal    *bool   `json:"signal,omitempty"`
}

func GetTelegramNotificationInfo(messageId int) (TelegramNotificationInfo, error) {
	var info TelegramNotificationInfo

	rows, err := db.DB.Query(context.Background(), `SELECT name FROM module_notification_telegram_bot`)
	if err != nil {
		return info, err
	}
	defer rows.Close()

	var bots []string
	for rows.Next() {
		var botName string
		if err := rows.Scan(&botName); err != nil {
			return info, err
		}
		bots = append(bots, botName)
	}
	info.Bots = bots

	if messageId == 0 {
		return info, nil
	}

	err = db.DB.QueryRow(context.Background(),
		`SELECT b.name, b.token, b.chat_id, m.id, m.element_id, m.message, m.signal
         FROM module_notification_telegram_message m
         LEFT JOIN module_notification_telegram_bot b ON b.id = m.bot_id
         WHERE m.id = $1
         LIMIT 1`, messageId).Scan(&info.Name, &info.Token, &info.ChatId, &info.Id, &info.UniqueId, &info.Message, &info.Signal)
	if err != nil {
		return info, err
	}

	return info, nil
}

func IsBotExists(botName string) (bool, error) {
	ctx := context.Background()
	var exists bool

	err := db.DB.QueryRow(ctx, `
    SELECT EXISTS(
        SELECT 1
        FROM module_notification_telegram_bot
        WHERE name = $1
	)`, botName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func SaveTelegramNotificationInfo(info TelegramNotificationCreate) (int, error) {
	ctx := context.Background()
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var botID int

	switch {
	case info.BotID != nil:
		botID = *info.BotID
	case info.Bot != nil && info.Bot.Token != "" && info.Bot.ChatId != "":
		err = tx.QueryRow(ctx, `
			INSERT INTO module_notification_telegram_bot (name, token, chat_id)
			VALUES ($1, $2, $3)
			RETURNING id
		`, info.Bot.Name, info.Bot.Token, info.Bot.ChatId).Scan(&botID)
		if err != nil {
			return 0, err
		}
	case info.Bot != nil && info.Bot.Token == "" && info.Bot.ChatId == "":
		err = tx.QueryRow(ctx, `
			SELECT id FROM module_notification_telegram_bot WHERE name = $1
		`, info.Bot.Name).Scan(&botID)
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("bot %s does not exist", info.Bot.Name)
		} else if err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("no valid bot_id or bot info provided")
	}

	var messageID int
	err = tx.QueryRow(ctx, `
		INSERT INTO module_notification_telegram_message (element_id, bot_id, message, signal)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, info.Message.ElementId, botID, info.Message.Message, info.Message.Signal).Scan(&messageID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return messageID, nil
}

func UpdateTelegramNotificationInfo(id int, patch TelegramNotificationPatch) error {
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if patch.BotName != nil {
		var botID int64
		err := db.DB.QueryRow(context.Background(),
			`SELECT id FROM module_notification_telegram_bot WHERE name = $1 LIMIT 1`, *patch.BotName).Scan(&botID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("bot with name %q not found", *patch.BotName)
			}
			return err
		}
		setClauses = append(setClauses, fmt.Sprintf("bot_id = $%d", argIndex))
		args = append(args, botID)
		argIndex++
	}

	if patch.ElementId != nil {
		setClauses = append(setClauses, fmt.Sprintf("element_id = $%d", argIndex))
		args = append(args, *patch.ElementId)
		argIndex++
	}

	if patch.Message != nil {
		setClauses = append(setClauses, fmt.Sprintf("message = $%d", argIndex))
		args = append(args, *patch.Message)
		argIndex++
	}

	if patch.Signal != nil {
		setClauses = append(setClauses, fmt.Sprintf("signal = $%d", argIndex))
		args = append(args, *patch.Signal)
		argIndex++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	args = append(args, id)
	query := fmt.Sprintf(`
        UPDATE module_notification_telegram_message
        SET %s
        WHERE id = $%d
    `, strings.Join(setClauses, ", "), argIndex)

	_, err := db.DB.Exec(context.Background(), query, args...)
	return err
}
