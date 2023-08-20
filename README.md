# Medium

Experimental Go code for writing web applications. This is a collection of
packages I ocassionally hack on to write some Go and play around with, trying to
use Go as a web application framework.

These packages are likely to change often and without warning given the (current) experimental nature.

## Packages

- **middleware/rescue** - Basic rescue middleware for router.
- **middleware/httpmethod** - Rewrites the HTTP method based on the \_method parameter. This is used to allow browsers to make PUT, PATCH, and DELETE requests.
- **middleware/httplogger** - Basic logger middleware for router.
- ~**view** - Wraps the [`html/template`](https://golang.org/html/template/) package to provide a slightly more friendly and ergonimic interface for web application usage.~ Use [bat](https://github.com/blakewilliams/bat) instead.
- **session** - Struct based, cookie backed session management using HMAC signatures to validate session contents.
- **mail** - Provides a basic mailer package that utilizes `template` for templating. Additionally provides a basic interface that can be used with `router` to see sent emails in development.
- **mlog** - Simple structured logger usable directly, or through context compatible API's.
- **set** - Basic Set data structure.
- **webpack** - Middleware that allows you to use webpack to serve assets in development.

## Medium (web framework)

Formerly `router`, this is a basic router that provides a few basic features:

- **Middleware** - Middleware allows you to change the request and response
  before and after the handler is called. (logging, authentication, session
  management, etc.)
- **Custom Handler Types** - Most other frameworks pass their
  own `context` object. In medium, generics are used to allow you to define your
  own handler types. This allows you to pass in any type of context you want, and
  have it be type safe.
- **Subrouters and Groups** - Medium allows you to
  consolidate behavior at the route level, allowing you to create subrouters for
  things like authentication, API versioning, or requiring a specific resource
  to be present and authorized.

### Getting started

To get started, install medium via `go get github.com/blakewilliams/medium`.

From there, you can create a new router and add a handler:

```go
import (
  "fmt"
  "html/template"
  "net/http"
  "github.com/blakewilliams/medium"
)
// Requests in medium can store a generic data type that is passed to each
// BeforeFunc and handler. This is useful for storing things like the current
// user, global rendering data, etc.
type ReqData struct {
  currentUser *User
}

// Routers are generic and must specify the type of Data they will pass to
// HandlerFunc/BeforeFuncs.
router := medium.New(func(req RootRequest) *ReqData {
  return &ReqData
})
// Fill in ReqData with the current user before each request. The BeforeFunc
// must return a Response that will be used to render the response. Calling next
// will continue to the next BeforeFunc or HandlerFunc.
router.Before(func(ctx context.Context, req *medium.Request[ReqData], next medium.Next) Response{
  req.Data.currentUser = findCurrentUser(req.Request)
  return next(ctx)
})

// Add a hello route
router.Get("/hello/:name", func(ctx context.Context, req *medium.Request[ReqData]) Response {
  return Render(req, "hello.html", map[string]any{"name": req.Params["name"], "currentUser": req.Data.currentUser})
})

fmt.Println("Listening on :8080")
server := http.Server{Addr: ":8080", Handler: router}
_ = server.ListenAndServe()
```

### Groups and Subrouters

Groups and subrouters allow you to consolidate behavior at the route level. For
example, you can create a group that requires a user to be logged in.

```go
router := medium.New(func(req RootRequest) *ReqData {
  return &ReqData
})

router.Before(func(ctx context.Context, req *medium.Request[ReqData], next medium.Next) Response {
  req.Data.currentUser = findCurrentUser(req.Request)
  return next(ctx)
})

// Create a group that requires a user to be logged in
authGroup := router.Group(func(r *medium.Request[ReqData]) *ReqData {})
authGroup.Before(func(ctx context.Context, req *medium.Request[ReqData], next medium.Next) Response {
  // If there is no current user, return a 404
  if a.currentUser != nil {
    res := medium.NewResponse()
    res.WriteStatus(http.StatusNotFound)
    res.WriteString("Not Found")

    return res
  }

  // Otherwise, continue to the next BeforeFunc/HandlerFunc
  return next(ctx)
}

// Add a route to the group that will redirect if the user is not logged in
authGroup.Get("/welcome", func(ctx context.Context, req *medium.Request[ReqData]) Response {
  return Render(ctx, "hello.html", map[string]any{"CurrentUser": a.currentUser})
})
```

Subrouters are similar to groups, but allow you to create a new router that
has a path prefix. This is useful for patterns like API versioning or requiring
a specific resource to be present and authorized.

```go
// Create a new router
router := medium.New(func(req RootRequest) *ReqData {
  currentUser := findCurrentUser(req.Request)
  return &ReqData{currentUser: currentUser}
})

// Create a type that will hold on to the current team
type TeamData struct {
  currentTeam *Team
  // Embed parent data type if you want to access the current user, or pass it
  // explicitly in the data creator function passed to SubRouter
  ReqData
}

// Create a subrouter that ensures a team is present and authorized
teamRouter := router.SubRouter("/teams/:teamID", func(r *medium.Request[ReqData]) *TeamData {
  team := findTeam(r.Params["teamID"])
  return &TeamData{ReqData: data, currentTeam: team}
})

// Ensure routes in the team router have a current team and that the current
// user is a member of the team
teamRouter.Before(func(ctx context.Context, req *medium.Request[TeamData], next medium.Next) Response {
  // If there is no current team, return a 404
  if req.Data.currentTeam == nil {
    res := medium.NewResponse()
    res.WriteStatus(http.StatusNotFound)
    res.WriteString("Not Found")

    return res
  }

  // If the current user is not a member of the team, return a 403
  if !team.IsMember(a.currentUser) {
    res := medium.NewResponse()
    res.WriteStatus(http.StatusForbidden)
    res.WriteString("Forbidden")

    return res
  }

  // Otherwise, continue to the next BeforeFunc/HandlerFunc
  return next(ctx)
})

// Add a route to render the team show page
teamRouter.Get("/", func (ctx context.Context, req *medium.Request[TeamData]) Response {
  return Render(ctx, "team.html", map[string]any{"Team": req.Data.currentTeam})
})


// Add a subrouter to the team subrouter that will render the team settings page
// if the current user is an admin
teamSettingsRouter := teamRouter.SubRouter("/settings", func(r *medium.Request[TeamData]) *TeamData { return r.Data })
teamSettingsRouter.Before(func(ctx context.Context, req *medium.Request[TeamData], next medium.Next) Response {
  if !r.Data.currentTeam.IsAdmin(a.currentUser) {
    res := medium.NewResponse()
    res.WriteStatus(http.StatusForbidden)
    res.WriteString("Forbidden")
  }

  return next(ctx)
})
```

This allows for flexible and safe composition of routes based on the current
state of the request.

### Middleware

Middleware are functions that use the Go `http` package types to modify the
request and response before and after the handler is called. This is useful for
compatibility with existing Go middleware packages and for adding generic
behavior to the router.

```go
// Create a new router
router := medium.New(func(req RootRequest) *ReqData {
  currentUser := findCurrentUser(req.Request)
  return &ReqData{currentUser: currentUser}
})

// Add a middleware that logs the request. Middleware work on raw HTTP types, not medium types.
router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  now := time.Now()
  log.Printf("Started: %s %s", a.Request.Method, a.Request.URL.Path)

  next(a)

  log.Printf("Served: %s %s in %s", a.Request.Method, a.Request.URL.Path, time.Since(now))
})
```

## Contributing

Contributions are welcome via pull requests and issues.
