# Medium

Experimental Go code for writing web applications. This is a collection of
packages I ocassionally hack on to write some Go and play around with, trying to
use Go as a web application framework.

These packages are likely to change often and without warning given the (current) experimental nature.

## Packages

- **pkg/rescue** - Basic rescue middleware for router.
- **pkg/httplogger** - Basic logger middleware for router.
- ~**pkg/view** - Wraps the [`html/template`](https://golang.org/pkg/html/template/) package to provide a slightly more friendly and ergonimic interface for web application usage.~ Use [bat](https://github.com/blakewilliams/bat) instead.:e
- **pkg/session** - Struct based, cookie backed session management using HMAC signatures to validate session contents.
- **pkg/mail** - Provides a basic mailer package that utilizes `pkg/template` for templating. Additionally provides a basic interface that can be used with `pkg/router` to see sent emails in development.
- **pkg/mlog** - Simple structured logger usable directly, or through context compatible API's.
- **pkg/set** - Basic Set data structure.
- **pkg/webpack** - Middleware that allows you to use webpack to serve assets in development.

## Medium (web framework)

Formerly `pkg/router`, this is a basic router that provides a few basic features:

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
// In medium "context" is called an action. It's typically a user provided type
// with app specific fields and app specific behavior.
type AppAction struct {
  currentUser *User
  medium.Action // Embed medium.Action to adhere to the router action constraint and get some default behavior
}

func (a *AppAction) Render(w io.Writer, templateName string, data interface{}) error {
  // Render a template using the app specific template engine
  template.New(templateName).ParseFiles("./templates/"+templateName).Execute(w, data)
}

// Create a new router. Routers, groups, and subrouters accept an "action
// factory" function that allows you to convert the previous action type into your custom action type.
//
// This is also where you write code that is typically handled by
// before/after/around actions in other frameworks, which is code that is meant
// to run before or after the route handler is called.
router := medium.New(func(a *medium.BaseAction) AppAction {
  currentUser := findCurrentUser(a.Request)
  return AppAction{Action: a, currentUser: currentUser}
})

// Add a hello route
router.Get("/hello/:name", func(a AppAction) {
  a.Render(a, "hello.html", map[string]any{"name": a.Params["name"]})
})

fmt.Println("Listening on :8080")
server := http.Server{Addr: ":8080", Handler: router}
_ = server.ListenAndServe()
```

### Groups and Subrouters

Groups and subrouters allow you to consolidate behavior at the route level. For
example, you can create a group that requires a user to be logged in, and then
add routes to that group. If the user is not logged in, the group will redirect
to the login page.

```go
// Create a new router
router := medium.New(func(a *medium.BaseAction) AppAction {
  currentUser := findCurrentUser(a.Request)
  return AppAction{Action: a, currentUser: currentUser}
})

// Create a group that requires a user to be logged in
authGroup := router.Group(func(a AppAction, next func(a AppAction)) {
  if a.currentUser != nil {
    a.Render404()
    return
  }
})

// Add a route to the group that will redirect if the user is not logged in
authGroup.Get("/welcome", func(a AppAction) {
  a.Render(a, "hello.html", map[string]any{"CurrentUser": a.currentUser})
})
```

Subrouters are similar to groups, but allow you to create a new router that
has a path prefix. This is useful for things like API versioning, or
requiring a specific resource to be present and authorized.

```go
// Create a new router
router := medium.New(func(a *medium.BaseAction) AppAction {
  currentUser := findCurrentUser(a.Request)
  return AppAction{Action: a, currentUser: currentUser}
})

// Create a type that will hold on to the current team
type TeamAction struct {
  // Embed AppAction to inherit all of the fields and methods
  AppAction
  currentTeam *Team
}

// Create a subrouter that ensures a team is present and authorized
teamRouter := router.Subrouter("/teams/:teamID", func(a AppAction, next func(a TeamAction)) {
  team := findTeam(a.Params["teamID"])
  if team == nil {
    a.Render404()
    return
  }

  if !team.IsMember(a.currentUser) {
    a.Render403()
    return
  }

  a.Team = team
  next(TeamAction{AppAction: a, currentTeam: team})
})

// Add a route to render the team show page
teamRouter.Get("/", func(a TeamAction) {
  a.Render(a, "team.html", map[string]any{"Team": a.currentTeam})
})
```

This is a really powerful way to compose routes, ensuring that the correct
resources are available and authorized before the route handler is called.

### Middleware

Middleware allows you to add generic behavior to the router. This is useful for
adding behavior like logging, tracing, tracking request ID's, rescuing and
reporting exceptions, etc.

```go
// Create a new router
router := medium.New(func(a *medium.BaseAction) AppAction {
  currentUser := findCurrentUser(a.Request)
  return AppAction{Action: a, currentUser: currentUser}
})

// Add a middleware that logs the request
// Note: Middleware accept medium.Action, not your custom action type.
router.Use(func(a medium.Action, next func(a medium.Action)) {
  now := time.Now()
  log.Printf("Started: %s %s", a.Request.Method, a.Request.URL.Path)

  next(a)

  log.Printf("Served: %s %s in %s", a.Request.Method, a.Request.URL.Path, time.Since(now))
})
```

## Contributing

Contributions are welcome via pull requests and issues.
