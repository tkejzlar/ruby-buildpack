package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("app using system yaml library", func() {
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

	It("displays metasyntactic variables as yaml", func() {
		PushAppAndConfirm(app)
		Expect(app.GetBody("/yaml")).To(ContainSubstring(`---
foo:
- bar
- baz
- quux
`))
	})
})
