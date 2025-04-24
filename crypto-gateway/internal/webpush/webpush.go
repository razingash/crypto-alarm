package webpush

import (
	"bytes"
	"crypto-gateway/config"
	"crypto/ecdh"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
)

func SendWebPush(endpoint string, p256dh string, auth string, messageJSON []byte) error {
	//ключи
	vapidECDSA, err := DecodeVAPIDPrivateKey(config.Vapid_Private_Key)
	if err != nil {
		return err
	}

	// ECDH
	ecdhPriv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// генерация JWT
	vapidJWT, err := GenerateVAPIDJWT(endpoint, "roumerchi@gmail.com", vapidECDSA)
	if err != nil {
		return err
	}

	// валидная кодировка
	clientPubRaw, err := base64.RawURLEncoding.DecodeString(p256dh)
	if err != nil {
		return fmt.Errorf("failed to decode p256dh: %w", err)
	}
	authRaw, err := base64.RawURLEncoding.DecodeString(auth)
	if err != nil {
		return fmt.Errorf("failed to decode auth: %w", err)
	}

	// ECDH + HKDF
	key, nonce, salt, _, serverPub, err := DeriveSharedSecretECDH(ecdhPriv, clientPubRaw, authRaw)
	if err != nil {
		return err
	}

	// шифрование
	ciphertext, err := EncryptPushPayload(messageJSON, key, nonce, serverPub)
	if err != nil {
		return err
	}

	body, err := BuildEncryptedBody(ciphertext, salt, serverPub)
	if err != nil {
		return err
	}

	req, err := BuildPushRequest(
		endpoint,
		body,
		salt,
		serverPub,
		elliptic.Marshal(elliptic.P256(), vapidECDSA.PublicKey.X, vapidECDSA.PublicKey.Y),
		vapidJWT)
	if err != nil {
		return err
	}

	resp, err := SendPush(req)
	if err != nil {
		return err
	}
	log.Println("Push sent, response:", resp)
	return nil
}

func SendPush(req *http.Request) (string, error) {
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("push request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	statusCode := resp.StatusCode
	status := resp.Status

	if statusCode != 201 {
		return "", fmt.Errorf("push failed: %s (code %d) - body: %s", status, statusCode, string(body))
	}

	log.Printf("Push sent successfully: %s\n", status)

	return string(body), nil
}

func BuildPushRequest(endpoint string, payload, salt, ecdhPub, vapidPub []byte, vapidJWT string) (*http.Request, error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("TTL", "86400") // 1 день, скорее хорошее значение чем плохое, нужно будет просто добавить историю
	req.Header.Set("Encryption", fmt.Sprintf("salt=%s", base64.RawURLEncoding.EncodeToString(salt)))
	req.Header.Set("Crypto-Key", fmt.Sprintf("dh=%s; p256ecdsa=%s",
		base64.RawURLEncoding.EncodeToString(ecdhPub),
		base64.RawURLEncoding.EncodeToString(vapidPub),
	))
	authHeader := fmt.Sprintf("vapid t=%s, k=%s", vapidJWT, base64.RawURLEncoding.EncodeToString(vapidPub))
	req.Header.Set("Authorization", authHeader)

	return req, nil
}
