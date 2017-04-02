package pho

import "crypto/rand"

const MetadataSocketKey = "MetadataSocketKey"

// RandString generates a random string used to assigne Socket ID
func RandString(length int) (string, error) {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	data := make([]byte, length)
	if _, err := rand.Read(data); err != nil {
		return "", nil
	}

	for index, b := range data {
		data[index] = alphanum[b%byte(len(alphanum))]
	}

	return string(data), nil
}

// chain builds a http.Handler composed of an inline middleware stack and endpoint
// handler in the order they are passed.
func Chain(middlewares []MiddlewareFunc, endpoint Handler) Handler {
	// Return ahead of time if there aren't any middlewares for the chain
	if len(middlewares) == 0 {
		return endpoint
	}

	// Wrap the end handler with the middleware chain
	h := middlewares[len(middlewares)-1](endpoint)
	for i := len(middlewares) - 2; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}

// Copy websockets
func Copy(ws WebSockets) WebSockets {
	copy := WebSockets{}
	for id, socket := range ws {
		copy[id] = socket
	}
	return copy
}

// Sockets returns all availble sockets
func Sockets(w ResponseWriter) WebSockets {
	return w.Metadata()[MetadataSocketKey].(WebSockets)
}
