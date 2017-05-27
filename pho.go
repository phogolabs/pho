package pho

import (
	"net/http"
)

// ErrorType defines the type of error Response and Request
const ErrorType = "error"

//go:generate counterfeiter -o ./fakes/FakeResponseWriter.go . ResponseWriter
//go:generate counterfeiter -o ./fakes/FakeSocketWriter.go . SocketWriter

// OnConnectFunc called on every connection
type OnConnectFunc func(w SocketWriter, r *http.Request)

// OnDisconnectFunc callend when the client closes connection
type OnDisconnectFunc func(w SocketWriter)

// The MiddlewareFunc type is a middeware contract
type MiddlewareFunc func(Handler) Handler

// The RouterFunc type is a router contract
type RouterFunc func(Router)

// OnErrorFunc called on every server side error
type OnErrorFunc func(err error)

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(SocketWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeRPC(w SocketWriter, r *Request) {
	f(w, r)
}

// A SocketWriter interface is used by an RPC handler to
// construct an RPC response.
type ResponseWriter interface {
	// Write writes to the client initiated the request
	Write(string, int, []byte) error
	// WriteError writes an errors with specified code
	WriteError(err error, code int) error
}

// Metadata of Response Writer
type Metadata map[string]interface{}

// A SocketWriter interface is used by an RPC handler to
// construct an RPC response.
type SocketWriter interface {
	// SocketID
	SocketID() string
	// UserAgent associated with this writer
	UserAgent() string
	// RemoteAddr is the client IP address
	RemoteAddr() string
	// Metadata for this response writer
	Metadata() Metadata
	// Write writes to the client initiated the request
	Write(string, int, []byte) error
	// WriteError writes an errors with specified code
	WriteError(err error, code int) error
}

// A Handler responds to an RPC request.
type Handler interface {
	ServeRPC(SocketWriter, *Request)
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
	ServeRPC(SocketWriter, *Request)

	// Use appends one of more middlewares onto the Router stack.
	Use(middlewares ...MiddlewareFunc)

	// The On-function adds callbacks by name of the event, that should be handled.
	On(verb string, handle HandlerFunc)

	// On-Connect func register callback invoked on each error
	OnError(fn OnErrorFunc)

	// On-Connect func register callback invoked on each connection
	OnConnect(fn OnConnectFunc)

	// On-Disconnect func register callback invoked every time when client is disconnected
	OnDisconnect(fn OnDisconnectFunc)

	// Mount attaches another http.Handler along the channel
	Mount(verb string, handler Handler)

	// Route creates a new Mux with a fresh middleware stack and mounts it
	// along the `pattern` as a subrouter. Effectively, this is a short-hand
	// call to Mount.
	Route(verb string, fn RouterFunc) Router

	// Close stops all connections
	Close()
}
