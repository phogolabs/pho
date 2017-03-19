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

var _ = Describe("Response", func() {
	var buffer *bytes.Buffer

	BeforeEach(func() {
		buffer = bytes.NewBufferString("")
	})

	Describe("Unmarshal", func() {
		BeforeEach(func() {
			buffer.Write([]byte(`{"verb":"my_verb","header":{"token":["my_token"]},"remote_addr":"localhost:9999","user_agent":"agent-0007"}`))
			buffer.WriteString("\n")
			buffer.WriteByte(0)
			buffer.WriteString("naked body")
		})

		It("decodes the response correctly", func() {
			response := &pho.Response{}
			Expect(response.Unmarshal(buffer)).To(Succeed())
			Expect(response.Header).To(HaveLen(1))
			Expect(response.Header).To(HaveKeyWithValue("token", []string{"my_token"}))

			data, err := ioutil.ReadAll(response.Body)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte("naked body")))
		})

		Context("when the body is empty", func() {
			BeforeEach(func() {
				buffer.Reset()
				buffer.Write([]byte(`{"verb":"my_verb","header":{"token":["my_token"]},"remote_addr":"localhost:9999","user_agent":"agent-0007"}`))
				buffer.WriteString("\n")
			})

			It("decodes the request correctly", func() {
				response := &pho.Response{}
				Expect(response.Unmarshal(buffer)).To(Succeed())
				Expect(response.Header).To(HaveLen(1))
				Expect(response.Header).To(HaveKeyWithValue("token", []string{"my_token"}))

				data, err := ioutil.ReadAll(response.Body)
				Expect(err).To(BeNil())
				Expect(data).To(Equal([]byte("")))
			})
		})

		Context("when the reader fails", func() {
			It("returns an error", func() {
				response := &pho.Response{}
				reader := new(fakes.FakeReader)
				reader.ReadReturns(0, fmt.Errorf("oh no!"))
				Expect(response.Unmarshal(reader)).To(MatchError("oh no!"))
			})
		})
	})
})
