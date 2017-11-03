package brats_test

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"ruby/brats/helper"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
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
			app.Buildpacks = []string{bpName + "_buildpack"}
		})
		AfterEach(func() {
			Expect(cutlass.DeleteBuildpack(bpName)).To(Succeed())
		})

		It("runs", func() {
			PushApp(app)
			Expect(app.Stdout.String()).To(ContainSubstring("-----> Download go "))

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

			newFile, err := helpers.ModifyBuildpack(path, func(path string, r io.Reader) (io.Reader, error) {
				if path == "VERSION" {
					return strings.NewReader("NewVersion"), nil
				}
				return r, nil
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(cutlass.CreateOrUpdateBuildpack(bpName, newFile)).To(Succeed())
			PushApp(app)
			Expect(app.Stdout.String()).To(MatchRegexp(`buildpack version changed from (\S+) to NewVersion`))
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
				appDir = CopyBratsRuby(rubyVersion)
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
					Expect(bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("Hello, bcrypt"))).ToNot(HaveOccurred())
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

	Describe("For all supported JRuby versions", func() {
		bpDir, err := cutlass.FindRoot()
		if err != nil {
			panic(err)
		}
		manifest, err := libbuildpack.NewManifest(bpDir, nil, time.Now())
		rubyVersions := manifest.AllDependencyVersions("jruby")
		var appDir string
		AfterEach(func() { os.RemoveAll(appDir) })

		for _, v := range rubyVersions {
			m := regexp.MustCompile(`ruby-(.*)-jruby-(.*)`).FindStringSubmatch(v)
			if len(m) != 3 {
				panic("Incorrect jruby version " + v)
			}
			fullRubyVersion := v
			rubyVersion := m[1]
			jrubyVersion := m[2]
			It("with JRuby version "+jrubyVersion+" and Ruby version "+rubyVersion, func() {
				appDir = CopyBratsJRuby(rubyVersion, jrubyVersion)
				app = cutlass.New(appDir)
				app.Memory = "512M"
				app.Buildpacks = []string{buildpacks.Cached}
				PushApp(app)

				By("installs the correct version of JRuby", func() {
					Expect(app.Stdout.String()).To(ContainSubstring("Installing jruby " + fullRubyVersion))
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
					Expect(bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("Hello, bcrypt"))).ToNot(HaveOccurred())
				})
				By("supports bson", func() {
					Expect(app.GetBody("/bson")).To(ContainSubstring("00040000"))
				})
				By("supports postgres", func() {
					Expect(app.GetBody("/pg")).To(ContainSubstring("The connection attempt failed."))
				})
				By("supports mysql", func() {
					Expect(app.GetBody("/mysql")).To(ContainSubstring("Communications link failure"))
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
			file, err := helper.ModifyBuildpackManifest(buildpackFile, func(m *helper.Manifest) {
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

			appDir = CopyBratsRuby("~> 2.1.0")
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
			manifest, err := libbuildpack.NewManifest(bpDir, nil, time.Now())
			Expect(err).ToNot(HaveOccurred())
			raw := manifest.AllDependencyVersions("ruby")
			vs := make([]*semver.Version, len(raw))
			for i, r := range raw {
				vs[i], err = semver.NewVersion(r)
				Expect(err).ToNot(HaveOccurred())
			}
			sort.Sort(semver.Collection(vs))
			version := vs[0].Original()

			appDir = CopyBratsRuby(version)
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
			file, err := helper.ModifyBuildpackManifest(buildpackFile, func(m *helper.Manifest) {
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

			appDir = CopyBratsRuby("~> 2.1.0")
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

			appDir := CopyBratsRuby(dep.Version)
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

	Describe("deploying an app that has sensitive environment variables", func() {
		BeforeEach(func() {
			appDir := CopyBratsRuby("~> 2.4")
			app = cutlass.New(appDir)
			app.Buildpacks = []string{buildpacks.Cached}
			app.SetEnv("MY_SPECIAL_VAR", "SUPER SENSITIVE DATA")
			PushApp(app)
		})
		AfterEach(func() { os.RemoveAll(app.Path) })

		It("will not write credentials to the app droplet", func() {
			Expect(app.DownloadDroplet(filepath.Join(app.Path, "droplet.tgz"))).To(Succeed())
			file, err := os.Open(filepath.Join(app.Path, "droplet.tgz"))
			Expect(err).ToNot(HaveOccurred())
			defer file.Close()
			gz, err := gzip.NewReader(file)
			Expect(err).ToNot(HaveOccurred())
			defer gz.Close()
			tr := tar.NewReader(gz)

			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				b, err := ioutil.ReadAll(tr)
				for _, content := range []string{"MY_SPECIAL_VAR", "SUPER SENSITIVE DATA"} {
					if strings.Contains(string(b), content) {
						Fail(fmt.Sprintf("Found sensitive string %s in %s", content, hdr.Name))
					}
				}
			}
		})
	})
})
