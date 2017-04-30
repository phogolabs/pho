package pho

import "encoding/json"

// Header information provided by the client
type Header map[string]string

// Terminator defines the end of header and start of the body
const Terminator = 0x00

// A Request represents an RPC request received by a server
// or to be sent by a client.
type Request struct {
	// Type provides the name of the request
	Type string `json:"Type,omitempty"`

	// A Header represents the key-value pairs in an pho header.
	Header Header `json:"header,omitempty"`

	// Body is the request's body.
	Body json.RawMessage `json:"body"`
}
