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

var _ = Describe("Encoder", func() {
	var (
		encoder *pho.Encoder
		buffer  *bytes.Buffer
	)

	BeforeEach(func() {
		buffer = bytes.NewBufferString("")
		encoder = pho.NewEncoder(buffer)
	})

	It("encodes the request correctly", func() {
		request := &pho.Request{
			Verb: "my_verb",
			Body: bytes.NewBufferString("naked body"),
			Header: http.Header{
				"token": []string{"my_token"},
			},
			RemoteAddr: "localhost:9999",
			UserAgent:  "agent-0007",
		}

		Expect(encoder.Encode(request)).To(Succeed())
		Expect(buffer.String()).To(ContainSubstring(`{"verb":"my_verb","header":{"token":["my_token"]},"body":"bmFrZWQgYm9keQ==","remote_addr":"localhost:9999","user_agent":"agent-0007"}`))
	})

	Context("when the body is empty", func() {
		It("encodes the request correctly", func() {
			request := &pho.Request{
				Verb: "my_verb",
				Header: http.Header{
					"token": []string{"my_token"},
				},
				RemoteAddr: "localhost:9999",
				UserAgent:  "agent-0007",
			}

			Expect(encoder.Encode(request)).To(Succeed())
			Expect(buffer.String()).To(ContainSubstring(`{"verb":"my_verb","header":{"token":["my_token"]},"remote_addr":"localhost:9999","user_agent":"agent-0007"}`))
		})
	})

	Context("when the writer fails", func() {
		It("returns an error", func() {
			writer := new(fakes.FakeWriter)
			writer.WriteReturns(0, fmt.Errorf("oh no!"))
			Expect(pho.NewEncoder(writer).Encode(&pho.Request{})).To(MatchError("oh no!"))
		})
	})
})
