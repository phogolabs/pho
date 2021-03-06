package pho

import "io"

// Writer comforms io.Writer interface and adopts ResponseWriter
type Writer struct {
	verb   string
	status int
	writer ResponseWriter
}

// NewWriter constructs a new writer
func NewWriter(verb string, status int, w ResponseWriter) io.Writer {
	return &Writer{verb: verb, status: status, writer: w}
}

// Write writes data
func (w *Writer) Write(data []byte) (int, error) {
	if err := w.writer.Write(w.verb, w.status, data); err != nil {
		return 0, err
	}
	return len(data), nil
}
