package pho

import "encoding/json"

// A Response represents an RPC response sent by a server
type Response struct {
	// Type provides the name of the request
	Type string `json:"type,omitempty"`

	// StatusCode of the response (ex. similar to HTTP)
	StatusCode int `json:"status_code,omitempty"`

	// A Header represents the key-value pairs in an pho header.
	Header Header `json:"header,omitempty"`

	// Payload is the response's payload.
	Payload json.RawMessage `json:"payload"`
}
