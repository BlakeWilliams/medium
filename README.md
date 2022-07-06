# Medium

Experimental Go code for writing web applications. This is a collection of packages I ocassionally hack on to write some Go and play around with using Go to build robust web applications.

These packages are likely to change often and without warning given the (current) experimental nature.

## Packages

- **pkg/router** - A simple router package that allows you to define routes and handle requests using your own custom types.
  - **pkg/router/rescue** - Basic rescue handler for router.
- **pkg/view** - Wraps the [`html/template`](https://golang.org/pkg/html/template/) package to provide a slightly more friendly and ergonimic interface for web application usage.
- **pkg/session** - Struct based, cookie backed session management using HMAC signatures to validate session contents.
- **pkg/mail** - Provides a basic mailer package that utilizes `pkg/template` for templating. Additionally provides a basic interface that can be used with `pkg/router` to see sent emails in development.
- **pkg/mlog** - Simple structured logger usable directly, or through context compatible API's.
- **pkg/set** - Basic Set data structure.
- **pkg/tell** - A simple subscription system that allows you to define subscriptions and handle requests using your own custom types. Good for implementing custom logs, tracing, etc.
- **pkg/webpack** - Middleware that allows you to use webpack to serve assets in development.

## Contributing

Contributions are welcome via pull requests and issues.
