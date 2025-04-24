package webpush

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
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
		"aud": origin,
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

// тут главная проблема с шифрованием
func DeriveSharedSecretECDH(
	serverPriv *ecdh.PrivateKey,
	clientPubBytes, authSecret []byte,
) (contentEncryptionKey, nonce, salt, clientPubOut, serverPubOut []byte, err error) {
	curve := ecdh.P256()

	clientPub, err := curve.NewPublicKey(clientPubBytes)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	serverPub := serverPriv.PublicKey().Bytes()
	clientPubOut = clientPubBytes
	serverPubOut = serverPub

	sharedSecret, err := serverPriv.ECDH(clientPub)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	salt = make([]byte, 16)
	if _, err = rand.Read(salt); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	hash := sha256.New

	// IKM: HKDF-Expand(prk, salt, "WebPush: info\x00", 32)
	info := append([]byte("WebPush: info\x00"), append(clientPubBytes, serverPub...)...)
	ikmHKDF := hkdf.New(hash, sharedSecret, authSecret, info)
	ikm := make([]byte, 32)
	if _, err = io.ReadFull(ikmHKDF, ikm); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	// CEK = HKDF-Expand(ikm, salt, "Content-Encoding: aes128gcm\x00", 16)
	cekHKDF := hkdf.New(hash, ikm, salt, []byte("Content-Encoding: aes128gcm\x00"))
	contentEncryptionKey = make([]byte, 16)
	if _, err = io.ReadFull(cekHKDF, contentEncryptionKey); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	// nonce: HKDF-Expand(ikm, salt, "Content-Encoding: nonce\x00", 12)
	nonceHKDF := hkdf.New(hash, ikm, salt, []byte("Content-Encoding: nonce\x00"))
	nonce = make([]byte, 12)
	if _, err = io.ReadFull(nonceHKDF, nonce); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return contentEncryptionKey, nonce, salt, clientPubOut, serverPubOut, nil
}

func EncryptPushPayload(message, key, nonce []byte, ecdhPub []byte) ([]byte, error) {
	// сборка заранее, чтоб не было херни
	headerLen := 16 /* salt */ + 4 /* recordSize */ + 1 + len(ecdhPub)

	dataBuf := bytes.NewBuffer(message)
	dataBuf.WriteByte(0x02)

	recordSize := 4096
	maxData := recordSize - 16 - headerLen
	if dataBuf.Len() > maxData {
		return nil, fmt.Errorf("this message too big")
	}
	padding := make([]byte, maxData-dataBuf.Len())
	dataBuf.Write(padding)

	// AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aesgcm.Seal(nil, nonce, dataBuf.Bytes(), nil), nil
}

func BuildEncryptedBody(ciphertext, salt, ecdhPub []byte) ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.Write(salt)

	// recordSize
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, 4096)
	buf.Write(tmp)

	buf.WriteByte(byte(len(ecdhPub)))
	buf.Write(ecdhPub)

	buf.Write(ciphertext)

	return buf.Bytes(), nil
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
