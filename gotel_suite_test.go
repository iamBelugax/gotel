package gotel_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGotel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gotel Suite")
}
