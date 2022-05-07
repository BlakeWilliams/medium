# Medium

Experimental Go code for writing web applications. This is a collection of packages I ocassionally hack on to write some Go and play around with using Go to build robust web applications.

These packages are likely to change often and without warning given the (current) experimental nature.

## Packages

* **pkg/router** - A simple router package that allows you to define routes and handle requests using your own custom types.
* **pkg/template** - Wraps the [`html/template`](https://golang.org/pkg/html/template/) package to provide a slightly more friendly and ergonimic interface for web application usage.
* **pkg/session** - Struct based, cookie backed session management using HMAC signatures to validate session contents.

## Contributing

Contributions are welcome via pull requests and issues.
