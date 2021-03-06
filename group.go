package medium

import (
	"context"
	"net/http"
)

// Group represents a collection of routes that share a common set of
// Around/Before/After callbacks and Action type (T)
type Group[P Action, T Action] struct {
	routes        []*Route[T]
	actionFactory func(context.Context, P, func(context.Context, T))
	subgroups     []dispatchable[T]
}

// NewGroup creates a new route Group that can be used to group around
// filters and other common behavior.
//
// A factory function is passed to the NewGroup, so that it can reference
// fields from the parent action type.
func NewGroup[P Action, T Action](factory func(context.Context, P, func(context.Context, T))) *Group[P, T] {
	return &Group[P, T]{routes: make([]*Route[T], 0), actionFactory: factory}
}

// Match is used to add a new Route to the Router
func (t *Group[P, T]) Match(method string, path string, handler HandlerFunc[T]) {
	t.routes = append(t.routes, newRoute(method, path, handler))
}

// Defines a new Route that responds to GET requests.
func (t *Group[P, T]) Get(path string, handler HandlerFunc[T]) {
	t.Match(http.MethodGet, path, handler)
}

// Defines a new Route that responds to POST requests.
func (t *Group[P, T]) Post(path string, handler HandlerFunc[T]) {
	t.Match(http.MethodPost, path, handler)
}

// Implements Dispatchable so groups can be registered on routers
func (g *Group[P, T]) dispatch(rw http.ResponseWriter, r *http.Request) (bool, map[string]string, func(context.Context, P)) {
	if route, params := g.routeFor(r); route != nil {
		return true, params, func(ctx context.Context, action P) {
			g.actionFactory(ctx, action, func(ctx context.Context, action T) {
				route.handler(ctx, action)
			})
		}
	}

	for _, group := range g.subgroups {
		if ok, params, handler := group.dispatch(rw, r); ok {
			return true, params, func(ctx context.Context, action P) {
				g.actionFactory(ctx, action, func(ctx context.Context, action T) {
					handler(ctx, action)
				})
			}
		}
	}

	return false, nil, nil
}

func (g *Group[P, T]) routeFor(r *http.Request) (*Route[T], map[string]string) {
	for _, route := range g.routes {
		if ok, params := route.IsMatch(r); ok {
			return route, params
		}
	}

	return nil, nil
}
