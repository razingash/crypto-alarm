package auth

import (
	"context"
	"crypto-gateway/config"
	"crypto-gateway/internal/db"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

/*
сейчас используется симметричное шифрование HS256, подходит пока валидация токенов на одном сервере
если нужно чтобы проверка была на нескольких то нужно использовать ES256 - тогда будет еще и публичный токен,
которым можно будет валидировать их
*/

type Payload struct {
	Sub string `json:"sub"` // uuid
	Exp int64  `json:"exp"` // expiration time
	Iat int64  `json:"iat"` // issued at
}

func GenerateAccessToken(userUUID string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	expiry := int64(600) // 10 минут

	payloadStruct := Payload{
		Sub: userUUID,
		Exp: time.Now().Unix() + expiry,
		Iat: time.Now().Unix(),
	}
	payloadBytes, _ := json.Marshal(payloadStruct)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signature := signHMAC(header+"."+payload, config.SecretKey)

	token := header + "." + payload + "." + signature
	return token
}

func GenerateRefreshToken(userUUID string) string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)

	h := hmac.New(sha256.New, []byte(config.SecretKey))
	h.Write([]byte(userUUID))
	h.Write(randomBytes)

	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// получает uuid из полезной нагрузки токена доступа, должен быть передан свалидированный токен
func ExtractUUID(token string) (string, int) {
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		if ValidateAccessToken(parts) {
			payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				return "", 1 // errors.New("failed to decode payload")
			}

			var payloadData Payload
			err = json.Unmarshal(payloadBytes, &payloadData)
			if err != nil {
				return "", 2 // errors.New("failed to unmarshal payload")
			}

			// Возвращаем UUID пользователя из полезной нагрузки
			return payloadData.Sub, 0
		}
	}
	return "", 3 // неизвестная ошибка
}

func ValidateToken(token string) bool {
	// пока возвращать просто булевые значения, если понадобится добавить и текст
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		return ValidateAccessToken(parts)
	} else {
		return ValidateRefreshToken(token)
	}
}

func ValidateRefreshToken(token string) bool {
	var expTime time.Time
	var revoked bool

	err := db.DB.QueryRow(context.Background(), `
		SELECT expires_at, revoked FROM refresh_tokens WHERE token = $1
	`, token).Scan(&expTime, &revoked)

	if err != nil || revoked || time.Now().After(expTime) {
		return false
	}

	return true
}

func ValidateAccessToken(token_parts []string) bool {
	header, payload, receivedSig := token_parts[0], token_parts[1], token_parts[2]
	expectedSig := signHMAC(header+"."+payload, config.SecretKey)

	if receivedSig != expectedSig {
		return false //, "Invalid signature"
	}

	return CheckAccessTokenRelevance(payload)
}

func CheckAccessTokenRelevance(payload string) bool {
	// не смог сделать сразу для двух типов
	payloadBytes, _ := base64.RawURLEncoding.DecodeString(payload)
	var payloadData Payload
	json.Unmarshal(payloadBytes, &payloadData)

	if time.Now().Unix() > payloadData.Exp {
		return false // "Token expired"
	}

	return true // "Valid token"
}

// возвращает новый токен досутпа если токен перезарядки в порядке, в противном случае ошибка
func GetNewAccessToken(token string) (int, string) {
	var userUUID string
	var expTime time.Time
	var revoked bool

	err := db.DB.QueryRow(context.Background(), `
		SELECT user_uuid, expires_at, revoked 
		FROM refresh_tokens 
		WHERE token = $1
	`, token).Scan(&userUUID, &expTime, &revoked)

	if err != nil || revoked || time.Now().After(expTime) {
		return 1, ""
	}

	return 0, GenerateAccessToken(userUUID)
}

// генерация подписи
func signHMAC(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(signature)
}
