package pho

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// SocketOptions provides the socket options
type SocketOptions struct {
	Conn      *websocket.Conn
	UserAgent string
	ServeRPC  HandlerFunc
	StopChan  chan struct{}
}

// Socket represents a single client connection
// to the RPC server
type Socket struct {
	id        string
	userAgent string
	conn      *websocket.Conn
	stopChan  chan struct{}
	metadata  Metadata
	serveRPC  HandlerFunc
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
		id:        socketID,
		conn:      options.Conn,
		userAgent: options.UserAgent,
		stopChan:  options.StopChan,
		serveRPC:  options.ServeRPC,
		metadata:  Metadata{},
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
		Body:       bytes.NewBuffer(data),
	}

	return c.write(response)
}

// WriteError writes an errors with specified code
func (c *Socket) WriteError(err error, code int) error {
	response := &Response{
		Verb:       "error",
		StatusCode: code,
		Body:       strings.NewReader(err.Error()),
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

	if err = response.Marshal(writer); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			err = fmt.Errorf("%v: %v", err, closeErr)
		}
		return err
	}
	return writer.Close()
}

// run listens for server responses
func (c *Socket) run() error {
	for {
		select {
		case <-c.stopChan:
			if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
				if connErr := c.conn.Close(); connErr != nil {
					err = fmt.Errorf("%v: %v", err, connErr)
				}
				return err
			}
			return c.conn.Close()
		default:
			if err := c.conn.SetReadDeadline(time.Now().Add(ReadDeadline)); err != nil {
				continue
			}

			msgType, reader, err := c.conn.NextReader()
			if err != nil {
				return err
			}

			if msgType == websocket.CloseMessage {
				return nil
			}

			if msgType != websocket.BinaryMessage {
				continue
			}

			request := &Request{}
			if err := request.Unmarshal(reader); err != nil {
				continue
			}

			c.serveRPC(c, request)
		}
	}
}
