package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRails4(t *testing.T) {
	t.Parallel()
	spec.Run(t, "Rails 4 App", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "rails4")
			Expect(err).ToNot(HaveOccurred())
		})

		when("in an offline environment", func() {
			it.Before(func() { SkipUnlessUncached(t) })

			it("", func() {
				Expect(app.PushAndConfirm()).To(Succeed())

				Expect(app.GetBody("/")).To(ContainSubstring("The Kessel Run"))
				Expect(app.Log()).To(ContainSubstring("Copy [/"))
			})

			// TODO
			// AssertNoInternetTraffic("rails4")
		})

		when("in an online environment", func() {
			it.Before(func() { SkipUnlessCached(t) })

			it("app has dependencies", func() {
				Expect(app.PushAndConfirm()).To(Succeed())
				Expect(app.Log()).To(ContainSubstring("Installing node 4."))
				Expect(app.Log()).To(ContainSubstring("Download [https://"))

				Expect(app.GetBody("/")).To(ContainSubstring("The Kessel Run"))
			})

			when("app has non vendored dependencies", func() {
				it.Before(func() {
					app, err = cluster.NewApp(bpDir, "rails4")
					Expect(err).ToNot(HaveOccurred())
				})
				it("", func() {
					// TODO
					// Expect(filepath.Join(app.Path, "vendor")).ToNot(BeADirectory())

					Expect(app.PushAndConfirm()).To(Succeed())

					Expect(app.GetBody("/")).To(ContainSubstring("The Kessel Run"))
				})

				// TODO
				// AssertUsesProxyDuringStagingIfPresent("rails4_not_vendored")
			})
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
