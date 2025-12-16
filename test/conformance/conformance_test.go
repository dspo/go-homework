package conformance

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	_ "github.com/dspo/go-homework/conformance"
	"github.com/dspo/go-homework/sdk"
	"github.com/dspo/go-homework/test/framework"
)

func TestConformance(t *testing.T) {
	// do your init jobs, e.g. deploy service, database, prepare data

	var _ = sdk.NewSDK("http://app0:80") // init sdk with your app server address

	gomega.RegisterFailHandler(ginkgo.Fail)
	framework.GetFramework(t).DeployComponents()

	ginkgo.RunSpecs(t, "Conformance Suite")
}
