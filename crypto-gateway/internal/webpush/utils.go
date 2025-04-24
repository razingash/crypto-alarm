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

	sigBytes := append(padTo32(r.Bytes()), padTo32(s.Bytes())...)
	encodedSig := base64.RawURLEncoding.EncodeToString(sigBytes)

	return encodedHeader + "." + encodedPayload + "." + encodedSig, nil
}

func DeriveSharedSecretECDH(
	serverPriv *ecdh.PrivateKey,
	clientPubBytes, authSecret []byte,
) (contentEncryptionKey, nonce, salt, clientPubOut, serverPubOut []byte, err error) {
	curve := ecdh.P256()

	clientPub, err := curve.NewPublicKey(clientPubBytes)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("invalid client pubkey: %w", err)
	}

	serverPub := serverPriv.PublicKey().Bytes()
	clientPubOut = clientPubBytes
	serverPubOut = serverPub

	sharedSecret, err := serverPriv.ECDH(clientPub)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to derive shared secret: %w", err)
	}

	salt = make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("salt gen failed: %w", err)
	}

	// HKDF info
	info := append([]byte("WebPush: info\x00"), append(clientPubBytes, serverPub...)...)

	prk := hkdf.Extract(sha256.New, authSecret, sharedSecret)
	hkdfReader := hkdf.New(sha256.New, prk, salt, info)

	contentEncryptionKey = make([]byte, 16) // AES-128-GCM
	nonce = make([]byte, 12)
	if _, err := io.ReadFull(hkdfReader, contentEncryptionKey); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("HKDF key gen failed: %w", err)
	}
	if _, err := io.ReadFull(hkdfReader, nonce); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("HKDF nonce gen failed: %w", err)
	}

	return contentEncryptionKey, nonce, salt, clientPubOut, serverPubOut, nil
}

func EncryptPushPayload(message []byte, key, nonce []byte) ([]byte, error) {
	paddingLen := 0
	record := make([]byte, 1+paddingLen+len(message))
	record[0] = byte(paddingLen)
	copy(record[1+paddingLen:], message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, record, nil)
	return ciphertext, nil
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

func padTo32(b []byte) []byte {
	if len(b) >= 32 {
		return b
	}
	pad := make([]byte, 32-len(b))
	return append(pad, b...)
}
