package render

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/svett/pho"
)

// Respond handles streaming JSON and XML responses, automatically setting the
// Content-Type based on request headers. It will default to a JSON response.
func Respond(w pho.SocketWriter, r *pho.Request, v interface{}) error {
	verb, ok := r.Context().Value(verbCtxKey).(string)
	if !ok {
		verb = r.Type
	}

	status, ok := r.Context().Value(statusCtxKey).(int)
	if !ok {
		status = http.StatusOK
	}

	buffer := &bytes.Buffer{}
	enc := json.NewEncoder(buffer)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		return err
	}

	return w.Write(verb, status, buffer.Bytes())
}
