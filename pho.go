package pho

import (
	"io"
	"net/http"
)

// A ResponseWriter interface is used by an RPC handler to
// construct an RPC response.
type ResponseWriter interface {
	// AsWriter returns a writer that writes to the client via specified channel
	AsWriter(string) io.Writer

	// AsMultiWriter returns a writer via specified channel to all clients
	AsMultiWriter(string) io.Writer

	// Write writes to this client initiated the request
	Write(string, []byte) (int, error)

	// MultiWrite writes the data to the all connections of this channel
	MultiWrite(string, []byte) (int, error)
}

// A Handler responds to an RPC request.
type Handler interface {
	ServeRPC(ResponseWriter, *Request)
}

// The MiddlewareFunc type is a middeware contract
type MiddlewareFunc func(Handler) Handler

// The RouterFunc type is a router contract
type RouterFunc func(Router) Router

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeRPC(w ResponseWriter, r *Request) {
	f(w, r)
}

// NewRouter returns a new Mux object that implements the Router interface.
func NewRouter() Router {
	return NewMux()
}

// Router consisting of the core routing methods used by pho's Mux,
// using only the standard net/http.
type Router interface {
	// A Handler responds to an HTTP request.
	ServeHTTP(http.ResponseWriter, *http.Request)

	// A Handler responds to an RPC request.
	ServeRPC(ResponseWriter, *Request)

	// Use appends one of more middlewares onto the Router stack.
	Use(middlewares ...MiddlewareFunc)

	// The On-function adds callbacks by name of the event, that should be handled.
	On(verb string, handle HandlerFunc)

	// Mount attaches another http.Handler along the channel
	Mount(verb string, handler Handler)

	// Route creates a new Mux with a fresh middleware stack and mounts it
	// along the `pattern` as a subrouter. Effectively, this is a short-hand
	// call to Mount.
	Route(verb string, fn RouterFunc) Router

	// Close stops all connections
	Close()
}
