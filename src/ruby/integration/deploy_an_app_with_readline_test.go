package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ruby Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("in an online environment", func() {
		BeforeEach(func() {
			if cutlass.Cached {
				Skip("uncached tests")
			}

			app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "with_readline"))
		})

		It("", func() {
			PushAppAndConfirm(app)
			Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
			Expect(app.Stdout.String()).ToNot(ContainSubstring("cannot open shared object file"))
		})
	})
})
