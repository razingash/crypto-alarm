package handlers

import (
	"context"
	"crypto-gateway/crypto-gateway/internal/auth"
	"crypto-gateway/crypto-gateway/internal/db"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// возможно сделать структуру лучше - сейчас похожа на кал

func Register(c fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	body.Username = c.FormValue("username")
	body.Password = c.FormValue("password")

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

	log.Println(user)
	accessToken := auth.GenerateAccessToken(user.UUID)
	refreshToken := auth.GenerateRefreshToken(user.UUID)

	_, err = db.DB.Exec(context.Background(), `
		INSERT INTO access_tokens (user_uuid, token, expires_at, created_at) 
		VALUES ($1, $2, $3, $4)`,
		user.UUID, accessToken, time.Now().Add(15*time.Minute), time.Now())
	if err != nil {
		return nil
	}

	_, err = db.DB.Exec(context.Background(), `
		INSERT INTO refresh_tokens (user_uuid, token, expires_at, created_at) 
		VALUES ($1, $2, $3, $4)`,
		user.UUID, refreshToken, time.Now().Add(24*time.Hour), time.Now())
	if err != nil {
		return nil
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
