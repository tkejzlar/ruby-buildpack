package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestSystemYaml(t *testing.T) {
	t.Parallel()
	spec.Run(t, "app using system yaml library", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "sinatra")
			Expect(err).ToNot(HaveOccurred())
		})
		it("displays metasyntactic variables as yaml", func() {
			Expect(app.PushAndConfirm()).To(Succeed())
			Expect(app.GetBody("/yaml")).To(ContainSubstring(`---
foo:
- bar
- baz
- quux
`))
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
