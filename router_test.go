package medium

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	router := New(WithNoData)

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	router.Before(func(ctx context.Context, req *Request[NoData], next Next) Response {
		res := next(ctx)
		res.Header().Add("x-from-before", "amazing")

		return res
	})

	router.Get("/hello/:name", func(ctx context.Context, r *Request[NoData]) Response {
		require.Equal(t, "/hello/:name", r.MatchedPath())
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
	require.Equal(t, "amazing", rw.Header().Get("x-from-before"))
}

func TestGroup_RouteMethods(t *testing.T) {
	testCases := map[string]struct {
		method string
	}{
		"Get":    {method: http.MethodGet},
		"Post":   {method: http.MethodPost},
		"Put":    {method: http.MethodPut},
		"Patch":  {method: http.MethodPatch},
		"Delete": {method: http.MethodDelete},
	}
	router := New(WithNoData)

	for name, tc := range testCases {
		path := reflect.ValueOf("/")
		var handler HandlerFunc[NoData] = func(ctx context.Context, r *Request[NoData]) Response {
			res := NewResponse()

			res.WriteStatus(http.StatusOK)
			res.WriteString("hello")

			return res
		}

		handlerValue := reflect.ValueOf(handler)

		t.Run(name, func(t *testing.T) {
			reflect.ValueOf(router).MethodByName(name).Call([]reflect.Value{path, handlerValue})

			req := httptest.NewRequest(tc.method, "/", nil)
			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, req)

			require.Equal(t, 200, rw.Code)
		})
	}
}

func TestRouter_Post(t *testing.T) {
	router := New(WithNoData)

	router.Post("/hello/:name", func(ctx context.Context, r *Request[NoData]) Response {
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodPost, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
}

func TestRouter_MissingRoute_NoHandler(t *testing.T) {
	router := New(WithNoData)

	req := httptest.NewRequest(http.MethodGet, "/where/do/i/go", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "404 not found", rw.Body.String())
	require.Equal(t, 404, rw.Result().StatusCode)
}

func TestRouter_MissingRoute_WithHandler(t *testing.T) {
	router := New(WithNoData)

	router.Missing(func(ctx context.Context, r *Request[NoData]) Response {
		return StringResponse(http.StatusNotFound, "Sorry, can't find that page.")
	})

	req := httptest.NewRequest(http.MethodGet, "/where/do/i/go", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "Sorry, can't find that page.", rw.Body.String())
	require.Equal(t, 404, rw.Result().StatusCode)
}

func TestCustomActionType(t *testing.T) {
	router := New[*MyData](func(rootRequest RootRequest) *MyData {
		return &MyData{Value: 1}
	})

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	router.Get("/hello/:name", func(ctx context.Context, r *Request[*MyData]) Response {
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s, data %d", r.Params()["name"], r.Data.Value))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder, data 1", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

type myResponseWriter struct {
	orw     http.ResponseWriter
	onWrite func()
}

var _ http.ResponseWriter = (*myResponseWriter)(nil)

func (mrw *myResponseWriter) Header() http.Header { return mrw.orw.Header() }
func (mrw *myResponseWriter) WriteHeader(s int)   { mrw.orw.WriteHeader(s) }
func (mrw *myResponseWriter) Write(b []byte) (int, error) {
	mrw.onWrite()
	return mrw.orw.Write(b)
}

func TestCustomResponseWriter(t *testing.T) {
	mrwCalled := false

	router := New(WithNoData)
	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		mrw := &myResponseWriter{orw: rw, onWrite: func() { mrwCalled = true }}
		next(mrw, r)
	})

	called := false
	router.Get("/", func(ctx context.Context, r *Request[NoData]) Response {
		called = true

		return OK()
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	require.True(t, called)
	require.True(t, mrwCalled)
}

func TestCustomResponse(t *testing.T) {
	router := New(WithNoData)
	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		ctx := context.WithValue(r.Context(), "foo", "bar")
		next(rw, r.WithContext(ctx))
	})

	called := false
	router.Get("/", func(ctx context.Context, r *Request[NoData]) Response {
		require.Equal(t, "bar", r.Request().Context().Value("foo"))
		called = true

		return OK()
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	require.True(t, called)
}

func TestWritesHeaders(t *testing.T) {
	router := New(WithNoData)
	router.Get("/", func(ctx context.Context, r *Request[NoData]) Response {
		res := NewResponse()

		res.WriteStatus(http.StatusOK)
		res.Header().Add("x-foo", "bar")
		res.WriteString("hello")

		return res
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "bar", rw.Header().Get("x-foo"))
}

func TestBefore_EarlyExit(t *testing.T) {
	router := New(WithNoData)

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	firstBeforeCalled := false
	router.Before(func(ctx context.Context, req *Request[NoData], next Next) Response {
		firstBeforeCalled = true
		return StringResponse(http.StatusNotFound, "not found")
	})

	secondBeforeCalled := false
	router.Before(func(ctx context.Context, req *Request[NoData], next Next) Response {
		secondBeforeCalled = true
		return next(ctx)
	})

	routeCalled := false
	router.Get("/hello/:name", func(ctx context.Context, r *Request[NoData]) Response {
		routeCalled = true
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, http.StatusNotFound, rw.Code)
	require.True(t, firstBeforeCalled)
	require.False(t, secondBeforeCalled)
	require.False(t, routeCalled)
}

func TestBefore_DifferentDataType(t *testing.T) {
	router := New(func(rootRequest RootRequest) *MyData {
		return &MyData{Value: 1}
	})

	router.Before(func(ctx context.Context, req *Request[*MyData], next Next) Response {
		return GenericBefore(ctx, req, next)
	})

	router.Get("/hello/:name", func(ctx context.Context, r *Request[*MyData]) Response {
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)
}

func TestBefore_Context(t *testing.T) {
	router := New(WithNoData)

	router.Before(func(ctx context.Context, req *Request[NoData], next Next) Response {
		ctx = context.WithValue(ctx, "first", "bar")
		return GenericBefore(ctx, req, next)
	})

	router.Before(func(ctx context.Context, req *Request[NoData], next Next) Response {
		ctx = context.WithValue(ctx, "second", "baz")
		return next(ctx)
	})

	called := false
	var handlerCtx context.Context
	router.Get("/hello/:name", func(ctx context.Context, r *Request[NoData]) Response {
		handlerCtx = ctx
		called = true
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.True(t, called)
	require.NotNil(t, handlerCtx)
	require.Equal(t, "bar", handlerCtx.Value("first"))
	require.Equal(t, "baz", handlerCtx.Value("second"))
}

func GenericBefore[T any](ctx context.Context, req *Request[T], next Next) Response {
	return next(ctx)
}

func Test_BeforeModifiesData(t *testing.T) {
	router := New(func(rootRequest RootRequest) *MyData {
		return &MyData{Value: 1}
	})

	router.Before(func(ctx context.Context, req *Request[*MyData], next Next) Response {
		require.Equal(t, 1, req.Data.Value)
		req.Data.Value = 2
		return next(ctx)
	})

	router.Before(func(ctx context.Context, req *Request[*MyData], next Next) Response {
		require.Equal(t, 2, req.Data.Value)
		req.Data.Value = 3
		return next(ctx)
	})

	router.Get("/hello/:name", func(ctx context.Context, r *Request[*MyData]) Response {
		return OK()
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)
	require.Equal(t, http.StatusOK, rw.Code)
}
