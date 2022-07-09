package apptest

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/blakewilliams/medium/pkg/router"
)

var RedirectCodes = []int{
	http.StatusMovedPermanently,
	http.StatusFound,
	http.StatusSeeOther,
	http.StatusTemporaryRedirect,
	http.StatusPermanentRedirect,
}

// Represents a request, or requests to a given medium application.
type AppRequest[T router.Action] struct {
	CookieJar    http.CookieJar
	LastResponse *httptest.ResponseRecorder
	router       *router.Router[T]
}

func New[T router.Action](router *router.Router[T]) *AppRequest[T] {
	req := &AppRequest[T]{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	req.CookieJar = jar
	req.router = router

	return req
}

func (ar *AppRequest[T]) Get(route string, headers http.Header) {
	ar.makeRequest(http.MethodGet, route, headers, nil)
}

func (ar *AppRequest[T]) PostForm(route string, headers http.Header, formValues url.Values) {
	if headers == nil {
		headers = make(http.Header)
	}

	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	data := strings.NewReader(formValues.Encode())
	ar.makeRequest(http.MethodPost, route, headers, data)
}

func (ar *AppRequest[T]) IsRedirect() bool {
	if ar.LastResponse == nil {
		return false
	}

	for _, redirectCode := range RedirectCodes {
		if ar.LastResponse.Code == redirectCode {
			return true
		}
	}

	return false
}

func (ar *AppRequest[T]) IsOK() bool {
	if ar.LastResponse == nil {
		return false
	}

	return ar.LastResponse.Code == 200
}

func (ar *AppRequest[T]) Code() int {
	return ar.LastResponse.Code
}

func (ar *AppRequest[T]) FollowRedirect() {
	ar.makeRequest(http.MethodGet, ar.LastResponse.Header().Get("Location"), nil, nil)
}

func (ar *AppRequest[T]) Body() string {
	return ar.LastResponse.Body.String()
}
func (ar *AppRequest[T]) makeRequest(method string, route string, headers http.Header, data io.Reader) {
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

	ar.LastResponse = httptest.NewRecorder()

	ar.router.ServeHTTP(ar.LastResponse, req)
	ar.CookieJar.SetCookies(req.URL, ar.LastResponse.Result().Cookies())
}
