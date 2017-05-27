package render

import (
	"context"

	"github.com/svett/pho"
)

var (
	statusCtxKey = &contextKey{"Status"}
	verbCtxKey   = &contextKey{"Verb"}
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "pho render context value " + k.name
}

// Status sets status into request context.
func Status(r *pho.Request, status int) {
	*r = *r.WithContext(context.WithValue(r.Context(), statusCtxKey, status))
}

// Verb sets response verb into request context.
func Verb(r *pho.Request, responseType string) {
	*r = *r.WithContext(context.WithValue(r.Context(), verbCtxKey, responseType))
}
