package utils

import "net/http"

// Middleware is a function that wraps an http handler with additional functionality.
type Middleware func(http.Handler) http.Handler

func ApplyMiddleWares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}