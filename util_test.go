package pho_test

import (
	"github.com/svett/pho"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {
	It("chains functions in correct order", func() {
		middlewareCnt := 0
		handlerCnt := 0

		handler := pho.Chain([]pho.MiddlewareFunc{
			func(h pho.Handler) pho.Handler {
				middlewareCnt++
				Expect(handlerCnt).To(Equal(0))
				return h
			},
			func(h pho.Handler) pho.Handler {
				middlewareCnt++
				Expect(handlerCnt).To(Equal(0))
				Expect(middlewareCnt).To(Equal(1))
				return h
			},
		}, pho.HandlerFunc(func(w pho.SocketWriter, r *pho.Request) {
			handlerCnt++
			Expect(middlewareCnt).To(Equal(2))
		}))

		handler.ServeRPC(nil, nil)
		Expect(handlerCnt).To(Equal(1))
		Expect(middlewareCnt).To(Equal(2))
	})
})
