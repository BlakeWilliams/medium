/*
The mlog package provides a simple structured logger.

Example:
	package main

	import "mlog"

	func main() {
		listen := ":8080"
		logger := mlog.New(os.Stdout, mlog.JSONFormatter())
		logger.Debug("Starting application", mlog.Fields{"listen": port})

		// Expose logger with default fields to context
		ctx := mlog.Inject(ctx, logger.WithDefaults(mlog.Fields{"app": "myapp"})

		doSomething(ctx)
	}

	func doSomething(ctx) {
		// Use the logger on ctx if provided
		mlog.Info(ctx, "doing something", mlog.Fields{})
	}
*/
package mlog
