package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"crypto-gateway/internal/webpush"
)

func SendPushNotifications(formulasID []int, message string) error {
	placeholders := make([]string, len(formulasID))
	args := make([]interface{}, len(formulasID))
	for i, id := range formulasID {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	rows, err := DB.Query(context.Background(), `
		SELECT id, owner_id, name, last_triggered, cooldown
		FROM trigger_formula
		WHERE id IN (%s)
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	type Formula struct {
		ID            int
		OwnerID       int
		Name          string
		LastTriggered *time.Time
		Cooldown      int
	}

	now := time.Now()
	grouped := make(map[int][]Formula)

	for rows.Next() {
		var f Formula
		err := rows.Scan(&f.ID, &f.OwnerID, &f.Name, &f.LastTriggered, &f.Cooldown)
		if err != nil {
			return err
		}

		if f.LastTriggered != nil {
			nextAvailable := f.LastTriggered.Add(time.Duration(f.Cooldown) * time.Second)
			if now.Before(nextAvailable) {
				continue // cooldown не прошёл
			}
		}

		grouped[f.OwnerID] = append(grouped[f.OwnerID], f)
	}

	// разделение на одиночные и множественные
	singleTriggers := make(map[int]Formula)
	multiTriggers := make(map[int][]Formula)

	for ownerID, formulas := range grouped {
		if len(formulas) == 1 {
			singleTriggers[ownerID] = formulas[0]
		} else {
			multiTriggers[ownerID] = formulas
		}
	}

	for userID, formulas := range grouped {
		var payload string

		if len(formulas) == 1 {
			payload = fmt.Sprintf("Сработала стратегия: %s", formulas[0].Name)
		} else {
			payload = "980"
		}

		rows, err := DB.Query(context.Background(), `
            SELECT endpoint, p256dh, auth
            FROM trigger_push_subscription
            WHERE user_id = $1
        `, userID)
		if err != nil {
			log.Printf("ошибка получения push-подписок для user %d: %v", userID, err)
			continue
		}

		for rows.Next() {
			var endpoint, p256dh, auth string
			if err := rows.Scan(&endpoint, &p256dh, &auth); err != nil {
				log.Printf("ошибка сканирования подписки user %d: %v", userID, err)
				continue
			}

			err := webpush.SendWebPush(endpoint, p256dh, auth, payload)
			if err != nil {
				log.Printf("ошибка отправки пуша user %d: %v", userID, err)
			}
		}

		rows.Close()
	}

	updateLastTriggered(formulasID)

	return nil
}

// обновляет в самом конце модели где сработали формулы
func updateLastTriggered(ids []int) {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	_, err := DB.Exec(context.Background(), `
		UPDATE trigger_formula SET last_triggered = NOW()
		WHERE id IN (%s)
	`, args...)

	if err != nil {
		log.Printf("Ошибка при обновлении last_triggered: %v", err)
	}
}

func SaveSubscription(endpoint string, p256dh string, auth string) error {
	now := time.Now()
	_, err := DB.Exec(context.Background(), `
			INSERT INTO trigger_push_subscription (endpoint, p256dh, auth, created_at)
			VALUES ($1, $2, $3, $4)
		`, endpoint, p256dh, auth, now)

	if err != nil {
		return err
	}

	return nil
}
