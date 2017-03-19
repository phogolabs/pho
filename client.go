package pho

import (
	"bytes"
	"fmt"
	"io"
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

// ResponseFunc is a callback function that occurs when response is received
type ResponseFunc func(r *Response)

// A Client is an RPC client.
type Client struct {
	handlers map[string]ResponseFunc
	rw       *sync.RWMutex
	stopChan chan struct{}
	conn     *websocket.Conn
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
		handlers: map[string]ResponseFunc{},
		rw:       &sync.RWMutex{},
		stopChan: make(chan struct{}),
		conn:     conn,
	}

	go client.run()

	return client, nil
}

// Write writes data to the server
func (c *Client) Write(verb string, body []byte) error {
	return c.Do(&Request{
		Verb: verb,
		Body: bytes.NewBuffer(body),
	})
}

// WriteFrom writes data to the server
func (c *Client) WriteFrom(verb string, reader io.Reader) error {
	return c.Do(&Request{
		Verb: verb,
		Body: reader,
	})
}

// Do sends an RPC request and returns an RPC response
func (c *Client) Do(req *Request) error {
	if req.Verb == "" {
		return fmt.Errorf("The Request does not have verb")
	}

	w, err := c.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	defer w.Close()
	return req.Marshal(w)
}

// On register callback function called when response with provided verb occurs
func (c *Client) On(verb string, fn ResponseFunc) {
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

			response := &Response{}
			if err := response.Unmarshal(reader); err != nil {
				continue
			}

			c.rw.RLock()
			handler, ok := c.handlers[response.Verb]
			c.rw.RUnlock()

			if ok {
				handler(response)
			}
		}
	}
}
