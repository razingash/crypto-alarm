package webpush

import (
	"bytes"
	"crypto-gateway/config"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
)

func SendWebPush(endpoint string, p256dh string, auth string, messageJSON string) error {
	// разделить vapidPub и serverECDHPub

	//ключи
	priv, err := DecodeVAPIDPrivateKey(config.Vapid_Private_Key)
	if err != nil {
		return err
	}
	priv2, err := DecodeVAPIDPrivateKeyECDH(config.Vapid_Private_Key)
	if err != nil {
		return err
	}
	vapidPub := priv2.PublicKey().Bytes()

	// генерация JWT
	vapidJWT, err := GenerateVAPIDJWT(endpoint, "roumerchi@gmail.com", priv)
	if err != nil {
		return err
	}

	// ECDH + HKDF
	key, salt, err := DeriveSharedSecret(priv2, []byte(p256dh), []byte(auth))
	if err != nil {
		return err
	}

	// шифрование
	ciphertext, nonce, err := EncryptPushPayload([]byte(messageJSON), key, salt)
	if err != nil {
		return err
	}

	serverECDHPub := priv2.PublicKey().Bytes()
	req, err := BuildPushRequest(
		endpoint,
		ciphertext,
		nonce,
		salt,
		serverECDHPub,
		vapidPub,
		vapidJWT,
	)
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
		return "", fmt.Errorf("failed to send push request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func BuildPushRequest(endpoint string, payload []byte, nonce []byte, salt []byte,
	ecdhPub []byte, vapidPub []byte, vapidJWT string) (*http.Request, error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("TTL", "2419200") // 4 недели (мало?много? вроде вообще 30 должно быть)
	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(payload)))

	vapidHeader := fmt.Sprintf("vapid t=%s, k=%s",
		vapidJWT,
		base64.RawURLEncoding.EncodeToString(vapidPub),
	)
	req.Header.Set("Authorization", vapidHeader)

	req.Header.Set("Encryption", fmt.Sprintf("salt=%s", base64.RawURLEncoding.EncodeToString(salt)))
	req.Header.Set("Crypto-Key", fmt.Sprintf("dh=%s; p256ecdsa=%s",
		base64.RawURLEncoding.EncodeToString(ecdhPub),
		base64.RawURLEncoding.EncodeToString(vapidPub),
	))
	return req, nil
}
