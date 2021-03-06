package pho_test

import (
	"errors"
	"io"
	"net/http"

	"github.com/svett/pho"
	"github.com/svett/pho/fakes"

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
		writer = pho.NewWriter("test", http.StatusOK, response)
	})

	It("writes data successfully", func() {
		_, err := writer.Write([]byte("hi"))
		Expect(err).To(BeNil())
		Expect(response.WriteCallCount()).To(Equal(1))
		verb, status, data := response.WriteArgsForCall(0)
		Expect(status).To(Equal(http.StatusOK))
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
