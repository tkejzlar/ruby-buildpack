package integration_test

import (
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rails 5.1 (Webpack/Yarn) App", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("in an online environment", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "rails51"))
		})

		It("", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(MatchRegexp("Downloaded.*node-6\\."))

			Expect(app.GetBody("/")).To(ContainSubstring("Hello World"))
			Eventually(func() string { return app.Stdout.String() }, 10*time.Second).Should(ContainSubstring(`Started GET "/" for`))
		})
	})
})
