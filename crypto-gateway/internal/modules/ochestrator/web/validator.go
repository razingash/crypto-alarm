package web

import (
	"crypto-gateway/internal/modules/ochestrator/repo"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/gofiber/fiber/v3"
)

func ValidateOrchestrator(c fiber.Ctx) error {
	var inputs []repo.OrchestratorInput
	if err := json.Unmarshal(c.Body(), &inputs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(inputs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "inputs cannot be empty",
		})
	}

	for i, input := range inputs {
		if input.Tag == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("input[%d]: tag cannot be empty", i),
			})
		}

		if _, err := validateSignal(input.Formula); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("input[%d]: invalid formula: %s", i, err.Error()),
			})
		}

		if len(input.Sources) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("input[%d]: sources cannot be empty", i),
			})
		}

		for j, src := range input.Sources {
			if !slices.Contains([]string{"binance"}, src.SourceType) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("input[%d].sources[%d]: invalid source_type '%s'", i, j, src.SourceType),
				})
			}
			if src.SourceID <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("input[%d].sources[%d]: invalid source_id '%d'", i, j, src.SourceID),
				})
			}
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

	re := regexp.MustCompile(`formula\d+`)
	normalized := re.ReplaceAllString(signal, "true")

	normalized = strings.ReplaceAll(normalized, "AND", "&&")
	normalized = strings.ReplaceAll(normalized, "OR", "||")
	normalized = strings.ReplaceAll(normalized, "NOT", "!")

	expr, err := govaluate.NewEvaluableExpression(normalized)
	if err != nil {
		return "", fmt.Errorf("invalid formula: %w", err)
	}

	_, err = expr.Evaluate(nil)
	if err != nil {
		return "", fmt.Errorf("formula evaluation failed: %w", err)
	}

	return signal, nil
}
