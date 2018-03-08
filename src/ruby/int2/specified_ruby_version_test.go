package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestSpecifiedRubyVersion(t *testing.T) {
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
			app, err = cluster.NewApp(bpDir, "specified_ruby_version")
			Expect(err).ToNot(HaveOccurred())
		})

		it("", func() {
			Expect(app.PushAndConfirm()).To(Succeed())

			By("uses the specified ruby version", func() {
				Expect(app.Log()).To(ContainSubstring("Installing ruby 2.2."))
			})

			By("running a task", func() {
				if !cluster.HasTask() {
					t.Skip("Running against CF without run task support")
				}

				By("can find the specifed ruby in the container", func() {
					_, err := app.RunTask(`echo "RUNNING A TASK: $(ruby --version)"`)
					Expect(err).ToNot(HaveOccurred())

					Eventually(app.Log).Should(ContainSubstring("RUNNING A TASK: ruby 2.2."))
				})
			})
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
