package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestNoGemfile(t *testing.T) {
	t.Parallel()
	spec.Run(t, "App with No Gemfile", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "no_gemfile")
			Expect(err).ToNot(HaveOccurred())
		})

		when("Single/Final buildpack", func() {
			it.Before(func() {
				app.Buildpacks([]string{"ruby_buildpack"})
			})
			it("fails in finalize", func() {
				Expect(app.Push()).ToNot(Succeed())
				Expect(app.ConfirmBuildpack("")).To(Succeed())
				Eventually(app.Log).Should(ContainSubstring("Gemfile.lock required"))
			})
		})

		when("Supply buildpack", func() {
			it.Before(func() {
				if !cluster.HasMultiBuildpack() {
					t.Skip("API does not have multi buildpack support")
				}
				app.Buildpacks([]string{"ruby_buildpack", "binary_buildpack"})
			})
			it("deploys", func() {
				Expect(app.PushAndConfirm()).To(Succeed())
				Expect(app.Log()).To(ContainSubstring("Installing ruby"))

				By("running with the supplied ruby version", func() {
					defaultRubyVersion, err := DefaultVersion("ruby")
					Expect(err).ToNot(HaveOccurred())
					Expect(app.GetBody("/")).To(ContainSubstring("Ruby Version: " + defaultRubyVersion))
				})
			})
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
