package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCustomGemfile(t *testing.T) {
	t.Parallel()
	spec.Run(t, "App with custom Gemfile", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "custom_gemfile")
			Expect(err).ToNot(HaveOccurred())
		})

		it("uses the version of ruby specified in Gemfile-APP", func() {
			Expect(app.PushAndConfirm()).To(Succeed())
			Expect(app.Log()).To(ContainSubstring("Installing ruby 2.2.9"))
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
