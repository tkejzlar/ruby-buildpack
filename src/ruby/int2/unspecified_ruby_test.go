package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnspecifiedRuby(t *testing.T) {
	t.Parallel()
	spec.Run(t, "CF Ruby Buildpack", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "unspecified_ruby")
			Expect(err).ToNot(HaveOccurred())
		})

		it("uses the default ruby version when ruby version is not specified", func() {
			Expect(app.PushAndConfirm()).To(Succeed())
			defaultRubyVersion, err := DefaultVersion("ruby")
			Expect(err).ToNot(HaveOccurred())

			Eventually(app.Log).Should(ContainSubstring("Installing ruby %s", defaultRubyVersion))
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
