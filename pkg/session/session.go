package session

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

// Store is a wrapper around a http.Cookie that provides signed messages,
// allowing you to securely store data in a cookie.
//
// The data stored is still readable by the client, so secrets and sensitive
// data should not be stored in Store.Data.
type Store[T any] struct {
	verifier     Verifier
	name         string
	Data         T
	originalData T
}

// New creates a new Store with the given name and verifies Data using the
// passed in Verifier.
func New[T any](name string, verifier Verifier) *Store[T] {
	return &Store[T]{
		name:     name,
		verifier: verifier,
	}
}

// FromRequest reads the cookie with the provided name from the Request. The
// data is then decoded and verified using the Verifier.
func (s *Store[T]) FromRequest(r *http.Request) error {
	cookie, err := r.Cookie(s.name)

	if err != nil {
		return fmt.Errorf("Could not create session from request: %w", err)
	}

	return s.FromCookie(cookie)
}

// FromCookie attempts to decode the data from the passed in Cookie and verifies
// the data hasn't been tampered with.
func (s *Store[T]) FromCookie(cookie *http.Cookie) error {
	decodedMessage, err := s.verifier.Decode([]byte(cookie.Value))

	if err != nil {
		return err
	}

	err = json.Unmarshal(decodedMessage, &s.Data)

	if err != nil {
		return fmt.Errorf("Could not decode session: %w", err)
	}
	err = json.Unmarshal(decodedMessage, &s.originalData)
	if err != nil {
		return err
	}

	return nil
}

// Write encodes the Data into a JSON object, signs it using the message
// verifier, then writes it to the passed in http.ResponseWriter using the name
// provided by New.
func (s *Store[T]) Write(w http.ResponseWriter) error {
	jsonValue, err := json.Marshal(s.Data)

	if err != nil {
		return fmt.Errorf("Could not marshal session data: %w", err)
	}

	encodedData := s.verifier.Encode(jsonValue)

	http.SetCookie(w, &http.Cookie{
		Name:  s.name,
		Value: string(encodedData),
	})

	return nil
}

// Writes the session to the response writer only if the underlying data has
// changed.
func (s *Store[T]) WriteIfChanged(w http.ResponseWriter) error {
	if reflect.DeepEqual(s.originalData, s.Data) {
		return nil
	}

	return s.Write(w)
}
