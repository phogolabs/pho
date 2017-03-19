package pho

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// RequestBucket is JSON representation of the Request
type RequestBucket struct {
	// Verb provides the name of the request
	Verb string `json:"verb,omitempty"`

	// A Header represents the key-value pairs in an pho header.
	Header http.Header `json:"header,omitempty"`

	// Body is the request's body.
	//
	// For server requests the Request Body is always non-nil
	// but will return EOF immediately when no body is present.
	// The Server will close the request body. The ServeHTTP
	// Handler does not need to.
	Body []byte `json:"body,omitempty"`

	// RemoteAddr allows HTTP servers and other software to record
	// the network address that sent the request, usually for
	// logging. This field is not filled in by ReadRequest and
	// has no defined format. The HTTP server in this package
	// sets RemoteAddr to an "IP:port" address before invoking a
	// handler.
	// This field is ignored by the RPC client.
	RemoteAddr string `json:"remote_addr,omitempty"`

	// UserAgent returns the client's User-Agent, if sent in the request.
	UserAgent string `json:"user_agent,omitempty"`
}

// An Encoder writes Request to an output stream.
type Encoder struct {
	writer io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer: w,
	}
}

// Encode writes the JSON encoding of v to the stream,
// followed by a newline character.
func (e *Encoder) Encode(request *Request) error {
	bucket := &RequestBucket{
		Verb:       request.Verb,
		Header:     request.Header,
		RemoteAddr: request.RemoteAddr,
		UserAgent:  request.UserAgent,
	}

	if request.Body != nil {
		var err error
		bucket.Body, err = ioutil.ReadAll(request.Body)
		if err != nil {
			return err
		}
	}

	return json.NewEncoder(e.writer).Encode(bucket)
}
