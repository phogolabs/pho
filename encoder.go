package pho

import (
	"encoding/json"
	"io"
	"net/http"
)

// Terminator defines the end of header and start of the body
const Terminator = 0x00

// RequestHeader is JSON representation of the Request
type RequestHeader struct {
	// Verb provides the name of the request
	Verb string `json:"verb,omitempty"`

	// A Header represents the key-value pairs in an pho header.
	Header http.Header `json:"header,omitempty"`

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
	header := &RequestHeader{
		Verb:       request.Verb,
		Header:     request.Header,
		RemoteAddr: request.RemoteAddr,
		UserAgent:  request.UserAgent,
	}

	if err := json.NewEncoder(e.writer).Encode(header); err != nil {
		return err
	}

	if _, err := e.writer.Write([]byte{Terminator}); err != nil {
		return err
	}

	if request.Body == nil {
		return nil
	}

	if _, err := io.Copy(e.writer, request.Body); err != nil {
		return err
	}

	return nil
}
