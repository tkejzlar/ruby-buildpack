package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestReadline(t *testing.T) {
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
			app, err = cluster.NewApp(bpDir, "with_readline")
			Expect(err).ToNot(HaveOccurred())
		})

		when("in an online environment", func() {
			it.Before(func() {
				SkipUnlessUncached(t)
			})

			it("", func() {
				Expect(app.PushAndConfirm()).To(Succeed())
				Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
				Expect(app.Log()).ToNot(ContainSubstring("cannot open shared object file"))
			})
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
