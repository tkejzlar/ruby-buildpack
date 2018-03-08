package integration_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRails51(t *testing.T) {
	t.Parallel()
	spec.Run(t, "Rails 5.1 (Webpack/Yarn) App", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "rails51")
			Expect(err).ToNot(HaveOccurred())
			app.SetEnv("BP_DEBUG", "1")
		})

		it("Installs node6 and runs", func() {
			Expect(app.PushAndConfirm()).To(Succeed())
			Expect(app.Log()).To(ContainSubstring("Installing node 6."))

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World"))
			Eventually(app.Log).Should(ContainSubstring(`Started GET "/" for`))

			By("Make sure supply does not change BuildDir", func() {
				Expect(app).To(HaveUnchangedAppdir("BuildDir Checksum Before Supply", "BuildDir Checksum After Supply"))
			})

			By("Make sure binstubs work", func() {
				command := exec.Command("cf", "ssh", app.Name(), "-c", "/tmp/lifecycle/launcher /home/vcap/app 'rails about' ''")
				var writer bytes.Buffer
				session, err := gexec.Start(command, &writer, &writer)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, 10, 0.25).Should(gexec.Exit(0))
			})
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
