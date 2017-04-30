package pho_test

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/svett/pho"
)

var _ = Describe("Request", func() {
	It("unmarshal request from JSON successfully", func() {
		reader := bytes.NewBufferString(`{"type":"document:fetch","body":{"page_index":0,"page_size":0}}`)
		request := &pho.Request{}
		Expect(json.NewDecoder(reader).Decode(request)).To(Succeed())
		Expect(request.Type).To(Equal("document:fetch"))
		Expect(string(request.Body)).To(Equal(`{"page_index":0,"page_size":0}`))
	})
})
