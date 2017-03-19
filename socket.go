package pho

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// SocketOptions provides the socket options
type SocketOptions struct {
	Header    http.Header
	UserAgent string
	Conn      *websocket.Conn
	StopChan  chan struct{}
	OnRequest SocketRequestFunc
}

// SocketRequestFunc is a func called when a data is received by the socket
type SocketRequestFunc func(socket *Socket, req *Request)

// Socket represents a single client connection
// to the RPC server
type Socket struct {
	ID        string
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
		ID:        socketID,
		conn:      options.Conn,
		stopChan:  options.StopChan,
		userAgent: options.Header.Get("User-Agent"),
		fn:        options.OnRequest,
	}

	return socket, nil
}

// run listens for server responses
func (c *Socket) Run() {
	for {
		select {
		case <-c.stopChan:
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			c.conn.Close()
			return
		default:
			if err := c.conn.SetReadDeadline(time.Now().Add(ReadDeadline)); err != nil {
				continue
			}

			msgType, reader, err := c.conn.NextReader()
			if err != nil {
				return
			}

			if msgType == websocket.CloseMessage {
				return
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
