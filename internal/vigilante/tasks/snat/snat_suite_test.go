package snat_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSnat(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SNAT task suite")
}
