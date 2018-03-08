package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestExecJS(t *testing.T) {
	t.Parallel()
	spec.Run(t, "requiring execjs", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "with_execjs")
			Expect(err).ToNot(HaveOccurred())
			app.SetEnv("BP_DEBUG", "1")
		})

		it("", func() {
			Expect(app.PushAndConfirm()).To(Succeed())
			Expect(app.Log()).To(ContainSubstring("Installing node 4."))
			Expect(app).To(HaveUnchangedAppdir("BuildDir Checksum Before Supply", "BuildDir Checksum After Supply"))

			Expect(app.GetBody("/")).To(ContainSubstring("Successfully required execjs"))
			Expect(app.Log()).ToNot(ContainSubstring("ExecJS::RuntimeUnavailable"))

			Expect(app.GetBody("/npm")).To(ContainSubstring("Usage: npm <command>"))
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
