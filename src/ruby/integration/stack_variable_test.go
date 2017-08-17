package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stack environment should not change", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "sinatra"))
	})

	It("", func() {
		PushAppAndConfirm(app)
		PushAppAndConfirm(app)

		Expect(app.Stdout.String()).ToNot(ContainSubstring("Changing stack from"))
		Expect(app.Stdout.String()).ToNot(ContainSubstring("are the same file"))

		Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
	})
})
