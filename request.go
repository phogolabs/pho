package pho

import (
	"bufio"
	"encoding/json"
	"io"
)

type Header map[string]string

// Terminator defines the end of header and start of the body
const Terminator = 0x00

// A Request represents an RPC request received by a server
// or to be sent by a client.
type Request struct {
	// Verb provides the name of the request
	Verb string `json:"verb,omitempty"`

	// A Header represents the key-value pairs in an pho header.
	Header Header `json:"header,omitempty"`

	// Body is the request's body.
	//
	// For server requests the Request Body is always non-nil
	// but will return EOF immediately when no body is present.
	// The Server will close the request body. The ServeHTTP
	// Handler does not need to.
	Body io.Reader `json:"-"`

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

// Marshal returns the JSON encoding of r.
func (r *Request) Marshal(writer io.Writer) error {
	if err := json.NewEncoder(writer).Encode(r); err != nil {
		return err
	}

	if r.Body == nil {
		return nil
	}

	if _, err := writer.Write([]byte{Terminator}); err != nil {
		return err
	}

	if _, err := io.Copy(writer, r.Body); err != nil {
		return err
	}

	return nil
}

// Unmarshal the request from the reader
func (r *Request) Unmarshal(reader io.Reader) error {
	buffer := bufio.NewReader(reader)
	header, err := buffer.ReadBytes(Terminator)
	if err == nil {
		// remove the terminator from the header
		header = header[:len(header)-1]
	} else if err != io.EOF {
		return err
	}

	if err := json.Unmarshal(header, r); err != nil {
		return err
	}

	r.Body = buffer
	return nil
}
