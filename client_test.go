package pho_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/websocket"
	"github.com/svett/pho"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	It("sends data successfully", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
			Expect(err).To(BeNil())

			t, reader, err := conn.NextReader()
			Expect(err).To(BeNil())
			Expect(t).To(Equal(websocket.BinaryMessage))

			request := &pho.Request{}
			Expect(request.Unmarshal(reader)).To(Succeed())
			Expect(request.Verb).To(Equal("join"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).To(BeNil())
			Expect(body).To(Equal([]byte("jack")))

			Expect(conn.Close()).To(Succeed())
		}))

		defer server.Close()

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Send("join", []byte("jack"))).To(Succeed())
	})

	It("writes data successfully", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
			Expect(err).To(BeNil())

			t, reader, err := conn.NextReader()
			Expect(err).To(BeNil())
			Expect(t).To(Equal(websocket.BinaryMessage))

			request := &pho.Request{}
			Expect(request.Unmarshal(reader)).To(Succeed())
			Expect(request.Verb).To(Equal("join"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).To(BeNil())
			Expect(body).To(Equal([]byte("jack")))

			Expect(conn.Close()).To(Succeed())
		}))

		defer server.Close()

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Write("join", bytes.NewBufferString("jack"))).To(Succeed())
	})

	It("processes request successfully", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
			Expect(err).To(BeNil())

			t, reader, err := conn.NextReader()
			Expect(err).To(BeNil())
			Expect(t).To(Equal(websocket.BinaryMessage))

			request := &pho.Request{}
			Expect(request.Unmarshal(reader)).To(Succeed())
			Expect(request.Verb).To(Equal("join"))

			body, err := ioutil.ReadAll(request.Body)
			Expect(err).To(BeNil())
			Expect(body).To(Equal([]byte("jack")))

			Expect(conn.Close()).To(Succeed())
		}))

		defer server.Close()

		client, err := pho.Dial(fmt.Sprintf("ws://%s", server.Listener.Addr().String()), nil)
		Expect(err).To(BeNil())

		Expect(client.Do(&pho.Request{Verb: "join", Body: bytes.NewBufferString("jack")})).To(Succeed())
	})

})
