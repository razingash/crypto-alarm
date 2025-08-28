package web

import (
	"crypto-gateway/internal/modules/ochestrator/repo"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/gofiber/fiber/v3"
)

func ValidateOrchestrator(c fiber.Ctx) error {
	var inputs []repo.OrchestratorInput
	if err := json.Unmarshal(c.Body(), &inputs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	for _, input := range inputs {
		if !slices.Contains([]string{"binance"}, input.SourceType) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid source_type"})
		}

		if _, err := validateSignal(input.Formula); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid formula: " + err.Error()})
		}
	}

	c.Locals("inputs", inputs)
	return c.Next()
}

func validateSignal(signal string) (string, error) {
	signal = strings.TrimSpace(signal)

	if signal == "" {
		return "", errors.New("empty formula")
	}

	_, err := govaluate.NewEvaluableExpression(signal)
	if err != nil {
		return "", err
	}

	return signal, nil
}
