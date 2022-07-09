/*
The router package provides a simple interface that maps routes to handlers. The router accepts a type argument which must implement the Action interface. This allows clients to define extend and enhance the BaseAction type with additional data and functionality.

Example:

	package main

	type MyAction struct {
		requestId string
		*router.BaseAction
	}

	func Run() {
		// Router that defines a context factory that returns a new MyAction for each request.
		router := New(func(ac *MyAction) {
			ac.requestId = randomString()
		}))

		// Add an Around handler that sets the requestId header
		router.Around(func(ac *MyAction, next func()) {
			ac.Response.Header().Add("X-Request-ID", ac.requestId)
			next()
		})

		// Echo back the requestId header
		router.Get("/echo", func(ac *MyAction) {
			ac.Write([]byte(ac.requestId))
		})

		http.ListenAndServe(":8080", router)
	}
*/
package medium
