package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deploy with an override.yml buildpack", func() {
	var app *cutlass.App
	var bpOverride string

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil

		if bpOverride != "" {
			cutlass.DeleteBuildpack(bpOverride)
			bpOverride = ""
		}
	})

	BeforeEach(func() {
		if !ApiHasMultiBuildpack() {
			Skip("API does not have multi buildpack support")
		}

		bpOverride = "ruby_override_yml_" + cutlass.RandStringRunes(6)
		cutlass.CreateOrUpdateBuildpack(bpOverride, filepath.Join(bpDir, "fixtures", "override_yml_buildpack"))

		app = cutlass.New(filepath.Join(bpDir, "fixtures", "sinatra"))
		app.Buildpacks = []string{bpOverride + "_buildpack", "ruby_buildpack"}
	})

	It("uses the newly specified default ruby version", func() {
		PushAppAndConfirm(app)

		Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby 2.2.7"))
		Expect(app.GetBody("/version")).To(ContainSubstring("Ruby Version: 2.2.7"))
	})
})
