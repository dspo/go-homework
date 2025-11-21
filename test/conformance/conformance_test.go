package conformance

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	_ "github.com/dspo/go-homework/conformance"
)

func TestConformance(t *testing.T) {
	// do your init jobs, e.g. deploy service, database, prepare data

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Conformance Suite")
}
