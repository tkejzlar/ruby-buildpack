package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRelatuiveGemspecPath(t *testing.T) {
	t.Parallel()
	spec.Run(t, "App with relative gemspec path", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "relative_gemspec_path")
			Expect(err).ToNot(HaveOccurred())
		})

		it("loads the gem with the relative gemspec path", func() {
			Expect(app.PushAndConfirm()).To(Succeed())
			Expect(app.Log()).To(ContainSubstring("Using hola 0.0.0 from source at `gems/hola`"))
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
