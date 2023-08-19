package medium

import (
	"context"
	"net/http"
	"regexp"
)

// WithNoData is a helper function that can be used to create a router that
// doesn't require any new data to be created.
func WithExistingData[T any](data T) func() T {
	return func() T {
		return data
	}
}

// registerable represents a type that can be registered on a router or a group
// to create subgroups/subrouters.
type registerable[Data any] interface {
	register(dispatchable[Data])
	prefix() string
}

// RouteGroup represents a collection of routes that share a common set of
// Around/Before/After callbacks and Action type (T)
type RouteGroup[ParentData any, Data any] struct {
	routes        []*Route[Data]
	actionCreator func(ParentData) Data
	subgroups     []dispatchable[Data]
	routePrefix   string
}

// SubRouter creates a new grouping of routes that will be routed to in addition
// to the routes defined on the primery router. These routes will be prefixed
// using the given prefix.
//
// Subrouters provide their own action creator, so common behavior can be
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
func SubRouter[ParentData any, Data any, Parent registerable[ParentData]](parent Parent, prefix string, creator func(ParentData) Data) *RouteGroup[ParentData, Data] {
	group := Group(parent, creator)
	group.routePrefix = parent.prefix() + prefix

	return group
}

// NewGroup creates a new route Group that can be used to group around
// filters and other common behavior.
//
// An action creator function is passed to the NewGroup, so that it can reference
// fields from the parent action type.
func Group[ParentData any, Data any, Parent registerable[ParentData]](parent Parent, creator func(ParentData) Data) *RouteGroup[ParentData, Data] {
	group := &RouteGroup[ParentData, Data]{routes: make([]*Route[Data], 0), actionCreator: creator}
	group.routePrefix = parent.prefix()
	parent.register(group)

	return group
}

// Match defines a new Route that responds to requests that match the given
// method and path.
func (g *RouteGroup[ParentData, Data]) Match(method string, path string, handler HandlerFunc[Data]) {
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
func (g *RouteGroup[ParentData, Data]) Get(path string, handler HandlerFunc[Data]) {
	g.Match(http.MethodGet, path, handler)
}

// Defines a new Route that responds to POST requests.
func (g *RouteGroup[ParentData, Data]) Post(path string, handler HandlerFunc[Data]) {
	g.Match(http.MethodPost, path, handler)
}

// Defines a new Route that responds to PUT requests.
func (t *RouteGroup[ParentData, Data]) Put(path string, handler HandlerFunc[Data]) {
	t.Match(http.MethodPut, path, handler)
}

// Defines a new Route that responds to PATCH requests.
func (g *RouteGroup[ParentData, Data]) Patch(path string, handler HandlerFunc[Data]) {
	g.Match(http.MethodPatch, path, handler)
}

// Defines a new Route that responds to DELETE requests.
func (g *RouteGroup[ParentData, Data]) Delete(path string, handler HandlerFunc[Data]) {
	g.Match(http.MethodDelete, path, handler)
}

// Implements Dispatchable so groups can be registered on routers
func (g *RouteGroup[ParentData, Data]) dispatch(rootRequest RootRequest) (bool, map[string]string, func(context.Context, Request[ParentData]) Response) {
	if route, params := g.routeFor(rootRequest); route != nil {
		return true, params, func(ctx context.Context, req Request[ParentData]) Response {
			data := g.actionCreator(req.Data)
			return route.handler(ctx, Request[Data]{root: rootRequest, routeParams: params, Data: data})
		}
	}

	for _, group := range g.subgroups {
		if ok, params, handler := group.dispatch(rootRequest); ok {
			return true, params, func(ctx context.Context, req Request[ParentData]) Response {
				data := g.actionCreator(req.Data)
				return handler(ctx, Request[Data]{root: rootRequest, routeParams: params, Data: data})
			}
		}
	}

	return false, nil, nil
}

func (g *RouteGroup[ParentData, Data]) routeFor(req RootRequest) (*Route[Data], map[string]string) {
	for _, route := range g.routes {
		if ok, params := route.IsMatch(req); ok {
			return route, params
		}
	}

	return nil, nil
}

// register implements the registerable interface and allows subgroups to be
// registered and routed to.
func (g *RouteGroup[ParentData, Data]) register(subgroup dispatchable[Data]) {
	g.subgroups = append(g.subgroups, subgroup)
}

// prefix implements the registerable interface and allows subgroups to register
// routes with the correct path.
func (g *RouteGroup[ParentData, Data]) prefix() string {
	return g.routePrefix
}

var trailingSlash = regexp.MustCompile("/+$")
var leadingSlash = regexp.MustCompile("^/+")

func joinPath(left string, right string) string {
	left = trailingSlash.ReplaceAllString(left, "")
	right = leadingSlash.ReplaceAllString(right, "")

	return left + "/" + right
}
