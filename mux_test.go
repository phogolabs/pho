package pho_test

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"

	"github.com/svett/pho"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mux", func() {
	var (
		router *pho.Mux
		server *httptest.Server
	)

	BeforeEach(func() {
		router = pho.NewMux()
		server = httptest.NewServer(router)
	})

	AfterEach(func() {
		router.Close()
		server.Close()
	})

	It("handles requests successfully", func() {
		cnt := 0
		router.On("page_change", func(w pho.ResponseWriter, req *pho.Request) {
			defer GinkgoRecover()
			cnt++
			Expect(req.Verb).To(Equal("page_change"))

			body, err := ioutil.ReadAll(req.Body)
			Expect(err).To(BeNil())
			Expect(body).To(Equal([]byte("jack")))
		})

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Write("page_change", []byte("jack"))).To(BeNil())
		Eventually(func() int { return cnt }).Should(Equal(1))
	})
})
