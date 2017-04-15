package pho

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// SocketOptions provides the socket options
type SocketOptions struct {
	Conn         *websocket.Conn
	UserAgent    string
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
		conn:           options.Conn,
		userAgent:      options.UserAgent,
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
func (c *Socket) Write(verb string, data []byte) error {
	response := &Response{
		Verb:       verb,
		StatusCode: http.StatusOK,
		Body:       data,
	}

	return c.write(response)
}

// WriteError writes an errors with specified code
func (c *Socket) WriteError(err error, code int) error {
	response := &Response{
		Verb:       "error",
		StatusCode: code,
		Body:       []byte(err.Error()),
	}

	return c.write(response)
}

// The client user agent
func (c *Socket) UserAgent() string {
	return c.userAgent
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

	if err = json.NewEncoder(writer).Encode(response); err != nil {
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
			err := c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(30*time.Second))
			if connErr := c.conn.Close(); connErr != nil {
				if err != nil {
					err = fmt.Errorf("%v: %v", err, connErr)
				} else {
					err = connErr
				}
			}
			c.handleError(err)
			return
		default:
			if err := c.conn.SetReadDeadline(time.Now().Add(ReadDeadline)); err != nil {
				continue
			}

			msgType, reader, err := c.conn.NextReader()
			if err != nil {
				c.onDisconnectFn(c)
				err := c.conn.Close()
				c.handleError(err)
				return
			}

			if msgType != websocket.BinaryMessage {
				continue
			}

			request := &Request{}
			if err := json.NewDecoder(reader).Decode(request); err != nil {
				continue
			}

			c.serveRPCFn(c, request)
		}
	}
}

func (c *Socket) handleError(err error) {
	if c.onErrorFn != nil {
		c.onErrorFn(err)
	}
}
