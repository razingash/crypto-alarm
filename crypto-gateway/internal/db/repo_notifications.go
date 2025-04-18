package db

import (
	"context"
	"fmt"
	"log"
	"time"
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

	// Передаём дальше или возвращаем
	log.Println("единичные случаи:")
	for id, f := range singleTriggers {
		log.Printf("user %d -> %s", id, f.Name)
	}

	log.Println("множественные случаи:")
	for id, fs := range multiTriggers {
		names := []string{}
		for _, f := range fs {
			names = append(names, f.Name)
		}
		log.Printf("user %d -> %v", id, names)
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
