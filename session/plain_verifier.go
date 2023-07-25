package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

// Generates and verifies messages to prevent tampering.
// Useful for session cookies and magic links.
type PlainVerifier struct {
	secret string
}

var _ Verifier = (*PlainVerifier)(nil)

// Returns a new Verifier that uses the provided secret to sign and verify
// messages.
func NewVerifier(secret string) PlainVerifier {
	return PlainVerifier{secret: secret}
}

// Accepts data as a byte array and returns a signed message.
func (v PlainVerifier) Encode(data []byte) (string, error) {
	encodedMessage := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encodedMessage, data)

	digest := hmac.New(sha256.New, []byte(v.secret)).Sum(encodedMessage)
	formattedDigest := make([]byte, base64.StdEncoding.EncodedLen(len(digest)))
	base64.StdEncoding.Encode(formattedDigest, digest)

	return fmt.Sprintf("%s--%s", encodedMessage, formattedDigest), nil
}

// Accepts a signed message and returns the original data if the message is
// valid.
func (v PlainVerifier) Decode(message string) ([]byte, error) {
	parts := strings.SplitN(message, "--", 2)

	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid message, message format invalid")
	}
	data := parts[0]
	rawDigest := parts[1]

	decodedDigest, err := base64.StdEncoding.DecodeString(rawDigest)

	if err != nil {
		return nil, fmt.Errorf("Invalid message, decoding error: %s", err)
	}

	expected := hmac.New(sha256.New, []byte(v.secret)).Sum([]byte(data))
	if !hmac.Equal(decodedDigest, expected) {
		return nil, fmt.Errorf("Invalid message, digest mismatch")
	}

	// decodedMessage := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	decodedMessage, err := base64.StdEncoding.DecodeString(data)

	if err != nil {
		return nil, fmt.Errorf("Invalid message, decoding error: %s", err)
	}

	return decodedMessage, nil
}
