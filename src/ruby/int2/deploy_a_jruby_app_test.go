package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestJRuby(t *testing.T) {
	t.Parallel()
	spec.Run(t, "JRuby App", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "sinatra_jruby")
			Expect(err).ToNot(HaveOccurred())
			// TODO
			// app.Memory("512M")
		})

		when("without start command", func() {
			it("", func() {
				Expect(app.PushAndConfirm()).To(Succeed())

				By("the buildpack logged it installed a specific version of JRuby", func() {
					Expect(app.Log()).To(ContainSubstring("Installing openjdk"))
					Expect(app.Log()).To(MatchRegexp("ruby-2.3.\\d+-jruby-9.1.\\d+.0"))
					Expect(app.GetBody("/ruby")).To(MatchRegexp("jruby 2.3.\\d+"))
				})

				By("the OpenJDK runs properly", func() {
					Expect(app.Log()).ToNot(ContainSubstring("OpenJDK 64-Bit Server VM warning"))
				})
			})

			when("a cached buildpack", func() {
				it.Before(func() { SkipUnlessCached(t) })

				// TODO
				// AssertNoInternetTraffic("sinatra_jruby")
			})
		})
		when("with a jruby start command", func() {
			it.Before(func() {
				app, err = cluster.NewApp(bpDir, "jruby_start_command")
				Expect(err).ToNot(HaveOccurred())
				// TODO
				// app.Memory("512M")
			})

			it("stages and runs successfully", func() {
				Expect(app.PushAndConfirm()).To(Succeed())
			})
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
