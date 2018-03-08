package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnsupportedDependency(t *testing.T) {
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
			app, err = cluster.NewApp(bpDir, "unsupported_ruby")
			Expect(err).ToNot(HaveOccurred())
		})

		it("displays a nice error message when Ruby 99.99.99 is specified", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.ConfirmBuildpack("")).To(Succeed())
			Eventually(app.Log).Should(ContainSubstring("No Matching versions, ruby = 99.99.99 not found in this buildpack"))
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
