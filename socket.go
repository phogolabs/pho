package pho

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// SocketOptions provides the socket options
type SocketOptions struct {
	Conn      *websocket.Conn
	UserAgent string
	StopChan  chan struct{}
	OnRequest SocketRequestFunc
}

// SocketRequestFunc is a func called when a data is received by the socket
type SocketRequestFunc func(socket *Socket, req *Request)

// Socket represents a single client connection
// to the RPC server
type Socket struct {
	id        string
	userAgent string
	conn      *websocket.Conn
	stopChan  chan struct{}
	fn        SocketRequestFunc
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
		stopChan:  options.StopChan,
		userAgent: options.UserAgent,
		fn:        options.OnRequest,
	}

	return socket, nil
}

// ID returns the socket identificator
func (c *Socket) ID() string {
	return c.id
}

// run listens for server responses
func (c *Socket) Run() error {
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

			request.UserAgent = c.userAgent
			request.RemoteAddr = c.conn.RemoteAddr().String()
			c.fn(c, request)
		}
	}
}
