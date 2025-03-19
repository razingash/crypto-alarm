package handlers

import (
	"crypto-gateway/crypto-gateway/internal/auth"
	"crypto-gateway/crypto-gateway/internal/db"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Register(c fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.JSON(&body); err != nil {
		return err
	}

	user, err := auth.RegisterUser(body.Username, body.Password)
	if err != nil {
		return err
	}

	accessToken, err := auth.GenerateAccessToken(user.UUID)
	if err != nil {
		return err
	}

	refreshToken, err := auth.GenerateRefreshToken(user.UUID)
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(`INSERT INTO access_tokens (user_uuid, token, expires_at, created_at) 
							VALUES ($1, $2, $3, $4)`,
		user.UUID, accessToken, time.Now().Add(15*time.Minute), time.Now())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
