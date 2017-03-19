package pho

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
)

// A Response represents an RPC response sent by a server
type Response struct {

	// StatusCode of the response (ex. similar to HTTP)
	StatusCode int `json:"status_code,omitempty"`

	// A Header represents the key-value pairs in an pho header.
	Header http.Header `json:"header,omitempty"`

	// Body is the request's body.
	//
	// For server requests the Request Body is always non-nil
	// but will return EOF immediately when no body is present.
	// The Server will close the request body. The ServeHTTP
	// Handler does not need to.
	Body io.Reader `json:"-"`
}

// Unmarshal the request from the reader
func (r *Response) Unmarshal(reader io.Reader) error {
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
