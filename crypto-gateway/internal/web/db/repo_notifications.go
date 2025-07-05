package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"crypto-gateway/internal/web/webpush"
)

func SendPushNotifications(formulasID []int, message string) error {
	placeholders := make([]string, len(formulasID))
	args := make([]interface{}, len(formulasID))
	for i, id := range formulasID {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
	    SELECT id, name, last_triggered, cooldown
	    FROM trigger_formula
	    WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := DB.Query(context.Background(), query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	type Formula struct {
		ID            int
		Name          string
		LastTriggered *time.Time
		Cooldown      int
	}

	now := time.Now().UTC().Truncate(time.Second)
	grouped := make(map[int][]Formula)

	for rows.Next() {
		var f Formula
		err := rows.Scan(&f.ID, &f.Name, &f.LastTriggered, &f.Cooldown)
		if err != nil {
			return err
		}

		if f.LastTriggered != nil {
			nextAvailable := f.LastTriggered.UTC().Truncate(time.Second).Add(time.Duration(f.Cooldown) * time.Second)
			if now.Before(nextAvailable) {
				continue // cooldown не прошёл
			}
		}
		grouped[f.ID] = append(grouped[f.ID], f)
	}

	var outdatedSubIDs []int

	for ownerID, formulas := range grouped {
		var data map[string]string
		if len(formulas) == 1 {
			data = map[string]string{
				"title": "Triggered",
				"body":  fmt.Sprintf("Strategy: %s", formulas[0].Name),
			}
		} else {
			data = map[string]string{
				"title": "Triggered",
				"body":  fmt.Sprintf("You have %d triggered strategies", len(formulas)),
			}
		}
		jsonPayload, _ := json.Marshal(data)

		rows, err := DB.Query(context.Background(), `
            SELECT endpoint, p256dh, auth, id
            FROM trigger_push_subscription
            WHERE user_id = $1
        `, ownerID)
		if err != nil {
			log.Printf("ошибка получения push-подписок для user %d: %v", ownerID, err)
			continue
		}

		for rows.Next() {
			var subID int
			var endpoint, p256dh, auth string
			if err := rows.Scan(&endpoint, &p256dh, &auth, &subID); err != nil {
				log.Printf("ошибка сканирования подписки user %d: %v", ownerID, err)
				continue
			}

			err := webpush.SendWebPush(endpoint, p256dh, auth, jsonPayload)
			if err != nil {
				log.Printf("ошибка отправки пуша user %d: %v", ownerID, err)
				outdatedSubIDs = append(outdatedSubIDs, subID)
			}
		}

		rows.Close()
	}

	err = updateLastTriggered(formulasID)
	if err != nil {
		log.Printf("Несовсем критическая ошибка при сохранении ласт апдейта формулы, будет критичной когда добавится история триггеров")
	}

	if len(outdatedSubIDs) > 0 {
		if err := deleteOutdatedSubscription(outdatedSubIDs); err != nil {
			log.Printf("ошибка при удалении устаревших подписок: %v", err)
		}
	}

	return nil
}

// обновляет в самом конце модели где сработали формулы
func updateLastTriggered(ids []int) error {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		UPDATE trigger_formula SET last_triggered = NOW() AT TIME ZONE 'UTC'
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	_, err := DB.Exec(context.Background(), query, args...)
	if err != nil {
		return err
	}
	return nil
}

func SaveSubscription(endpoint string, p256dh string, auth string) error {
	now := time.Now().UTC()
	_, err := DB.Exec(context.Background(), `
    INSERT INTO trigger_push_subscription (endpoint, p256dh, auth, created_at)
    VALUES ($1, $2, $3, $4)
    ON CONFLICT (endpoint) DO NOTHING
	`, endpoint, p256dh, auth, now)

	if err != nil {
		return err
	}

	return nil
}

func deleteOutdatedSubscription(subscriptionIds []int) error {
	if len(subscriptionIds) == 0 {
		return nil
	}

	placeholders := make([]string, len(subscriptionIds))
	args := make([]interface{}, len(subscriptionIds))
	for i, id := range subscriptionIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`DELETE FROM trigger_push_subscription WHERE id IN (%s)`, strings.Join(placeholders, ","))

	_, err := DB.Exec(context.Background(), query, args...)
	return err
}
