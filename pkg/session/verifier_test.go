package session

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerify(t *testing.T) {
	message := "Hello, world!"

	verifier := NewVerifier("TheTruthIsOutThere")
	encodedMessage := verifier.Encode([]byte(message))

	decodedMessage, err := verifier.Decode([]byte(encodedMessage))
	require.NoError(t, err)
	require.Equal(t, message, string(decodedMessage))
}
func TestVerify_DifferentKeys(t *testing.T) {
	message := "Hello, world!"

	verifier := NewVerifier("TheTruthIsOutThere")
	encodedMessage := verifier.Encode([]byte(message))

	otherVerifier := NewVerifier("NewVerifier")
	decodedMessage, err := otherVerifier.Decode([]byte(encodedMessage))
	require.Error(t, err)
	require.Nil(t, decodedMessage)
}

func TestVerify_TamperedHash(t *testing.T) {
	message := "Hello, world!"

	verifier := NewVerifier("TheTruthIsOutThere")
	encodedMessage := verifier.Encode([]byte(message))
	parts := strings.SplitN(encodedMessage, "--", 2)
	hash := parts[1]

	tamperedData := base64.StdEncoding.EncodeToString([]byte("TheTruthIsNOTOutThere"))
	tamperedMessage := fmt.Sprintf("%s--%s", tamperedData, hash)

	decodedMessage, err := verifier.Decode([]byte(tamperedMessage))
	require.Error(t, err)
	require.Nil(t, decodedMessage)
}

func TestVerify_InvalidMessage(t *testing.T) {
	verifier := NewVerifier("TheTruthIsOutThere")
	decodedMessage, err := verifier.Decode([]byte("breakplz"))

	require.Error(t, err)
	require.Nil(t, decodedMessage)
}

func TestVerify_InvalidBase64Message(t *testing.T) {
	verifier := NewVerifier("TheTruthIsOutThere")

	tamperedData := base64.StdEncoding.EncodeToString([]byte("TheTruthIsNOTOutThere"))
	tamperedMessage := fmt.Sprintf("%s--%s", tamperedData, "omgahash")

	decodedMessage, err := verifier.Decode([]byte(tamperedMessage))

	require.Error(t, err)
	require.Nil(t, decodedMessage)
}
