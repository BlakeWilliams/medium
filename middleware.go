package medium

import (
	"context"
	"net/http"
)

// MiddlewareFunc is a function that is called before the action is executed.
// See Router.Use for more information.
// type MiddlewareFunc func(c Action, next HandlerFunc[Action])
type MiddlewareFunc func(context.Context, *MiddlewareContext)

type MiddlewareContext struct {
	Request           *http.Request
	currentMiddleware int
	ResponseWriter    http.ResponseWriter
	middlewares       []MiddlewareFunc
}

func (m *MiddlewareContext) Run() {
	m.middlewares[m.currentMiddleware](m.Request.Context(), m)
}

func (m *MiddlewareContext) Next(ctx context.Context) {
	if len(m.middlewares) > m.currentMiddleware {
		current := m.currentMiddleware
		m.currentMiddleware++
		m.middlewares[current](ctx, m)
	}
}
