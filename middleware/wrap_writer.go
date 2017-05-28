package middleware

// The original work was derived from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"github.com/svett/pho"
)

// WrapSocketWriter is a proxy around an http.ResponseWriter that allows you to hook
// into various parts of the response process.
type WrapSocketWriter interface {
	pho.SocketWriter
	// Status returns the HTTP status of the request, or 0 if one has not
	// yet been sent.
	Status() int
	// BytesWritten returns the total number of bytes sent to the client.
	BytesWritten() int
}

type writer struct {
	pho.SocketWriter
	code  int
	bytes int
}

// NewWrapSocketWriter creates a new wrap socket writer
func NewWrapSocketWriter(socket pho.SocketWriter) WrapSocketWriter {
	return &writer{
		SocketWriter: socket,
	}
}

// Write writes the response
func (w *writer) Write(verb string, code int, data []byte) error {
	w.code = code
	w.bytes = len(data)
	return w.SocketWriter.Write(verb, code, data)
}

// Status returns the status
func (b *writer) Status() int {
	return b.code
}

// BytesWritten returns the size of bytes written
func (b *writer) BytesWritten() int {
	return b.bytes
}
