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

	// Body is the request's body.
	//
	// For server requests the Request Body is always non-nil
	// but will return EOF immediately when no body is present.
	// The Server will close the request body. The ServeHTTP
	// Handler does not need to.
	Body json.RawMessage `json:"body"`
}
