package conformance

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _exampleCase = ginkgo.Describe("This is an example case", func() {
	ginkgo.Context("This is an example Context", func() {
		ginkgo.It("This is an example node", func() {
			ginkgo.By("Assert 1 + 1 == 2")
			gomega.Ω(1 + 1).To(gomega.Equal(2))
		})

		ginkgo.It("This is an example which should fail", func() {
			ginkgo.By("Assert true is false")
			gomega.Ω(true).To(gomega.BeFalse())
		})
	})
})

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	// todo: deploy database

	// todo: init database

	// todo: deploy your app

	ginkgo.RunSpecs(t, "E2E")
}
