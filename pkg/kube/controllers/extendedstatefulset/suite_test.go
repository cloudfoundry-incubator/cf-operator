package extendedstatefulset_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestExtendedStatefulSet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ExtendedStatefulSet Suite")
}
