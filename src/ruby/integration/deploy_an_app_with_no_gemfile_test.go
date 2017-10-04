package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("App with No Gemfile", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "no_gemfile"))
	})

	// It("uses the version of ruby specified in Gemfile-APP", func() {
	// 	PushAppAndConfirm(app)
	// 	Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby 2.2.8"))
	// })

	Context("Single/Final buildpack", func() {
		BeforeEach(func() {
			app.Buildpacks = []string{"ruby_buildpack"}
		})
		It("fails in finalize", func() {
		})
	})

	Context("Supply buildpack", func() {
		BeforeEach(func() {
			app.Buildpacks = []string{"ruby_buildpack", "binary_buildpack"}
		})
		It("deploys", func() {
		})
	})
})
