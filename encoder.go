package pho

import (
	"encoding/json"
	"io"
)

// Terminator defines the end of header and start of the body
const Terminator = 0x00

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
	if err := json.NewEncoder(e.writer).Encode(request); err != nil {
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
