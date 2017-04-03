package pho_test

import (
	"errors"
	"io"

	"github.com/svett/fakes"
	"github.com/svett/pho"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Writer", func() {
	var (
		response *fakes.FakeResponseWriter
		writer   io.Writer
	)

	BeforeEach(func() {
		response = new(fakes.FakeResponseWriter)
		writer = pho.NewWriter("test", response)
	})

	It("writes data successfully", func() {
		_, err := writer.Write([]byte("hi"))
		Expect(err).To(BeNil())
		Expect(response.WriteCallCount()).To(Equal(1))
		verb, data := response.WriteArgsForCall(0)
		Expect(verb).To(Equal("test"))
		Expect(data).To(Equal([]byte("hi")))
	})

	Context("when the underlying writer fails", func() {
		It("returns the error", func() {
			response.WriteReturns(errors.New("oh no!"))

			len, err := writer.Write([]byte("hi"))
			Expect(err).To(MatchError("oh no!"))
			Expect(len).To(Equal(0))
		})
	})

})
