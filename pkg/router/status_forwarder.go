package router

import (
	"net/http"
)

type statusForwarder struct {
	originalResponseWriter http.ResponseWriter
	onWriteHeader          (func(int))
}

func (sf *statusForwarder) Header() http.Header {
	return sf.originalResponseWriter.Header()
}

func (sf *statusForwarder) Write(w []byte) (int, error) {
	return sf.originalResponseWriter.Write(w)
}

func (sf *statusForwarder) WriteHeader(statusCode int) {
	sf.onWriteHeader(statusCode)
	sf.originalResponseWriter.WriteHeader(statusCode)
}

var _ http.ResponseWriter = (*statusForwarder)(nil)
