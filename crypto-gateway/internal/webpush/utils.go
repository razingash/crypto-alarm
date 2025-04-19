package webpush

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"golang.org/x/crypto/hkdf"
)

// из-за того что гугл использует нормальную архитектуру нужно использовать более универсальный JWT - ES256
func GenerateVAPIDJWT(endpoint, subject string, privateKey *ecdsa.PrivateKey) (string, error) {
	header := map[string]string{
		"alg": "ES256",
		"typ": "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)

	origin := extractOrigin(endpoint)
	exp := time.Now().Unix() + 12*60*60 // 12 часов
	payload := map[string]interface{}{
		"aud": origin, // позде сменить на uuid?
		"exp": exp,
		"sub": subject,
	}
	payloadJSON, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// ES256
	hash := sha256.Sum256([]byte(encodedHeader + "." + encodedPayload))
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", err
	}

	sigBytes := append(r.Bytes(), s.Bytes()...)
	encodedSig := base64.RawURLEncoding.EncodeToString(sigBytes)

	return encodedHeader + "." + encodedPayload + "." + encodedSig, nil
}

func DeriveSharedSecret(serverPriv *ecdh.PrivateKey, clientPubBytes, authSecret []byte) ([]byte, []byte, error) {
	curve := ecdh.P256()
	clientPub, err := curve.NewPublicKey(clientPubBytes)
	if err != nil {
		return nil, nil, err
	}

	sharedSecret, err := serverPriv.ECDH(clientPub)
	if err != nil {
		return nil, nil, err
	}

	salt := make([]byte, 16) // возможно из-за этого могут быть баги
	rand.Read(salt)

	h := hkdf.New(sha256.New, sharedSecret, salt, authSecret)
	key := make([]byte, 16) // AES-128-GCM
	_, err = io.ReadFull(h, key)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

func EncryptPushPayload(payload, key, salt []byte) ([]byte, []byte, error) {
	nonce := make([]byte, 12) // также могут быть баги из-за слабой вариативности
	rand.Read(nonce)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, payload, nil)
	return ciphertext, nonce, nil
}

func DecodeVAPIDPrivateKey(b64PrivKey string) (*ecdsa.PrivateKey, error) {
	rawKey, err := base64.RawURLEncoding.DecodeString(b64PrivKey)
	if err != nil {
		return nil, err
	}
	if len(rawKey) != 32 {
		return nil, errors.New("invalid VAPID private key length")
	}

	d := new(big.Int).SetBytes(rawKey)

	curve := elliptic.P256()
	x, y := curve.ScalarBaseMult(rawKey) // зараза, исправить нельзя, потому что перестает работать если нормально сделать

	return &ecdsa.PrivateKey{
		D: d,
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
	}, nil
}

func DecodeVAPIDPrivateKeyECDH(b64PrivKey string) (*ecdh.PrivateKey, error) {
	rawKey, err := base64.RawURLEncoding.DecodeString(b64PrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 key: %w", err)
	}

	if len(rawKey) != 32 {
		return nil, errors.New("invalid ECDH private key length")
	}

	privKey, err := ecdh.P256().NewPrivateKey(rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDH private key: %w", err)
	}

	return privKey, nil
}

// получает 'https://fcm.googleapis.com' из endpoint
func extractOrigin(endpoint string) string {
	parts := strings.Split(endpoint, "/")
	return parts[0] + "//" + parts[2]
}
