package pho_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPho(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pho Suite")
}
