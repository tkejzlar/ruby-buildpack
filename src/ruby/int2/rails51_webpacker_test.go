package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRails51Webpacker(t *testing.T) {
	t.Parallel()
	spec.Run(t, "pushing a rails51 webpacker app", func(t *testing.T, when spec.G, it spec.S) {
		var app cfapi.App
		var err error
		var g *GomegaWithT
		var Expect func(actual interface{}, extra ...interface{}) GomegaAssertion
		var Eventually func(actual interface{}, intervals ...interface{}) GomegaAsyncAssertion
		it.Before(func() {
			g = NewGomegaWithT(t)
			Expect = g.Expect
			Eventually = g.Eventually
		})
		it.After(func() {
			if app != nil {
				app.Destroy()
			}
		})
		it.Before(func() {
			app, err = cluster.NewApp(bpDir, "rails51_webpacker")
			Expect(err).ToNot(HaveOccurred())
		})

		it("compiles assets with webpacker", func() {
			Expect(app.PushAndConfirm()).To(Succeed())
			Expect(app.Log()).To(ContainSubstring("Webpacker is installed"))
			Expect(app.Log()).To(ContainSubstring("Asset precompilation completed"))
			Expect(app.GetBody("/")).To(ContainSubstring("Welcome to Rails51 Webpacker!"))
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
