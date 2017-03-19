package pho

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

// A Client is an HTTP client. Its zero value (DefaultClient) is a
// usable client that uses DefaultTransport.
type Client struct {
	conn *websocket.Conn
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

	return &Client{
		conn: conn,
	}, nil
}

// Emit emits data to the server
func (c *Client) Emit(verb string, body []byte) error {
	return c.Do(&Request{
		Verb: verb,
		Body: bytes.NewBuffer(body),
	})
}

// Copy copy data to the server
func (c *Client) Copy(verb string, reader io.Reader) error {
	return c.Do(&Request{
		Verb: verb,
		Body: reader,
	})
}

// Do sends an RPC request and returns an RPC response
func (c *Client) Do(req *Request) error {
	w, err := c.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	return req.Marshal(w)
}

// Close closes the connection to the server
func (c *Client) Close() error {
	return c.conn.Close()
}
