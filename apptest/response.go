package apptest

import (
	"io"
	"net/http"
	"net/http/httptest"
)

type Response struct {
	RawResponse *httptest.ResponseRecorder
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
	body, _ := io.ReadAll(r.RawResponse.Result().Body)
	return string(body)
}

func (r *Response) Header() http.Header {
	return r.RawResponse.Result().Header
}
