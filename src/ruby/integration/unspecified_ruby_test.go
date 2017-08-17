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

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "unspecified_ruby"))
		app.SetEnv("BP_DEBUG", "1")
	})

	It("", func() {
		PushAppAndConfirm(app)

		By("uses the default ruby version when ruby version is not specified", func() {
			// TODO pull version from manifest.yml
			Expect(app.Stdout.String()).To(ContainSubstring("Using Ruby version: ruby-2.4.1"))
		})

		By("pulls the default version from the manifest for ruby", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("DEBUG: default_version_for ruby is"))
		})
	})
})
