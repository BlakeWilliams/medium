package medium

import (
	"context"
	"net/http"
	"regexp"
)

// registerable represents a type that can be registered on a router or a group
// to create subgroups/subrouters.
type registerable[T Action] interface {
	register(dispatchable[T])
	prefix() string
}

// Group represents a collection of routes that share a common set of
// Around/Before/After callbacks and Action type (T)
type Group[P Action, T Action] struct {
	routes        []*Route[T]
	actionFactory func(context.Context, *GroupContext[P, T])
	subgroups     []dispatchable[T]
	routePrefix   string
}

// GroupContext represents the current action and exposts a Next function that
// allows you to create routes for a new Action type composed from the old
// action.
type GroupContext[PrevActionT Action, NextActionT Action] struct {
	action PrevActionT
	next   func(context.Context, NextActionT)
}

func (rc *GroupContext[PrevActionT, NextActionT]) Action() PrevActionT {
	return rc.action
}

func (rc *GroupContext[PrevActionT, NextActionT]) Next(ctx context.Context, nextAction NextActionT) {
	rc.next(ctx, nextAction)
}

// Subrouter creates a new grouping of routes that will be routed to in addition
// to the routes defined on the primery router. These routes will be prefixed
// using the given prefix.
//
// Subrouters provide their own action factory, so common behavior can be
// grouped via a custom action.
//
// Diagram of what the "inheritance" chain can look like:
//
//	router[GlobalAction]
//	|
//	|_Group[GlobalAction, LoggedInAction]
//	|
//	|-Group[LoggedInAction, Teams]
//	|
//	|-Group[LoggedInAction, Settings]
//	|
//	|-Group[LoggedInAction, Admin]
//	  |
//	  |_ Group[Admin, SuperAdmin]
func Subrouter[P Action, T Action, Y registerable[P]](parent Y, prefix string, factory func(context.Context, *GroupContext[P, T])) *Group[P, T] {
	group := NewGroup(parent, factory)
	group.routePrefix = parent.prefix() + prefix

	return group
}

// NewGroup creates a new route Group that can be used to group around
// filters and other common behavior.
//
// A factory function is passed to the NewGroup, so that it can reference
// fields from the parent action type.
func NewGroup[P Action, T Action, Y registerable[P]](parent Y, factory func(context.Context, *GroupContext[P, T])) *Group[P, T] {
	group := &Group[P, T]{routes: make([]*Route[T], 0), actionFactory: factory}
	group.routePrefix = parent.prefix()
	parent.register(group)

	return group
}

// Match defines a new Route that responds to requests that match the given
// method and path.
func (g *Group[P, T]) Match(method string, path string, handler HandlerFunc[T]) {
	if path == "/" {
		path = g.routePrefix
	} else {
		path = joinPath(g.routePrefix, path)
	}

	if path == "" {
		path = "/"
	}

	g.routes = append(g.routes, newRoute(method, path, handler))
}

// Defines a new Route that responds to GET requests.
func (g *Group[P, T]) Get(path string, handler HandlerFunc[T]) {
	g.Match(http.MethodGet, path, handler)
}

// Defines a new Route that responds to POST requests.
func (g *Group[P, T]) Post(path string, handler HandlerFunc[T]) {
	g.Match(http.MethodPost, path, handler)
}

// Defines a new Route that responds to PUT requests.
func (t *Group[P, T]) Put(path string, handler HandlerFunc[T]) {
	t.Match(http.MethodPut, path, handler)
}

// Defines a new Route that responds to PATCH requests.
func (g *Group[P, T]) Patch(path string, handler HandlerFunc[T]) {
	g.Match(http.MethodPatch, path, handler)
}

// Defines a new Route that responds to DELETE requests.
func (g *Group[P, T]) Delete(path string, handler HandlerFunc[T]) {
	g.Match(http.MethodDelete, path, handler)
}

// Implements Dispatchable so groups can be registered on routers
func (g *Group[P, T]) dispatch(ctx context.Context, r *http.Request) (bool, map[string]string, func(context.Context, P)) {
	if route, params := g.routeFor(r); route != nil {
		return true, params, func(ctx context.Context, action P) {
			groupContext := &GroupContext[P, T]{action: action, next: route.handler}
			g.actionFactory(ctx, groupContext)
		}
	}

	for _, group := range g.subgroups {
		if ok, params, handler := group.dispatch(ctx, r); ok {

			return true, params, func(ctx context.Context, action P) {
				groupContext := &GroupContext[P, T]{action: action, next: handler}
				g.actionFactory(ctx, groupContext)
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

// register implements the registerable interface and allows subgroups to be
// registered and routed to.
func (g *Group[P, T]) register(subgroup dispatchable[T]) {
	g.subgroups = append(g.subgroups, subgroup)
}

// prefix implements the registerable interface and allows subgroups to register
// routes with the correct path.
func (g *Group[P, T]) prefix() string {
	return g.routePrefix
}

var trailingSlash = regexp.MustCompile("/+$")
var leadingSlash = regexp.MustCompile("^/+")

func joinPath(left string, right string) string {
	left = trailingSlash.ReplaceAllString(left, "")
	right = leadingSlash.ReplaceAllString(right, "")

	return left + "/" + right
}
