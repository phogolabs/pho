package middleware

// Ported from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"bytes"
	"context"
	"log"
	"strings"
	"time"

	"github.com/svett/pho"
)

var WrapResponseWriterCtxKey = &contextKey{"ResponseWriter"}

// Logger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return. When standard output is a TTY, Logger will
// print in color, otherwise it will print in black and white.
//
// Logger prints a request ID if one is provided.
func Logger(next pho.Handler) pho.Handler {
	fn := func(w pho.SocketWriter, r *pho.Request) {
		reqID := GetReqID(r.Context())
		prefix := requestPrefix(reqID, w, r)

		// Create a new WrapResponseWriter and set it on the context for other
		// handlers to make us of the status code, or other features of the wrapWriter.
		ww := NewWrapSocketWriter(w)
		r = r.WithContext(context.WithValue(r.Context(), WrapResponseWriterCtxKey, ww))

		t1 := time.Now()
		defer func() {
			t2 := time.Now()
			printRequest(prefix, reqID, ww, t2.Sub(t1))
		}()

		next.ServeRPC(ww, r)
	}

	return pho.HandlerFunc(fn)
}

func requestPrefix(reqID string, w pho.SocketWriter, r *pho.Request) *bytes.Buffer {
	buf := &bytes.Buffer{}

	if reqID != "" {
		cW(buf, nYellow, "[%s] ", reqID)
	}
	cW(buf, nCyan, "\"")
	cW(buf, bMagenta, "%s ", strings.ToUpper(r.Type))

	if w.TLS() == nil {
		cW(buf, nCyan, "ws://%s\" ", w.EndpointAddr())
	} else {
		cW(buf, nCyan, "wss://%s\" ", w.EndpointAddr())
	}

	buf.WriteString("from ")
	buf.WriteString(w.RemoteAddr())
	buf.WriteString(" - ")

	return buf
}

func printRequest(buf *bytes.Buffer, reqID string, w WrapSocketWriter, dt time.Duration) {
	status := w.Status()
	switch {
	case status < 200:
		cW(buf, bBlue, "%03d", status)
	case status < 300:
		cW(buf, bGreen, "%03d", status)
	case status < 400:
		cW(buf, bCyan, "%03d", status)
	case status < 500:
		cW(buf, bYellow, "%03d", status)
	default:
		cW(buf, bRed, "%03d", status)
	}

	cW(buf, bBlue, " %dB", w.BytesWritten())

	buf.WriteString(" in ")
	if dt < 500*time.Millisecond {
		cW(buf, nGreen, "%s", dt)
	} else if dt < 5*time.Second {
		cW(buf, nYellow, "%s", dt)
	} else {
		cW(buf, nRed, "%s", dt)
	}

	log.Print(buf.String())
}
