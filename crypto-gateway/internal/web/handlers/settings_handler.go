package handlers

import (
	"bufio"
	"crypto-gateway/internal/web/db"
	"encoding/json"
	"os"

	"github.com/gofiber/fiber/v3"
)

func GetSettings(c fiber.Ctx) error {
	settings, err := db.FetchSettings()

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": settings,
	})
}

func PatchUpdateSettings(c fiber.Ctx) error {
	id := c.Locals("id").(int)
	cooldown := c.Locals("cooldown").(int)

	err := db.UpdateCooldown(id, cooldown)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "something went wrong",
		})
	}

	go updateApiCooldown(id)

	return c.SendStatus(fiber.StatusOK)
}

func GetAvailabilityMetrics(c fiber.Ctx) error {
	type AvailabilityMetric struct {
		Timestamp   string `json:"timestamp"`
		Level       string `json:"level"`
		Caller      string `json:"caller"`
		Message     string `json:"message"`
		Type        int    `json:"type"`
		Event       string `json:"event"`
		IsAvailable int    `json:"isAvailable"`
	}

	file, err := os.Open("logs/AvailabilityMetrics.log")
	if err != nil {
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"data": []interface{}{},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to read log file",
		})
	}
	defer file.Close()

	var metrics []AvailabilityMetric
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		var entry AvailabilityMetric
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			metrics = append(metrics, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error reading log file",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": metrics,
	})
}
