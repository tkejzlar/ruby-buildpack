package brats_test

import (
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"ruby/brats/helper"
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
		var bpName string
		BeforeEach(func() {
			bpName = "brats_ruby_unbuilt_" + buildpacks.BpVersion
			cmd := exec.Command("git", "archive", "-o", filepath.Join("/tmp", bpName+".zip"), "HEAD")
			cmd.Dir = bpDir
			Expect(cmd.Run()).To(Succeed())
			Expect(cutlass.CreateOrUpdateBuildpack(bpName, filepath.Join("/tmp", bpName+".zip"))).To(Succeed())
			Expect(os.Remove(filepath.Join("/tmp", bpName+".zip"))).To(Succeed())

			app = cutlass.New(filepath.Join(bpDir, "fixtures", "no_dependencies"))
			app.Buildpacks = []string{bpName}
		})
		AfterEach(func() {
			Expect(cutlass.DeleteBuildpack(bpName)).To(Succeed())
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

	Describe("For all supported Ruby versions", func() {
		bpDir, err := cutlass.FindRoot()
		if err != nil {
			panic(err)
		}
		manifest, err := libbuildpack.NewManifest(bpDir, nil, time.Now())
		rubyVersions := manifest.AllDependencyVersions("ruby")
		var appDir string
		AfterEach(func() { os.RemoveAll(appDir) })

		for _, v := range rubyVersions {
			rubyVersion := v
			It("Ruby version "+rubyVersion, func() {
				appDir = CopySimpleBrats(rubyVersion)
				app = cutlass.New(appDir)
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
					Expect(hashedPassword).ToNot(Equal(""))
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

	Describe("staging with ruby buildpack that sets EOL on dependency", func() {
		var (
			eolDate       string
			buildpackFile string
			bpName        string
			appDir        string
		)
		JustBeforeEach(func() {
			eolDate = time.Now().AddDate(0, 0, 10).Format("2006-01-02")
			file, err := helper.CopyBuildpack(buildpackFile, func(m *helper.Manifest) {
				for _, eol := range m.DependencyDeprecationDates {
					if eol.Name == "ruby" {
						eol.Date = eolDate
					}
				}
			})
			Expect(err).ToNot(HaveOccurred())
			bpName = "brats_ruby_eol_" + cutlass.RandStringRunes(6)
			Expect(cutlass.CreateOrUpdateBuildpack(bpName, file)).To(Succeed())
			os.Remove(file)

			appDir = CopySimpleBrats("~> 2.1.0")
			app = cutlass.New(appDir)
			app.Buildpacks = []string{bpName + "_buildpack"}
			PushApp(app)
		})
		AfterEach(func() {
			Expect(cutlass.DeleteBuildpack(bpName)).To(Succeed())
			Expect(os.RemoveAll(appDir)).To(Succeed())
		})

		Context("using an uncached buildpack", func() {
			BeforeEach(func() {
				buildpackFile = buildpacks.UncachedFile
			})
			It("warns about end of life", func() {
				Expect(app.Stdout.String()).To(MatchRegexp("WARNING.*ruby 2.1.x will no longer be available in new buildpacks released after"))
			})
		})

		Context("using a cached buildpack", func() {
			BeforeEach(func() {
				buildpackFile = buildpacks.CachedFile
			})
			It("warns about end of life", func() {
				Expect(app.Stdout.String()).To(MatchRegexp("WARNING.*ruby 2.1.x will no longer be available in new buildpacks released after"))
			})
		})
	})

	Describe("staging with a version of ruby that is not the latest patch release in the manifest", func() {
		var appDir string
		BeforeEach(func() {
			appDir = CopySimpleBrats("2.4.1") // FIXME determine from manifest
			app = cutlass.New(appDir)
			app.Buildpacks = []string{buildpacks.Cached}
			PushApp(app)
		})
		AfterEach(func() { os.RemoveAll(appDir) })

		It("logs a warning that tells the user to upgrade the dependency", func() {
			Expect(app.Stdout.String()).To(MatchRegexp("WARNING.*A newer version of ruby is available in this buildpack"))
		})
	})

	Describe("staging with custom buildpack that uses credentials in manifest dependency uris", func() {
		var (
			buildpackFile string
			bpName        string
			appDir        string
		)
		JustBeforeEach(func() {
			file, err := helper.CopyBuildpack(buildpackFile, func(m *helper.Manifest) {
				for _, d := range m.Dependencies {
					uri, err := url.Parse(d.URI)
					uri.User = url.UserPassword("login", "password")
					Expect(err).ToNot(HaveOccurred())
					d.URI = uri.String()
				}
			})
			Expect(err).ToNot(HaveOccurred())
			bpName = "brats_ruby_eol_" + cutlass.RandStringRunes(6)
			Expect(cutlass.CreateOrUpdateBuildpack(bpName, file)).To(Succeed())
			os.Remove(file)

			appDir = CopySimpleBrats("~> 2.1.0")
			app = cutlass.New(appDir)
			app.Buildpacks = []string{bpName + "_buildpack"}
			PushApp(app)
		})
		AfterEach(func() {
			Expect(cutlass.DeleteBuildpack(bpName)).To(Succeed())
			Expect(os.RemoveAll(appDir)).To(Succeed())
		})
		Context("using an uncached buildpack", func() {
			BeforeEach(func() {
				buildpackFile = buildpacks.UncachedFile
			})
			It("does not include credentials in logged dependency uris", func() {
				Expect(app.Stdout.String()).To(MatchRegexp(`ruby\-[\d\.]+\-linux\-x64\-[\da-f]+\.tgz`))
				Expect(app.Stdout.String()).ToNot(ContainSubstring("login"))
				Expect(app.Stdout.String()).ToNot(ContainSubstring("password"))
			})
		})
		Context("using a cached buildpack", func() {
			BeforeEach(func() {
				buildpackFile = buildpacks.UncachedFile
			})
			It("does not include credentials in logged dependency file paths", func() {
				Expect(app.Stdout.String()).To(MatchRegexp(`ruby\-[\d\.]+\-linux\-x64\-[\da-f]+\.tgz`))
				Expect(app.Stdout.String()).ToNot(ContainSubstring("login"))
				Expect(app.Stdout.String()).ToNot(ContainSubstring("password"))
			})
		})
	})

	Describe("deploying an app that has an executable .profile script", func() {
		BeforeEach(func() {
			manifest, err := libbuildpack.NewManifest(bpDir, nil, time.Now())
			dep, err := manifest.DefaultVersion("ruby")
			Expect(err).ToNot(HaveOccurred())

			appDir := CopySimpleBrats(dep.Version)
			AddDotProfileScriptToApp(appDir)
			app = cutlass.New(appDir)
			app.Buildpacks = []string{buildpacks.Cached}
			PushApp(app)
		})
		AfterEach(func() { os.RemoveAll(app.Path) })

		It("executes the .profile script", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("PROFILE_SCRIPT_IS_PRESENT_AND_RAN"))
		})
		It("does not let me view the .profile script", func() {
			_, headers, err := app.Get("/.profile", map[string]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(headers).To(HaveKeyWithValue("StatusCode", []string{"404"}))
		})
	})

	PDescribe("deploying an app that has sensitive environment variables", func() {
		It("will not write credentials to the app droplet", func() {
		})
	})
})
