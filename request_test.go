package pho_test

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/svett/fakes"
	"github.com/svett/pho"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Request", func() {
	var buffer *bytes.Buffer

	BeforeEach(func() {
		buffer = bytes.NewBufferString("")
	})

	Describe("Marshal", func() {
		var request *pho.Request

		BeforeEach(func() {
			request = &pho.Request{
				Verb: "my_verb",
				Body: bytes.NewBufferString("naked body"),
				Header: pho.Header{
					"token": "my_token",
				},
				RemoteAddr: "localhost:9999",
				UserAgent:  "agent-0007",
			}
		})

		It("encodes the request correctly", func() {
			Expect(request.Marshal(buffer)).To(Succeed())
			Expect(buffer.String()).To(ContainSubstring(`{"verb":"my_verb","header":{"token":"my_token"},"remote_addr":"localhost:9999","user_agent":"agent-0007"}` + "\n\x00" + "naked body"))
		})

		Context("when the body is empty", func() {
			BeforeEach(func() {
				request.Body = nil
			})

			It("encodes the request correctly", func() {
				Expect(request.Marshal(buffer)).To(Succeed())
				Expect(buffer.String()).To(ContainSubstring(`{"verb":"my_verb","header":{"token":"my_token"},"remote_addr":"localhost:9999","user_agent":"agent-0007"}`))
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

	Describe("Unmarshal", func() {
		BeforeEach(func() {
			buffer.Write([]byte(`{"verb":"my_verb","header":{"token":"my_token"},"remote_addr":"localhost:9999","user_agent":"agent-0007"}`))
			buffer.WriteString("\n")
			buffer.WriteByte(0)
			buffer.WriteString("naked body")
		})

		It("decodes the request correctly", func() {
			request := &pho.Request{}
			Expect(request.Unmarshal(buffer)).To(Succeed())
			Expect(request.Verb).To(Equal("my_verb"))
			Expect(request.Header).To(HaveLen(1))
			Expect(request.Header).To(HaveKeyWithValue("token", "my_token"))
			Expect(request.RemoteAddr).To(Equal("localhost:9999"))
			Expect(request.UserAgent).To(Equal("agent-0007"))

			data, err := ioutil.ReadAll(request.Body)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte("naked body")))
		})

		Context("when the body is empty", func() {
			BeforeEach(func() {
				buffer.Reset()
				buffer.Write([]byte(`{"verb":"my_verb","header":{"token":"my_token"},"remote_addr":"localhost:9999","user_agent":"agent-0007"}`))
				buffer.WriteString("\n")
			})

			It("decodes the request correctly", func() {
				request := &pho.Request{}
				Expect(request.Unmarshal(buffer)).To(Succeed())
				Expect(request.Verb).To(Equal("my_verb"))
				Expect(request.Header).To(HaveLen(1))
				Expect(request.Header).To(HaveKeyWithValue("token", "my_token"))
				Expect(request.RemoteAddr).To(Equal("localhost:9999"))
				Expect(request.UserAgent).To(Equal("agent-0007"))

				data, err := ioutil.ReadAll(request.Body)
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("")))
			})
		})

		Context("when the reader fails", func() {
			It("returns an error", func() {
				request := &pho.Request{}
				reader := new(fakes.FakeReader)
				reader.ReadReturns(0, fmt.Errorf("oh no!"))
				Expect(request.Unmarshal(reader)).To(MatchError("oh no!"))
			})
		})
	})
})
