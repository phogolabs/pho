package pho_test

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/svett/fakes"
	"github.com/svett/pho"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Request", func() {
	Describe("Marshal", func() {
		var (
			request *pho.Request
			buffer  *bytes.Buffer
		)

		BeforeEach(func() {
			buffer = bytes.NewBufferString("")

			request = &pho.Request{
				Verb: "my_verb",
				Body: bytes.NewBufferString("naked body"),
				Header: http.Header{
					"token": []string{"my_token"},
				},
				RemoteAddr: "localhost:9999",
				UserAgent:  "agent-0007",
			}
		})

		It("encodes the request correctly", func() {
			Expect(request.Marshal(buffer)).To(Succeed())
			Expect(buffer.String()).To(ContainSubstring(`{"verb":"my_verb","header":{"token":["my_token"]},"remote_addr":"localhost:9999","user_agent":"agent-0007"}` + "\n\x00" + "naked body"))
		})

		Context("when the body is empty", func() {
			BeforeEach(func() {
				request.Body = nil
			})

			It("encodes the request correctly", func() {
				Expect(request.Marshal(buffer)).To(Succeed())
				Expect(buffer.String()).To(ContainSubstring(`{"verb":"my_verb","header":{"token":["my_token"]},"remote_addr":"localhost:9999","user_agent":"agent-0007"}`))
			})
		})

		Context("when the writer fails", func() {
			It("returns an error", func() {
				writer := new(fakes.FakeWriter)
				writer.WriteReturns(0, fmt.Errorf("oh no!"))
				Expect(request.Marshal(writer)).To(MatchError("oh no!"))
			})
		})
	})
})
