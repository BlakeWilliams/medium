package apptest

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
)

// Represents a request, or requests to a given medium application.
type Session struct {
	CookieJar http.CookieJar
	medium    http.Handler
}

func New(medium http.Handler) *Session {
	req := &Session{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	req.CookieJar = jar
	req.medium = medium

	return req
}

func (ar *Session) Get(route string, headers http.Header) *Response {
	return ar.makeRequest(http.MethodGet, route, headers, nil)
}

func (ar *Session) PostForm(route string, headers http.Header, formValues url.Values) *Response {
	if headers == nil {
		headers = make(http.Header)
	}

	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	data := strings.NewReader(formValues.Encode())
	return ar.makeRequest(http.MethodPost, route, headers, data)
}

func (ar *Session) FollowRedirect(res *Response) *Response {
	return ar.makeRequest(http.MethodGet, res.Header().Get("Location"), nil, nil)
}

func (ar *Session) makeRequest(method string, route string, headers http.Header, data io.Reader) *Response {
	req := httptest.NewRequest(method, route, data)
	req.URL.Scheme = "http"
	req.URL.Host = "app.test"

	// Set headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set cookies
	for _, cookie := range ar.CookieJar.Cookies(req.URL) {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()

	ar.medium.ServeHTTP(recorder, req)
	ar.CookieJar.SetCookies(req.URL, recorder.Result().Cookies())

	return &Response{RawResponse: recorder}
}
