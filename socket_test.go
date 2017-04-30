package pho_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/svett/pho"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sockets", func() {
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

	It("returns the list of all sockets", func() {
		cnt := 0
		router.On("message", func(w pho.SocketWriter, req *pho.Request) {
			defer GinkgoRecover()
			cnt++
			sockets := pho.Sockets(w)
			Expect(sockets).To(HaveLen(1))
			Expect(sockets).To(HaveKeyWithValue(w.SocketID(), w))
		})

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Write("message", []byte(`""`))).To(BeNil())
		Eventually(func() int { return cnt }).Should(Equal(1))
	})

	Context("when writes an error", func() {
		It("raises OnError function", func() {
			router.On("message", func(w pho.SocketWriter, req *pho.Request) {
				defer GinkgoRecover()
				Expect(w.WriteError(fmt.Errorf("This is an error"), http.StatusBadRequest)).To(Succeed())
			})

			cnt := 0
			router.OnError(func(err error) {
				defer GinkgoRecover()
				cnt++
				Expect(err).To(MatchError("This is an error"))
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())

			client.OnError(func(err error) {
				defer GinkgoRecover()
				cnt++
				Expect(err).To(MatchError("This is an error"))
			})

			Expect(client.Write("message", []byte(`""`))).To(BeNil())
			Eventually(func() int { return cnt }).Should(Equal(2))
		})
	})

})
