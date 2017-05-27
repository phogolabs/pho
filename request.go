package pho

import (
	"context"
	"encoding/json"
)

// Header information provided by the client
type Header map[string]string

// Terminator defines the end of header and start of the body
const Terminator = 0x00

// A Request represents an RPC request received by a server
// or to be sent by a client.
type Request struct {
	// Request context
	ctx context.Context

	// Type provides the name of the request
	Type string `json:"Type,omitempty"`

	// A Header represents the key-value pairs in an pho header.
	Header Header `json:"header,omitempty"`

	// Body is the request's body.
	Body json.RawMessage `json:"body"`
}

func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

// WithContext returns a shallow copy of r with its context changed
// to ctx. The provided ctx must be non-nil.
func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx
	return r2
}
