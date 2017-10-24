package brats_test

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"golang.org/x/crypto/bcrypt"

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

	Describe("deploying an app with an updated version of the same buildpack", func() {
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

	FDescribe("For all supported Ruby versions", func() {
		bpDir, err := cutlass.FindRoot()
		if err != nil {
			panic(err)
		}
		manifest, err := libbuildpack.NewManifest(bpDir, nil, time.Now())
		rubyVersions := manifest.AllDependencyVersions("ruby")
		rubyVersions = rubyVersions[:2] // FIXME remove this line (debug only)

		for _, v := range rubyVersions {
			rubyVersion := v
			It("Ruby version "+rubyVersion, func() {
				dir, err := cutlass.CopyFixture(filepath.Join(bpDir, "fixtures", "simple_brats"))
				Expect(err).ToNot(HaveOccurred())
				data, err := ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
				Expect(err).ToNot(HaveOccurred())
				data = bytes.Replace(data, []byte("<%= ruby_version %>"), []byte(rubyVersion), -1)
				Expect(ioutil.WriteFile(filepath.Join(dir, "Gemfile"), data, 0644)).To(Succeed())

				app = cutlass.New(dir)
				app.Buildpacks = []string{buildpacks.Cached}
				PushApp(app)

				By("installs the correct version of Ruby", func() {
					Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby " + rubyVersion))
					Expect(app.GetBody("/version")).To(ContainSubstring(rubyVersion))
				})
				By("runs a simple webserver", func() {
					Expect(app.GetBody("/")).To(ContainSubstring("Hello, World"))
				})
				By("parses XML with nokogiri", func() {
					Expect(app.GetBody("/nokogiri")).To(ContainSubstring("Hello, World"))
				})
				By("supports EventMachine", func() {
					Expect(app.GetBody("/em")).To(ContainSubstring("Hello, EventMachine"))
				})
				By("encrypts with bcrypt", func() {
					hashedPassword, err := app.GetBody("/bcrypt")
					Expect(err).ToNot(HaveOccurred())
					Expect(bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("Hello, bcrypt"))).To(BeTrue())
				})
				By("supports bson", func() {
					Expect(app.GetBody("/bson")).To(ContainSubstring("00040000"))
				})
				By("supports postgres", func() {
					Expect(app.GetBody("/pg")).To(ContainSubstring("could not connect to server: No such file or directory"))
				})
				By("supports mysql2", func() {
					Expect(app.GetBody("/mysql2")).To(ContainSubstring("Unknown MySQL server host 'testing'"))
				})
			})
		}
	})
})
