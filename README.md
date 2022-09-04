# Medium

Experimental Go code for writing web applications. This is a collection of packages I ocassionally hack on to write some Go and play around with, trying to use Go as a web application framework.

These packages are likely to change often and without warning given the (current) experimental nature.

## Packages

- **pkg/rescue** - Basic rescue middleware for router.
- **pkg/httplogger** - Basic logger middleware for router.
- ~**pkg/view** - Wraps the [`html/template`](https://golang.org/pkg/html/template/) package to provide a slightly more friendly and ergonimic interface for web application usage.~ Use [bat](https://github.com/blakewilliams/bat) instead.
- **pkg/session** - Struct based, cookie backed session management using HMAC signatures to validate session contents.
- **pkg/mail** - Provides a basic mailer package that utilizes `pkg/template` for templating. Additionally provides a basic interface that can be used with `pkg/router` to see sent emails in development.
- **pkg/mlog** - Simple structured logger usable directly, or through context compatible API's.
- **pkg/set** - Basic Set data structure.
- **pkg/webpack** - Middleware that allows you to use webpack to serve assets in development.

## Contributing

Contributions are welcome via pull requests and issues.
