package apptest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

var RedirectCodes = []int{
	http.StatusMovedPermanently,
	http.StatusFound,
	http.StatusSeeOther,
	http.StatusTemporaryRedirect,
	http.StatusPermanentRedirect,
}

type Response struct {
	RawResponse *httptest.ResponseRecorder
	bodyCache   []byte
	mu          sync.Mutex
}

func (r *Response) Code() int {
	return r.RawResponse.Code
}

func (r *Response) IsRedirect() bool {
	for _, redirectCode := range RedirectCodes {
		if r.Code() == redirectCode {
			return true
		}
	}

	return false
}

func (r *Response) IsOK() bool {
	return r.Code() == http.StatusOK
}

func (r *Response) Body() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.bodyCache == nil {
		r.bodyCache, _ = io.ReadAll(r.RawResponse.Result().Body)
	}

	return string(r.bodyCache)
}

func (r *Response) Header() http.Header {
	return r.RawResponse.Result().Header
}
