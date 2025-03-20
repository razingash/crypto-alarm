package auth

import (
	"crypto-gateway/crypto-gateway/config"
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

func GenerateAccessToken(userID string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	expiry := int64(600) // 10 минут

	payloadStruct := Payload{
		Sub: userID,
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

func ValidateAccessToken(token string) (bool, string) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false, "Invalid token format"
	}

	header, payload, receivedSig := parts[0], parts[1], parts[2]
	expectedSig := signHMAC(header+"."+payload, config.SecretKey)

	if receivedSig != expectedSig {
		return false, "Invalid signature"
	}

	payloadBytes, _ := base64.RawURLEncoding.DecodeString(payload)
	var payloadData Payload
	json.Unmarshal(payloadBytes, &payloadData)

	if time.Now().Unix() > payloadData.Exp {
		return false, "Token expired"
	}

	return true, "Valid token"
}

// генерация подписи
func signHMAC(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(signature)
}
