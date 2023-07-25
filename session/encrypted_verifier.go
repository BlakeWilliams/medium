package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// EncryptedVerifier is a Verifier that uses AES-GCM to encrypt and decrypt
// messages.
type EncryptedVerifier struct {
	secret string
}

var _ Verifier = (*EncryptedVerifier)(nil)

// NewEncryptedVerifier creates a new EncryptedVerifier with the given secret.
// The provided secret must be 16, 24, or 32 bytes long to use AES-128, AES-192,
// or AES-256 respectively.
func NewEncryptedVerifier(secret string) EncryptedVerifier {
	return EncryptedVerifier{secret: secret}
}

// Encode encrypts the given data using AES-GCM and returns the encoded message.
func (v EncryptedVerifier) Encode(data []byte) (string, error) {
	block, err := aes.NewCipher([]byte(v.secret))

	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nil, nonce, []byte(data), nil)

	encodedNonce := base64.URLEncoding.EncodeToString(nonce)
	encodedCipherText := base64.URLEncoding.EncodeToString(cipherText)

	return fmt.Sprintf("%s--%s", encodedNonce, encodedCipherText), nil
}

// Decode decrypts the given message using AES-GCM and returns the decoded data.
func (v EncryptedVerifier) Decode(data string) ([]byte, error) {
	parts := strings.SplitN(data, "--", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid message, message format invalid")
	}

	nonce, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	cipherText, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher([]byte(v.secret))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("nonce size is incorrect")
	}

	plaintext, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
