package pho

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type MuxWriter struct {
	Conn *websocket.Conn
}

// Write a reponse
func (c *MuxWriter) Write(verb string, data []byte) error {
	response := &Response{
		Verb:       verb,
		StatusCode: http.StatusOK,
		Body:       bytes.NewBuffer(data),
	}

	return c.write(response)
}

// WriteError writes an errors with specified code
func (c *MuxWriter) WriteError(err error, code int) error {
	response := &Response{
		Verb:       "error",
		StatusCode: code,
		Body:       strings.NewReader(err.Error()),
	}

	return c.write(response)
}

func (c *MuxWriter) write(response *Response) error {
	writer, err := c.Conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	defer writer.Close()

	if err = response.Marshal(writer); err != nil {
		return err
	}
	return nil
}
