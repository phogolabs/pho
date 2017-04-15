package pho

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	WriteDeadline = 10 * time.Second

	// Time allowed to read the next message from the peer.
	ReadDeadline = 60 * time.Second
)

// OnResponseFunc is a callback function that occurs when response is received
type OnResponseFunc func(r *Response)

// A Client is an RPC client.
type Client struct {
	handlers     map[string]OnResponseFunc
	rw           *sync.RWMutex
	stopChan     chan struct{}
	onResponseFn OnResponseFunc
	onErrorFn    OnErrorFunc
	conn         *websocket.Conn
}

// Dial creates a new client connection. Use requestHeader to specify the
// origin (Origin), subprotocols (Sec-WebSocket-Protocol) and cookies (Cookie).
// Use the response.Header to get the selected subprotocol
// (Sec-WebSocket-Protocol) and cookies (Set-Cookie).
func Dial(url string, header http.Header) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		return nil, err
	}

	conn.SetReadLimit(0)

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(ReadDeadline))
	})

	client := &Client{
		handlers: map[string]OnResponseFunc{},
		rw:       &sync.RWMutex{},
		stopChan: make(chan struct{}),
		conn:     conn,
	}

	go client.run()

	return client, nil
}

// Write writes data to the server
func (c *Client) Write(requestType string, body []byte) error {
	return c.Do(&Request{
		Type: requestType,
		Body: body,
	})
}

// Do sends an RPC request and returns an RPC response
func (c *Client) Do(req *Request) error {
	if req.Type == "" {
		return fmt.Errorf("The Request does not have verb")
	}

	w, err := c.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	err = json.NewEncoder(w).Encode(req)
	errClose := w.Close()
	if err != nil && errClose != nil {
		err = fmt.Errorf("%v: %v", err, errClose)
	}
	return err
}

func (c *Client) OnResponse(fn OnResponseFunc) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.onResponseFn = fn
}

func (c *Client) OnError(fn OnErrorFunc) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.onErrorFn = fn
}

// On register callback function called when response with provided verb occurs
func (c *Client) On(verb string, fn OnResponseFunc) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.handlers[strings.ToLower(verb)] = fn
}

// Close closes the connection to the server
func (c *Client) Close() {
	close(c.stopChan)
}

// run listens for server responses
func (c *Client) run() {
	for {
		select {
		case <-c.stopChan:
			err := c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(30*time.Second))
			c.handleError(err)
		default:
			if err := c.conn.SetReadDeadline(time.Now().Add(ReadDeadline)); err != nil {
				continue
			}

			msgType, reader, err := c.conn.NextReader()
			if err != nil {
				err := c.conn.Close()
				c.handleError(err)
				return
			}

			if msgType != websocket.BinaryMessage {
				continue
			}

			response := &Response{}
			if err := json.NewDecoder(reader).Decode(response); err != nil {
				continue
			}

			c.rw.RLock()
			if c.onResponseFn != nil {
				c.onResponseFn(response)
			}

			handler, ok := c.handlers[response.Type]

			if response.Type == "error" {
				err := fmt.Errorf("%s", string(response.Body))
				c.handleError(err)
			}

			c.rw.RUnlock()

			if ok {
				handler(response)
			}

		}
	}
}

func (c *Client) handleError(err error) {
	if err == nil {
		return
	}

	if c.onErrorFn != nil {
		c.onErrorFn(err)
	}
}
