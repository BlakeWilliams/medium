package medium

import (
	"bytes"
	"io"
	"net/http"
)

type ResponseWriter interface {
	Status() int
	Header() http.Header
	WriteStatus(status int)
	Write(data []byte) (int, error)
}

type responseWriter struct {
	status int
	header http.Header
	w      http.ResponseWriter
	body   bytes.Buffer
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) WriteStatus(status int) {
	rw.status = status
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	return rw.body.Write(data)
}

func (rw *responseWriter) Flush() (int, error) {
	written, err := io.Copy(rw.w, &rw.body)
	return int(written), err
}

func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	return &responseWriter{
		w: w,
	}
}

func (rw *responseWriter) response() Response {
	res := NewResponse()

	res.WriteStatus(rw.status)
	for key, values := range rw.header {
		for _, value := range values {
			res.header.Add(key, value)
		}
	}

	io.Copy(res, &rw.body)

	return res
}
