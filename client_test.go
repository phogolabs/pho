package pho_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/websocket"
	"github.com/svett/pho"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	It("sends data successfully", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()

			conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
			Expect(err).To(BeNil())

			t, reader, err := conn.NextReader()
			Expect(err).To(BeNil())
			Expect(t).To(Equal(websocket.BinaryMessage))

			request := &pho.Request{}
			Expect(json.NewDecoder(reader).Decode(request)).To(Succeed())
			Expect(request.Type).To(Equal("join"))
			Expect(string(request.Body)).To(Equal(`"jack"`))

			Expect(conn.Close()).To(Succeed())
		}))

		defer server.Close()

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Write("join", []byte(`"jack"`))).To(Succeed())
	})

	It("processes request successfully", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()

			conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
			Expect(err).To(BeNil())

			t, reader, err := conn.NextReader()
			Expect(err).To(BeNil())
			Expect(t).To(Equal(websocket.BinaryMessage))

			request := &pho.Request{}
			Expect(json.NewDecoder(reader).Decode(request)).To(Succeed())
			Expect(request.Type).To(Equal("join"))
			Expect(string(request.Body)).To(Equal(`"jack"`))

			Expect(conn.Close()).To(Succeed())
		}))

		defer server.Close()

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Do(&pho.Request{Type: "join", Body: []byte(`"jack"`)})).To(Succeed())
	})

	Context("when the verb is missing", func() {
		It("return an error", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := websocket.Upgrade(w, r, nil, 1024, 1024)
				Expect(err).To(BeNil())
			}))

			defer server.Close()

			client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
			Expect(err).To(BeNil())

			Expect(client.Do(&pho.Request{Type: "", Body: []byte(`"jack"`)})).To(MatchError("The Request does not have verb"))
		})
	})

	It("receives server responses successfully", func() {
		defer GinkgoRecover()
		var conn *websocket.Conn

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			var err error
			conn, err = websocket.Upgrade(w, r, nil, 1024, 1024)
			Expect(conn.SetWriteDeadline(time.Now().Add(pho.WriteDeadline))).To(Succeed())
			Expect(err).To(BeNil())
		}))

		defer server.Close()

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())
		defer client.Close()

		cnt := 0

		client.On("ping", func(resp *pho.Response) {
			defer GinkgoRecover()
			cnt++

			Expect(resp.Type).To(Equal("ping"))
			Expect(resp.Header).To(HaveKeyWithValue("token", "my_token"))
			Expect(string(resp.Payload)).To(Equal(`"naked body"`))
		})

		cntOnResponse := 0
		client.OnResponse(func(resp *pho.Response) {
			defer GinkgoRecover()
			cntOnResponse++
			Expect(resp.Type).To(Equal("ping"))
			Expect(resp.Header).To(HaveKeyWithValue("token", "my_token"))
			Expect(string(resp.Payload)).To(Equal(`"naked body"`))
		})

		body := []byte(`"naked body"`)
		response := &pho.Response{
			Type:    "ping",
			Payload: body,
			Header: pho.Header{
				"token": "my_token",
			},
		}

		w, err := conn.NextWriter(websocket.BinaryMessage)
		Expect(err).To(BeNil())
		Expect(json.NewEncoder(w).Encode(response)).To(Succeed())
		Expect(w.Close()).To(Succeed())

		response.Payload = []byte(`"naked body"`)

		w, err = conn.NextWriter(websocket.BinaryMessage)
		Expect(err).To(BeNil())
		Expect(json.NewEncoder(w).Encode(response)).To(Succeed())
		Expect(w.Close()).To(Succeed())

		Eventually(func() int { return cnt }).Should(Equal(2))
		Eventually(func() int { return cntOnResponse }).Should(Equal(2))
	})

	Context("when cannot connect to the server", func() {
		It("returns the error", func() {
			client, err := pho.Dial("ws://test.com", nil)
			Expect(client).To(BeNil())
			Expect(err).To(MatchError("websocket: bad handshake"))
		})
	})

	Context("when the client is disconnected", func() {
		Context("when perform request", func() {
			It("returns the error", func() {
				defer GinkgoRecover()
				var conn *websocket.Conn

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					conn, err = websocket.Upgrade(w, r, nil, 1024, 1024)
					Expect(conn.SetWriteDeadline(time.Now().Add(pho.WriteDeadline))).To(Succeed())
					Expect(err).To(BeNil())
				}))

				client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
				Expect(err).To(BeNil())
				client.Close()
				server.Close()

				Eventually(func() error { return client.Write("message", nil) }).Should(MatchError("websocket: close sent"))
			})
		})
	})
})
