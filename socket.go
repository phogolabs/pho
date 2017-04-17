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
func (c *Socket) Write(responseType string, data []byte) error {
	response := &Response{
		Type:       responseType,
		StatusCode: http.StatusOK,
		Body:       data,
	}

	return c.write(response)
}

// WriteJSON encodes and writes JSON to the client
func (c *Socket) WriteJSON(responseType string, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return c.Write(responseType, data)
}

// WriteError writes an errors with specified code
func (c *Socket) WriteError(err error, code int) error {
	response := &Response{
		Type:       "error",
		StatusCode: code,
		Body:       []byte(err.Error()),
	}

	c.onErrorFn(err)
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

			if msgType != websocket.BinaryMessage {
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
