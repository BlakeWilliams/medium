package session

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptedVerify(t *testing.T) {
	message := "Hello, world!"

	verifier := NewEncryptedVerifier("TheTruthIsOutThere1234567890!@#$")
	encodedMessage, err := verifier.Encode([]byte(message))
	require.NoError(t, err)

	decodedMessage, err := verifier.Decode(encodedMessage)
	require.NoError(t, err)
	require.Equal(t, message, string(decodedMessage))
}

func TestEncryptedVerifier_DifferentKeys(t *testing.T) {
	message := "Hello, world!"

	verifier := NewEncryptedVerifier("TheTruthIsOutThere1234567890!@#$")
	encodedMessage, err := verifier.Encode([]byte(message))
	require.NoError(t, err)

	otherVerifier := NewEncryptedVerifier("NewEncryptedVerifier1234567890!@")
	decodedMessage, err := otherVerifier.Decode(encodedMessage)
	require.Error(t, err)
	require.Nil(t, decodedMessage)
}

func TestEncryptedVerifier_InvalidMessage(t *testing.T) {
	verifier := NewEncryptedVerifier("TheTruthIsOutThere1234567890!@#$")
	decodedMessage, err := verifier.Decode("breakplz")

	require.Error(t, err)
	require.Nil(t, decodedMessage)
}

func TestEncryptedVerifier_InvalidNonceSize(t *testing.T) {
	b64data := base64.URLEncoding.EncodeToString([]byte("foo"))
	verifier := NewEncryptedVerifier("TheTruthIsOutThere1234567890!@#$")
	decodedMessage, err := verifier.Decode(b64data + "--" + b64data)

	require.ErrorContains(t, err, "nonce size is incorrect")
	require.Nil(t, decodedMessage)
}
