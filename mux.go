package pho

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// Sockets is the list of all sockets
type WebSockets map[string]ResponseWriter

// Mux is a simple WebSocket route multiplexer
//
// Mux is designed to be fast, minimal and offer a powerful API for building
// modular and composable HTTP services with a large set of handlers. It's
// particularly useful for writing large REST API services that break a handler
// into many smaller parts composed of middlewares and end handlers.
type Mux struct {
	// sockets is the list of all available sockets
	sockets WebSockets
	// The websocket upgrader
	upgrader *websocket.Upgrader
	// The handlers stack
	handlers map[string]Handler
	// The middleware stack
	middlewares []MiddlewareFunc
	// onConnectFn called after each new connection
	onConnectFn OnConnectFunc
	// stopChan stops all sockets
	stopChan chan struct{}
}

// NewMux creates an instance of *Mux
func NewMux() *Mux {
	return &Mux{
		handlers:    map[string]Handler{},
		sockets:     WebSockets{},
		middlewares: []MiddlewareFunc{},
		upgrader: &websocket.Upgrader{
			EnableCompression: true,
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
		},
		stopChan: make(chan struct{}),
	}
}

// ServeRPC is the single method of the pho.Handler interface that makes
// Mux nestable in order to build hierarchies
func (m *Mux) ServeRPC(w ResponseWriter, r *Request) {
	handler, ok := m.handlers[strings.ToLower(r.Verb)]
	if !ok {
		if err := w.WriteError(fmt.Errorf("The route %q does not exist", r.Verb), http.StatusNotFound); err != nil {
			log.Println(err)
		}
		return
	}

	w.Metadata()["Sockets"] = m.sockets
	handler = Chain(m.middlewares, handler)
	handler.ServeRPC(w, r)
}

// ServeHTTP is the single method of the http.Handler interface that makes
// Mux interoperable with the standard library.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	header := http.Header{}
	if protocols := websocket.Subprotocols(r); len(protocols) > 0 {
		header = http.Header{"Sec-Websocket-Protocol": {protocols[0]}}
	}

	conn, err := m.upgrader.Upgrade(w, r, header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	socket, err := NewSocket(&SocketOptions{
		UserAgent: r.UserAgent(),
		Conn:      conn,
		ServeRPC:  m.ServeRPC,
		StopChan:  m.stopChan,
	})

	if err != nil {
		if connErr := conn.Close(); connErr != nil {
			err = fmt.Errorf("%s: %s", err, connErr)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.sockets[socket.SocketID()] = socket
	go func() {
		if err := socket.run(); err != nil {
			log.Println(err)
		}
	}()

	if m.onConnectFn != nil {
		m.onConnectFn(socket, r)
	}
}

// Use appends one of more middlewares onto the Router stack.
func (m *Mux) Use(middlewares ...MiddlewareFunc) {
	m.middlewares = append(m.middlewares, middlewares...)
}

func (m *Mux) On(method string, handler HandlerFunc) {
	m.handlers[strings.ToLower(method)] = handler
}

// OnConnect register a callback function called on conection
func (m *Mux) OnConnect(fn OnConnectFunc) {
	m.onConnectFn = fn
}

// Mount attaches another http.Handler along the channel
func (m *Mux) Mount(method string, handler Handler) {
	m.handlers[strings.ToLower(method)] = handler
}

// Route creates a new Mux with a fresh middleware stack and mounts it
// along the `pattern` as a subrouter. Effectively, this is a short-hand
// call to Mount.
func (m *Mux) Route(verb string, fn RouterFunc) Router {
	mux := NewRouter()

	if fn != nil {
		fn(mux)
	}

	m.Mount(verb, mux)
	return mux
}

// Close stops all connections
func (m *Mux) Close() {
	close(m.stopChan)
}
