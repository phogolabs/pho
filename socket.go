package pho

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type SocketError struct {
	Error string `json:"error"`
}

// SocketOptions provides the socket options
type SocketOptions struct {
	Conn         *websocket.Conn
	UserAgent    string
	Host         string
	RequestURI   string
	TLS          *tls.ConnectionState
	ServeRPC     HandlerFunc
	OnDisconnect OnDisconnectFunc
	OnError      OnErrorFunc
	StopChan     chan struct{}
}

// Socket represents a single client connection
// to the RPC server
type Socket struct {
	id             string
	userAgent      string
	host           string
	requestUri     string
	tls            *tls.ConnectionState
	conn           *websocket.Conn
	stopChan       chan struct{}
	metadata       Metadata
	serveRPCFn     HandlerFunc
	onDisconnectFn OnDisconnectFunc
	onErrorFn      OnErrorFunc
}

// NewSocket creates a new socket
func NewSocket(options *SocketOptions) (*Socket, error) {
	var (
		socketID string
		err      error
	)

	if socketID, err = RandString(20); err != nil {
		return nil, err
	}

	socket := &Socket{
		id:             socketID,
		tls:            options.TLS,
		conn:           options.Conn,
		userAgent:      options.UserAgent,
		host:           options.Host,
		requestUri:     options.RequestURI,
		stopChan:       options.StopChan,
		serveRPCFn:     options.ServeRPC,
		onDisconnectFn: options.OnDisconnect,
		onErrorFn:      options.OnError,
		metadata:       Metadata{},
	}

	return socket, nil
}

// ID returns the socket identificator
func (c *Socket) SocketID() string {
	return c.id
}

// Metadata for this socket
func (c *Socket) Metadata() Metadata {
	return c.metadata
}

// Write a reponse
func (c *Socket) Write(responseType string, status int, data []byte) error {
	response := &Response{
		Type:       responseType,
		StatusCode: status,
		Payload:    data,
	}

	return c.write(response)
}

// WriteError writes an errors with specified code
func (c *Socket) WriteError(err error, code int) error {
	body, _ := json.Marshal(&SocketError{
		Error: err.Error(),
	})

	response := &Response{
		Type:       "error",
		StatusCode: code,
		Payload:    body,
	}

	c.onErrorFn(err)
	return c.write(response)
}

// The client user agent
func (c *Socket) UserAgent() string {
	return c.userAgent
}

// Host
func (c *Socket) EndpointAddr() string {
	return fmt.Sprintf("%s%s", c.host, c.requestUri)
}

// TLS
func (c *Socket) TLS() *tls.ConnectionState {
	return c.tls
}

// RemoteAddr provides client IP
func (c *Socket) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *Socket) write(response *Response) error {
	writer, err := c.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(writer)
	enc.SetEscapeHTML(true)

	if err = enc.Encode(response); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			err = fmt.Errorf("%v: %v", err, closeErr)
		}
		return err
	}
	return writer.Close()
}

// run listens for server responses
func (c *Socket) run() {
	for {
		select {
		case <-c.stopChan:
			c.onDisconnectFn(c)
			c.onErrorFn(c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(30*time.Second)))
			c.onErrorFn(c.conn.Close())
			return
		default:
			if err := c.conn.SetReadDeadline(time.Now().Add(ReadDeadline)); err != nil {
				c.onErrorFn(err)
				continue
			}

			msgType, reader, err := c.conn.NextReader()
			if err != nil {
				c.onDisconnectFn(c)
				c.onErrorFn(c.conn.Close())
				return
			}

			if msgType != websocket.TextMessage && msgType != websocket.BinaryMessage {
				continue
			}

			request := &Request{}
			if err := json.NewDecoder(reader).Decode(request); err != nil {
				c.onErrorFn(err)
				continue
			}

			c.serveRPCFn(c, request)
		}
	}
}
