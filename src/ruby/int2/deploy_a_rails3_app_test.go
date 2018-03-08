package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRails3(t *testing.T) {
	t.Parallel()
	spec.Run(t, "Rails 3 App", func(t *testing.T, when spec.G, it spec.S) {
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
			app, err = cluster.NewApp(bpDir, "rails3_mri_200")
			Expect(err).ToNot(HaveOccurred())
		})

		it("in an online environment", func() {
			app.SetEnv("DATABASE_URL", "sqlite3://db/test.db")
			Expect(app.PushAndConfirm()).To(Succeed())

			By("the app can be visited in the browser", func() {
				Expect(app.GetBody("/")).To(ContainSubstring("hello"))
			})

			By("the app did not include the static asset or logging gems from Heroku", func() {
				By("the rails 3 plugins are installed automatically", func() {
					files, err := app.Files("/app/vendor/plugins")
					Expect(err).ToNot(HaveOccurred())
					Expect(files).To(ContainElement("/app/vendor/plugins/rails3_serve_static_assets/init.rb"))
					Expect(files).To(ContainElement("/app/vendor/plugins/rails_log_stdout/init.rb"))
				})
			})

			By("we include a rails logger message in the initializer", func() {
				By("the log message is visible in the cf cli app logging", func() {
					Expect(app.Log()).To(ContainSubstring("Logging is being redirected to STDOUT with rails_log_stdout plugin"))
				})
			})

			By("we include a static asset", func() {
				By("app serves the static asset", func() {
					Expect(app.GetBody("/assets/application.css")).To(ContainSubstring("body{color:red}"))
				})
			})
		})

		// TODO
		// AssertNoInternetTraffic("rails3_mri_200")
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
