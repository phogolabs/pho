package pho

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

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
	rw *sync.RWMutex
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
	// onDisconnectFn called after each connection is closed
	onDisconnectFn OnDisconnectFunc
	// onErrorFn called after each error
	onErrorFn OnErrorFunc
	// stopChan stops all sockets
	stopChan chan struct{}
}

// NewMux creates an instance of *Mux
func NewMux() *Mux {
	return &Mux{
		rw:          &sync.RWMutex{},
		handlers:    map[string]Handler{},
		sockets:     WebSockets{},
		middlewares: []MiddlewareFunc{},
		upgrader: &websocket.Upgrader{
			CheckOrigin:       func(r *http.Request) bool { return true },
			EnableCompression: true,
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
		},
		stopChan: make(chan struct{}),
	}
}

// ServeRPC is the single method of the pho.Handler interface that makes
// Mux nestable in order to build hierarchies
func (m *Mux) ServeRPC(w SocketWriter, r *Request) {
	attrb := strings.SplitN(r.Type, ":", 2)
	verb := strings.ToLower(attrb[0])

	if len(attrb) > 1 {
		r.Type = attrb[1]
	}

	if verb == ErrorType {
		err := fmt.Errorf("%s", string(r.Body))
		m.handleError(err)
	}

	handler, ok := m.handlers[verb]
	if !ok {
		err := w.WriteError(fmt.Errorf("The route %q does not exist", r.Type), http.StatusNotFound)
		m.handleError(err)
		return
	}

	m.prepareWriter(w)
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
		m.handleError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	socket, err := NewSocket(&SocketOptions{
		UserAgent:    r.UserAgent(),
		Conn:         conn,
		OnDisconnect: m.removeSocket,
		OnError:      m.handleError,
		ServeRPC:     m.ServeRPC,
		StopChan:     m.stopChan,
	})

	if err != nil {
		m.handleError(err)
		m.handleError(conn.Close())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.rw.Lock()
	m.sockets[socket.SocketID()] = socket
	m.rw.Unlock()

	go socket.run()

	if m.onConnectFn != nil {
		m.prepareWriter(socket)
		m.onConnectFn(socket, r)
	}
}

// Use appends one of more middlewares onto the Router stack.
func (m *Mux) Use(middlewares ...MiddlewareFunc) {
	m.middlewares = append(m.middlewares, middlewares...)
}

// On registers a handler for particular type of request
func (m *Mux) On(method string, handler HandlerFunc) {
	m.handlers[strings.ToLower(method)] = handler
}

// OnConnect register a callback function called on error
func (m *Mux) OnError(fn OnErrorFunc) {
	m.onErrorFn = fn
}

// OnConnect register a callback function called on conection
func (m *Mux) OnConnect(fn OnConnectFunc) {
	m.onConnectFn = fn
}

// OnDisconnect register a callback function called on disconnect
func (m *Mux) OnDisconnect(fn OnDisconnectFunc) {
	m.onDisconnectFn = fn
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
	m.stopChan = make(chan struct{})
}

func (m *Mux) prepareWriter(w SocketWriter) {
	m.rw.RLock()
	w.Metadata()[MetadataSocketKey] = Copy(m.sockets)
	m.rw.RUnlock()
}

func (m *Mux) handleError(err error) {
	if err == nil {
		return
	}

	if m.onErrorFn != nil {
		m.onErrorFn(err)
	}
}

func (m *Mux) removeSocket(w SocketWriter) {
	m.rw.Lock()
	delete(m.sockets, w.SocketID())
	m.rw.Unlock()

	if m.onDisconnectFn != nil {
		m.prepareWriter(w)
		m.onDisconnectFn(w)
	}
}
