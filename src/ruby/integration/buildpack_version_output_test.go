package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version output", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("in an online environment", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "sinatra"))
		})

		It("logs buildpack version", func() {
			PushAppAndConfirm(app)
			Expect(app.Stdout.String()).To(ContainSubstring("-------> Buildpack version "))
		})
	})
})
