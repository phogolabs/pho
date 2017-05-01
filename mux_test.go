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
		router.On("page_change", func(w pho.SocketWriter, req *pho.Request) {
			defer GinkgoRecover()
			cnt++
			Expect(req.Type).To(Equal("page_change"))
			Expect(string(req.Body)).To(Equal(`"jack"`))

			Expect(w.RemoteAddr()).To(ContainSubstring("127.0.0.1"))
			Expect(w.UserAgent()).To(Equal("Go-http-client/1.1"))
		})

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Write("page_change", []byte(`"jack"`))).To(BeNil())
		Eventually(func() int { return cnt }).Should(Equal(1))
	})

	It("handles error requests successfully", func() {
		cnt := 0
		router.OnError(func(err error) {
			defer GinkgoRecover()
			if cnt == 0 {
				Expect(err).To(MatchError(`"oh no"`))
			} else {
				Expect(err).To(MatchError(fmt.Errorf("The route %q does not exist", "error")))
			}
			cnt++
		})

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Write("error", []byte(`"oh no"`))).To(BeNil())
		Eventually(func() int { return cnt }).Should(Equal(2))
	})

	It("writes response successfully", func() {
		router.OnConnect(func(w pho.SocketWriter, req *http.Request) {
			Expect(w.Write("message", []byte(`"naked body"`))).To(Succeed())
		})

		cnt := 0
		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		client.On("message", func(resp *pho.Response) {
			defer GinkgoRecover()
			cnt++

			Expect(resp.Type).To(Equal("message"))
			Expect(string(resp.Payload)).To(Equal(`"naked body"`))
		})

		Eventually(func() int { return cnt }).Should(Equal(1))
	})

	It("calls OnConnect function for each new connection", func() {
		cnt := 0
		router.OnConnect(func(w pho.SocketWriter, req *http.Request) {
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
			router.OnConnect(func(w pho.SocketWriter, req *http.Request) {
				w.Metadata()["user"] = "root"
			})
		})

		It("provides the metadata for each request", func() {
			cnt := 0
			router.On("message", func(w pho.SocketWriter, req *pho.Request) {
				defer GinkgoRecover()
				cnt++
				Expect(w.Metadata()).To(HaveKeyWithValue("user", "root"))
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			defer client.Close()

			Expect(client.Write("message", []byte(`""`))).To(BeNil())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when two client are connect", func() {
		It("they can talk to each other", func() {
			cnt := 0
			clientA, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())

			router.On("message", func(w pho.SocketWriter, r *pho.Request) {
				defer GinkgoRecover()

				for _, c := range pho.Sockets(w) {
					Expect(c.Write("message", r.Body)).To(Succeed())
				}
			})

			clientB, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			Expect(clientB.Write("message", []byte(`"Hi from B"`))).To(Succeed())

			clientA.On("message", func(resp *pho.Response) {
				defer GinkgoRecover()
				cnt++

				Expect(resp.Type).To(Equal("message"))
				Expect(string(resp.Payload)).To(Equal(`"Hi from B"`))
			})

			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when error occurs", func() {
		It("returns the error via error channel", func() {
			cnt := 0

			router.OnConnect(func(w pho.SocketWriter, req *http.Request) {
				defer GinkgoRecover()
				Expect(w.WriteError(fmt.Errorf("oh no!"), http.StatusForbidden)).To(Succeed())
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())

			client.On("error", func(resp *pho.Response) {
				defer GinkgoRecover()
				cnt++

				Expect(string(resp.Payload)).To(Equal(`{"error":"oh no!"}`))
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

				Expect(string(resp.Payload)).To(Equal(`{"error":"The route \"message\" does not exist"}`))
			})

			Expect(client.Write("message", []byte(`"Hi"`))).To(Succeed())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when a router is mount", func() {
		It("delegates all client requests to it", func() {
			cnt := 0
			subrouter := pho.NewRouter()
			subrouter.On("insert", func(w pho.SocketWriter, r *pho.Request) {
				defer GinkgoRecover()
				cnt++

				Expect(r.Type).To(Equal("insert"))
				Expect(string(r.Body)).To(Equal(`"Hi"`))
			})

			router.Mount("message", subrouter)

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			Expect(client.Write("message:insert", []byte(`"Hi"`))).To(Succeed())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when a sub route is defined", func() {
		It("delegates all client requests to it", func() {
			cnt := 0

			router.Route("message", func(r pho.Router) {
				r.On("insert", func(w pho.SocketWriter, r *pho.Request) {
					defer GinkgoRecover()
					cnt++

					Expect(r.Type).To(Equal("insert"))
					Expect(string(r.Body)).To(Equal(`"Hi"`))
				})
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			Expect(client.Write("message:insert", []byte(`"Hi"`))).To(Succeed())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when middleware is registered", func() {
		It("calls it before handling the request", func() {
			cnt := 0
			router.On("message", func(w pho.SocketWriter, req *pho.Request) {})

			router.Use(func(h pho.Handler) pho.Handler {
				return pho.HandlerFunc(func(w pho.SocketWriter, r *pho.Request) {
					defer GinkgoRecover()
					h.ServeRPC(w, r)
					cnt++
				})
			})

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			defer client.Close()

			Expect(client.Write("message", []byte(`""`))).To(BeNil())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when client is disconnected by the server", func() {
		It("removes the client from the list of all sockets", func() {
			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())
			defer client.Close()

			cnt := 0

			router.OnDisconnect(func(w pho.SocketWriter) {
				defer GinkgoRecover()
				cnt++
				Expect(pho.Sockets(w)).To(BeEmpty())
			})

			router.Close()
			Expect(client.Write("hello", []byte(`"world"`))).To(Succeed())
			Eventually(func() int { return cnt }).Should(Equal(1))
		})
	})

	Context("when client disconnect from the server", func() {
		It("removes the client from the list of all sockets", func() {
			cnt := 0

			router.OnDisconnect(func(w pho.SocketWriter) {
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
