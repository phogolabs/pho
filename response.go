package pho

import (
	"bufio"
	"encoding/json"
	"io"
)

// A Response represents an RPC response sent by a server
type Response struct {
	// Verb provides the name of the request
	Verb string `json:"verb,omitempty"`

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
	Body io.Reader `json:"-"`
}

// Marshal returns the JSON encoding of r.
func (r *Response) Marshal(writer io.Writer) error {
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
