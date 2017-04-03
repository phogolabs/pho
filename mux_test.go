package pho_test

import (
	"fmt"
	"net/http"
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
			Expect(req.Body).To(Equal([]byte("jack")))

			Expect(w.RemoteAddr()).To(ContainSubstring("127.0.0.1"))
			Expect(w.UserAgent()).To(Equal("Go-http-client/1.1"))
		})

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Write("page_change", []byte("jack"))).To(BeNil())
		Eventually(func() int { return cnt }).Should(Equal(1))
	})

	It("writes response successfully", func() {
		router.OnConnect(func(w pho.ResponseWriter, req *http.Request) {
			Expect(w.Write("message", []byte("naked body"))).To(Succeed())
		})

		cnt := 0
		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		client.On("message", func(resp *pho.Response) {
			defer GinkgoRecover()
			cnt++

			Expect(resp.Verb).To(Equal("message"))
			Expect(resp.Body).To(Equal([]byte("naked body")))
		})

		Eventually(func() int { return cnt }).Should(Equal(1))
	})

	It("calls OnConnect function for each new connection", func() {
		cnt := 0
		router.OnConnect(func(w pho.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			cnt++

			Expect(w.RemoteAddr()).To(ContainSubstring("127.0.0.1"))
			Expect(w.UserAgent()).To(Equal("Go-http-client/1.1"))
		})

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())
		defer client.Close()

		Eventually(func() int { return cnt }).Should(Equal(1))
	})

	Context("when the metadata is set", func() {
		BeforeEach(func() {
			router.OnConnect(func(w pho.ResponseWriter, req *http.Request) {
				w.Metadata()["user"] = "root"
			})
		})

		It("provides the metadata for each request", func() {
			cnt := 0
			router.On("message", func(w pho.ResponseWriter, req *pho.Request) {
				defer GinkgoRecover()
				cnt++
				Expect(w.Metadata()).To(HaveKeyWithValue("user", "root"))
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			defer client.Close()

			Expect(client.Write("message", []byte(""))).To(BeNil())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when two client are connect", func() {
		It("they can talk to each other", func() {
			cnt := 0
			clientA, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())

			router.On("message", func(w pho.ResponseWriter, r *pho.Request) {
				defer GinkgoRecover()

				for _, c := range pho.Sockets(w) {
					Expect(c.Write("message", r.Body)).To(Succeed())
				}
			})

			clientB, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			Expect(clientB.Write("message", []byte("Hi from B"))).To(Succeed())

			clientA.On("message", func(resp *pho.Response) {
				defer GinkgoRecover()
				cnt++

				Expect(resp.Verb).To(Equal("message"))
				Expect(resp.Body).To(Equal([]byte("Hi from B")))
			})

			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when error occurs", func() {
		It("returns the error via error channel", func() {
			cnt := 0

			router.OnConnect(func(w pho.ResponseWriter, req *http.Request) {
				defer GinkgoRecover()
				Expect(w.WriteError(fmt.Errorf("oh no!"), http.StatusForbidden)).To(Succeed())
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())

			client.On("error", func(resp *pho.Response) {
				defer GinkgoRecover()
				cnt++

				Expect(resp.Body).To(Equal([]byte("oh no!")))
			})

			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when the path is not found", func() {
		It("returns an error", func() {
			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			cnt := 0

			client.On("error", func(resp *pho.Response) {
				defer GinkgoRecover()
				cnt++

				Expect(resp.Body).To(Equal([]byte(`The route "message" does not exist`)))
			})

			Expect(client.Write("message", []byte("Hi"))).To(Succeed())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when a router is mount", func() {
		It("delegates all client requests to it", func() {
			cnt := 0
			subrouter := pho.NewRouter()
			subrouter.On("insert", func(w pho.ResponseWriter, r *pho.Request) {
				defer GinkgoRecover()
				cnt++

				Expect(r.Verb).To(Equal("insert"))
				Expect(r.Body).To(Equal([]byte("Hi")))
			})

			router.Mount("message", subrouter)

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			Expect(client.Write("message:insert", []byte("Hi"))).To(Succeed())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when a sub route is defined", func() {
		It("delegates all client requests to it", func() {
			cnt := 0

			router.Route("message", func(r pho.Router) {
				r.On("insert", func(w pho.ResponseWriter, r *pho.Request) {
					defer GinkgoRecover()
					cnt++

					Expect(r.Verb).To(Equal("insert"))
					Expect(r.Body).To(Equal([]byte("Hi")))
				})
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			Expect(client.Write("message:insert", []byte("Hi"))).To(Succeed())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when middleware is registered", func() {
		It("calls it before handling the request", func() {
			cnt := 0
			router.On("message", func(w pho.ResponseWriter, req *pho.Request) {})

			router.Use(func(h pho.Handler) pho.Handler {
				return pho.HandlerFunc(func(w pho.ResponseWriter, r *pho.Request) {
					defer GinkgoRecover()
					h.ServeRPC(w, r)
					cnt++
				})
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			defer client.Close()

			Expect(client.Write("message", []byte(""))).To(BeNil())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when client is disconnected by the server", func() {
		It("removes the client from the list of all sockets", func() {
			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			defer client.Close()

			cnt := 0

			router.OnDisconnect(func(w pho.ResponseWriter) {
				defer GinkgoRecover()
				cnt++
				Expect(pho.Sockets(w)).To(BeEmpty())
			})

			router.Close()
			Expect(client.Write("hello", []byte("world"))).To(Succeed())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when client disconnect from the server", func() {
		It("removes the client from the list of all sockets", func() {
			cnt := 0

			router.OnDisconnect(func(w pho.ResponseWriter) {
				defer GinkgoRecover()
				cnt++
				Expect(pho.Sockets(w)).To(BeEmpty())
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			client.Close()

			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})
})