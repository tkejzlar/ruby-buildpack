package brats_test

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ruby buildpack", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	Context("Unbuilt buildpack (eg github)", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "no_dependencies"))
			app.Buildpacks = []string{buildpacks.Unbuilt}
		})

		It("runs", func() {
			PushApp(app)
			Expect(app.Stdout.String()).To(ContainSubstring("-----> Download go 1.9"))

			Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby"))
			Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
		})
	})

	Context("deploying an app with an updated version of the same buildpack", func() {
		var bpName string
		BeforeEach(func() {
			bpName = "brats_ruby_changing_" + cutlass.RandStringRunes(6)

			app = cutlass.New(filepath.Join(bpDir, "fixtures", "no_dependencies"))
			app.Buildpacks = []string{bpName + "_buildpack"}
		})
		AfterEach(func() {
			Expect(cutlass.DeleteBuildpack(bpName)).To(Succeed())
		})

		It("prints useful warning message to stdout", func() {
			Expect(cutlass.CreateOrUpdateBuildpack(bpName, buildpacks.CachedFile)).To(Succeed())
			PushApp(app)
			Expect(app.Stdout.String()).ToNot(ContainSubstring("buildpack version changed from"))

			newFile := filepath.Join("/tmp", filepath.Base(buildpacks.CachedFile))
			Expect(libbuildpack.CopyFile(buildpacks.CachedFile, newFile)).To(Succeed())
			Expect(ioutil.WriteFile("/tmp/VERSION", []byte("NewVerson"), 0644)).To(Succeed())
			Expect(exec.Command("zip", "-d", newFile, "VERSION").Run()).To(Succeed())
			Expect(exec.Command("zip", "-j", "-u", newFile, "/tmp/VERSION").Run()).To(Succeed())

			Expect(cutlass.CreateOrUpdateBuildpack(bpName, newFile)).To(Succeed())
			PushApp(app)
			Expect(app.Stdout.String()).To(ContainSubstring("buildpack version changed from"))
		})
	})
})
