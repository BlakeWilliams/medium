/*
The tell package implements a basic pub/sub system that enables the subscribing
and publishing of events. The primary use-case is for packages to allow a
notifier to be passed so that they can emit events. This allows consumers of the
package to subscribe to events and implement their own logging and observability
interface or implementation.

Here's a basic example of how to use a notifier for logging and tracing:

	notifier := sub.New()
	notifier.Subscribe("web.request.serve", func(e sub.Event) {
		// implement OTEL tracing
		var span trace.Span
		ctx, span := tracer.Start(context.TODO(), fmt.Sprintf("http.request.%s", e.Payload["method"]))
		defer span.End()

		duration := e.FinishedAt.Sub(e.StartedAt)
		logger.Infof("Served HTTP %s for path %s in %s", e.Payload["method"], e.Payload["path"], duration)

	})

	// HTTP middleware that will emit a "web.request.serve" event
	func notifierMiddleware(notifier sub.Notifier) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				event := notifier.Start("web.request.serve", sub.Payload{"method": r.Method, "path": r.URL.Path})
				defer event.Finish()

				next().ServeHTTP(w, r)
			})
		})
	}
*/
package tell
