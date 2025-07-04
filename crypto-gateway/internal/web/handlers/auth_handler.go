package handlers

import (
	"context"
	"crypto-gateway/internal/web/auth"
	"crypto-gateway/internal/web/db"
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// возможно сделать структуру лучше - сейчас похожа на кал

func Register(c fiber.Ctx) error {
	body := c.Locals("body").(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	})

	user, err := auth.RegisterUser(body.Username, body.Password)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Пользователь с таким именем уже существует",
			})
		}

		// 500 для всех остальных
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка при регистрации пользователя",
		})
	}
	if err := c.JSON(&body); err != nil {
		return err
	}

	accessToken := auth.GenerateAccessToken(user.UUID)
	refreshToken := auth.GenerateRefreshToken(user.UUID)

	err = db.SaveAccessToken(user.UUID, accessToken)
	if err != nil {
		return nil
	}

	err = db.SaveRefreshToken(user.UUID, refreshToken)
	if err != nil {
		return nil
	}

	return c.JSON(fiber.Map{
		"access":  accessToken,
		"refresh": refreshToken,
	})
}

func Login(c fiber.Ctx) error {
	body := c.Locals("body").(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	})

	errCode, user := auth.LoginUser(body.Username, body.Password)
	if errCode != 0 {
		switch errCode {
		case auth.ErrCodeUserNotFound:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Incorrect login or password", // тут так и есть
			})
		case auth.ErrCodeInvalidPassword:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Incorrect login or password", // проблема с паролем
			})
		case auth.ErrCodeDBError:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Incorrect login or password", // нет пользователя с таким логином
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Неизвестная ошибка", // какая-то херня которой не должно быть
			})
		}
	}

	accessToken := auth.GenerateAccessToken(user.UUID)   // 10 минут
	refreshToken := auth.GenerateRefreshToken(user.UUID) // 1 день

	var err error
	err = db.SaveAccessToken(user.UUID, accessToken)
	if err != nil {
		return nil
	}

	err = db.SaveRefreshToken(user.UUID, refreshToken)
	if err != nil {
		return nil
	}

	return c.JSON(fiber.Map{
		"access":  accessToken,
		"refresh": refreshToken,
	})
}

// позже добавить учет черного списка
func ValidateToken(c fiber.Ctx) error {
	// both refresh and access
	var body struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	isTokenValid := auth.ValidateToken(body.Token)

	return c.JSON(fiber.Map{
		"isValid": isTokenValid,
	})
}

func RefreshAccessToken(c fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	errCode, newAccessToken := auth.GetNewAccessToken(body.Token)

	if errCode == 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	return c.JSON(fiber.Map{
		"access": newAccessToken,
	})
}

func Logout(c fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}
	userUUID := c.Locals("userUUID").(string)

	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	isTokenValid := auth.ValidateRefreshToken(body.Token)
	if !isTokenValid {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	var err error
	_, err = db.DB.Exec(context.Background(), `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE user_uuid = $1
	`, userUUID)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}
