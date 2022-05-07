package session

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
)

// Generates and verifies messages to prevent tampering.
// Useful for session cookies and magic links.
type Verifier struct {
	secret string
}

// Returns a new Verifier that uses the provided secret to sign and verify
// messages.
func NewVerifier(secret string) Verifier {
	return Verifier{secret: secret}
}

// Accepts data as a byte array and returns a signed message.
func (v *Verifier) Encode(data []byte) string {
	encodedMessage := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encodedMessage, data)

	digest := hmac.New(sha1.New, []byte(v.secret)).Sum(encodedMessage)
	formattedDigest := make([]byte, base64.StdEncoding.EncodedLen(len(digest)))
	base64.StdEncoding.Encode(formattedDigest, digest)

	return fmt.Sprintf("%s--%s", encodedMessage, formattedDigest)
}

// Accepts a signed message and returns the original data if the message is
// valid.
func (v *Verifier) Decode(message []byte) ([]byte, error) {
	parts := bytes.SplitN(message, []byte("--"), 2)

	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid message, message format invalid")
	}
	data := parts[0]
	rawDigest := parts[1]

	decodedDigest := make([]byte, base64.StdEncoding.DecodedLen(len(rawDigest)))
	digestLength, err := base64.StdEncoding.Decode(decodedDigest, rawDigest)

	if err != nil {
		return nil, fmt.Errorf("Invalid message, decoding error: %s", err)
	}

	if !hmac.Equal(decodedDigest[:digestLength], hmac.New(sha1.New, []byte(v.secret)).Sum(data)) {
		return nil, fmt.Errorf("Invalid message, digest mismatch")
	}

	decodedMessage := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	length, err := base64.StdEncoding.Decode(decodedMessage, data)

	if err != nil {
		return nil, fmt.Errorf("Invalid message, decoding error: %s", err)
	}

	return decodedMessage[:length], nil
}
