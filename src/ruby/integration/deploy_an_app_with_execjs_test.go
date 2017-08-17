package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("requiring execjs", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "with_execjs"))
		app.SetEnv("BP_DEBUG", "1")
	})

	It("", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(MatchRegexp("Downloaded.*node-4\\."))

		Expect(app.GetBody("/")).To(ContainSubstring("Successfully required execjs"))
		Expect(app.Stdout.String()).ToNot(ContainSubstring("ExecJS::RuntimeUnavailable"))

		Expect(app.GetBody("/npm")).To(ContainSubstring("Usage: npm <command>"))
	})
})
