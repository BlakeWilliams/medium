package medium

import (
	"bytes"
	"fmt"
	"net/http"
)

// A response writer that defers writing the response/header information until
// Flush is called.
type DeferredResponseWriter struct {
	StatusCode             int
	content                bytes.Buffer
	originalResponseWriter http.ResponseWriter
}

func NewDeferredResponseWriter(rw http.ResponseWriter) *DeferredResponseWriter {
	return &DeferredResponseWriter{
		content:                *new(bytes.Buffer),
		StatusCode:             200,
		originalResponseWriter: rw,
	}
}

func (drw *DeferredResponseWriter) Header() http.Header {
	return drw.originalResponseWriter.Header()
}

func (drw *DeferredResponseWriter) Write(w []byte) (int, error) {
	return drw.content.Write(w)
}

func (drw *DeferredResponseWriter) WriteHeader(statusCode int) {
	drw.StatusCode = statusCode
}

func (drw *DeferredResponseWriter) Flush() error {
	drw.originalResponseWriter.WriteHeader(drw.StatusCode)
	_, err := drw.originalResponseWriter.Write(drw.content.Bytes())

	if err != nil {
		return fmt.Errorf("Error writing deferred response: %s", err)
	}

	return nil
}

var _ http.ResponseWriter = &DeferredResponseWriter{}
