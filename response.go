package pho

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
	Body []byte `json:"body"`
}
