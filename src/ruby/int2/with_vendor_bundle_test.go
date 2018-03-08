package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestWithVendorBundle(t *testing.T) {
	t.Parallel()
	spec.Run(t, "App with dependencies installed in vendor/bundle", func(t *testing.T, when spec.G, it spec.S) {
		var app cfapi.App
		var err error
		var g *GomegaWithT
		var Expect func(actual interface{}, extra ...interface{}) GomegaAssertion
		var Eventually func(actual interface{}, intervals ...interface{}) GomegaAsyncAssertion
		var By func(string, func())
		it.Before(func() {
			g = NewGomegaWithT(t)
			Expect = g.Expect
			Eventually = g.Eventually
			By = func(_ string, f func()) { f() }
		})
		it.After(func() {
			if app != nil {
				app.Destroy()
			}
		})
		it.Before(func() {
			app, err = cluster.NewApp(bpDir, "with_vendor_bundle")
			Expect(err).ToNot(HaveOccurred())
		})

		it("", func() {
			Expect(app.PushAndConfirm()).To(Succeed())

			By("remove vendor/bundle directory", func() {
				Expect(app.Log()).To(ContainSubstring("Removing `vendor/bundle`"))
				Expect(app.Log()).To(ContainSubstring("Checking in `vendor/bundle` is not supported. Please remove this directory and add it to your .gitignore. To vendor your gems with Bundler, use `bundle pack` instead."))

				files, err := app.Files("app/vendor")
				Expect(err).ToNot(HaveOccurred())
				Expect(files).ToNot(ContainElement("app/vendor/bundle"))
			})

			By("has required gems at runtime", func() {
				Expect(app.GetBody("/")).To(ContainSubstring("Healthy"))
				Eventually(app.Log).Should(ContainSubstring("This is red"))
				Eventually(app.Log).Should(ContainSubstring("This is blue"))
			})
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
