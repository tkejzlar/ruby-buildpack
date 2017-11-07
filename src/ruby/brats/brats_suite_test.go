package brats_test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	return bratshelper.InitBpData().Marshal()
}, func(data []byte) {
	// Run on all nodes
	bratshelper.Data.Unmarshal(data)

	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(bratshelper.Data.Cached, "_buildpack", "", 1))).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(bratshelper.Data.Uncached, "_buildpack", "", 1))).To(Succeed())
	Expect(os.Remove(bratshelper.Data.CachedFile)).To(Succeed())
})

func TestBrats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brats Suite")
}

func CopyBrats(rubyVersion string) *cutlass.App {
	dir, err := cutlass.CopyFixture(filepath.Join(bratshelper.Data.BpDir, "fixtures", "brats_ruby"))
	Expect(err).ToNot(HaveOccurred())
	data, err := ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
	Expect(err).ToNot(HaveOccurred())
	if rubyVersion == "" {
		manifest, err := libbuildpack.NewManifest(bratshelper.Data.BpDir, nil, time.Now())
		Expect(err).ToNot(HaveOccurred())
		dep, err := manifest.DefaultVersion("ruby")
		Expect(err).ToNot(HaveOccurred())
		rubyVersion = dep.Version
	} else if strings.Contains(rubyVersion, "x") {
		manifest, err := libbuildpack.NewManifest(bratshelper.Data.BpDir, nil, time.Now())
		Expect(err).ToNot(HaveOccurred())
		depVersions := manifest.AllDependencyVersions("ruby")
		rubyVersion, err = libbuildpack.FindMatchingVersion(rubyVersion, depVersions)
		Expect(err).ToNot(HaveOccurred())
	}
	data = bytes.Replace(data, []byte("<%= ruby_version %>"), []byte(rubyVersion), -1)
	Expect(ioutil.WriteFile(filepath.Join(dir, "Gemfile"), data, 0644)).To(Succeed())

	return cutlass.New(dir)
}

// func CopyBratsJRuby(rubyVersion, jrubyVersion string) *cutlass.App {
// 	dir, err := cutlass.CopyFixture(filepath.Join(bpDir, "fixtures", "brats_jruby"))
// 	Expect(err).ToNot(HaveOccurred())
// 	data, err := ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
// 	Expect(err).ToNot(HaveOccurred())
// 	if rubyVersion == "" {
// 		manifest, err := libbuildpack.NewManifest(bpDir, nil, time.Now())
// 		Expect(err).ToNot(HaveOccurred())
// 		dep, err := manifest.DefaultVersion("ruby")
// 		Expect(err).ToNot(HaveOccurred())
// 		rubyVersion = dep.Version
// 	}
// 	data = bytes.Replace(data, []byte("<%= ruby_version %>"), []byte(rubyVersion), -1)
// 	data = bytes.Replace(data, []byte("<%= engine_version %>"), []byte(jrubyVersion), -1)
// 	Expect(ioutil.WriteFile(filepath.Join(dir, "Gemfile"), data, 0644)).To(Succeed())
//  return cutlass.New(dir)
// }

// package brats_test

// import (
// 	"bytes"
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/cloudfoundry/libbuildpack/cutlass"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// )

// var bpDir string
// var buildpacks struct {
// 	BpVersion    string
// 	Cached       string
// 	CachedFile   string
// 	Uncached     string
// 	UncachedFile string
// }

// func init() {
// 	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
// 	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
// 	flag.Parse()
// }

// var _ = SynchronizedBeforeSuite(func() []byte {
// 	// Run once
// 	buildpacks.BpVersion = cutlass.RandStringRunes(6)
// 	buildpacks.Cached = "brats_ruby_cached_" + buildpacks.BpVersion
// 	buildpacks.Uncached = "brats_ruby_uncached_" + buildpacks.BpVersion

// 	var err error
// 	bpDir, err = cutlass.FindRoot()
// 	Expect(err).NotTo(HaveOccurred())

// 	var wg sync.WaitGroup
// 	wg.Add(2)
// 	go func() {
// 		defer wg.Done()
// 		fmt.Fprintln(os.Stderr, "Start build cached buildpack")
// 		cachedBuildpack, err := cutlass.PackageUniquelyVersionedBuildpackExtra(buildpacks.Cached, buildpacks.BpVersion, true)
// 		Expect(err).NotTo(HaveOccurred())
// 		buildpacks.CachedFile = cachedBuildpack.File
// 		fmt.Fprintln(os.Stderr, "Finish cached buildpack")
// 	}()
// 	go func() {
// 		defer wg.Done()
// 		fmt.Fprintln(os.Stderr, "Start build uncached buildpack")
// 		uncachedBuildpack, err := cutlass.PackageUniquelyVersionedBuildpackExtra(buildpacks.Uncached, buildpacks.BpVersion, false)
// 		Expect(err).NotTo(HaveOccurred())
// 		buildpacks.UncachedFile = uncachedBuildpack.File
// 		fmt.Fprintln(os.Stderr, "Finish uncached buildpack")
// 	}()
// 	wg.Wait()

// 	buildpacks.Cached = buildpacks.Cached + "_buildpack"
// 	buildpacks.Uncached = buildpacks.Uncached + "_buildpack"

// 	// Marshall for run all nodes
// 	data, err := json.Marshal(buildpacks)
// 	Expect(err).NotTo(HaveOccurred())

// 	return data
// }, func(data []byte) {
// 	// Run on all nodes
// 	err := json.Unmarshal(data, &buildpacks)
// 	Expect(err).NotTo(HaveOccurred())

// 	bpDir, err = cutlass.FindRoot()
// 	Expect(err).NotTo(HaveOccurred())

// 	cutlass.SeedRandom()
// 	cutlass.DefaultStdoutStderr = GinkgoWriter
// })

// var _ = SynchronizedAfterSuite(func() {
// 	// Run on all nodes
// }, func() {
// 	// Run once
// 	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
// 	Expect(cutlass.DeleteBuildpack(strings.Replace(buildpacks.Cached, "_buildpack", "", 1))).To(Succeed())
// 	Expect(cutlass.DeleteBuildpack(strings.Replace(buildpacks.Uncached, "_buildpack", "", 1))).To(Succeed())
// 	Expect(os.Remove(buildpacks.CachedFile)).To(Succeed())
// })

// func TestBrats(t *testing.T) {
// 	RegisterFailHandler(Fail)
// 	RunSpecs(t, "Brats Suite")
// }

// func PushApp(app *cutlass.App) {
// 	Expect(app.Push()).To(Succeed())
// 	Eventually(app.InstanceStates, 20*time.Second).Should(Equal([]string{"RUNNING"}))
// }

// func DestroyApp(app *cutlass.App) *cutlass.App {
// 	if app != nil {
// 		app.Destroy()
// 	}
// 	return nil
// }

// func CopyBratsRuby(rubyVersion string) string {
// 	dir, err := cutlass.CopyFixture(filepath.Join(bpDir, "fixtures", "brats_ruby"))
// 	Expect(err).ToNot(HaveOccurred())
// 	data, err := ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
// 	Expect(err).ToNot(HaveOccurred())
// 	data = bytes.Replace(data, []byte("<%= ruby_version %>"), []byte(rubyVersion), -1)
// 	Expect(ioutil.WriteFile(filepath.Join(dir, "Gemfile"), data, 0644)).To(Succeed())
// 	return dir
// }

// func CopyBratsJRuby(rubyVersion, jrubyVersion string) string {
// 	dir, err := cutlass.CopyFixture(filepath.Join(bpDir, "fixtures", "brats_jruby"))
// 	Expect(err).ToNot(HaveOccurred())
// 	data, err := ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
// 	Expect(err).ToNot(HaveOccurred())
// 	data = bytes.Replace(data, []byte("<%= ruby_version %>"), []byte(rubyVersion), -1)
// 	data = bytes.Replace(data, []byte("<%= engine_version %>"), []byte(jrubyVersion), -1)
// 	Expect(ioutil.WriteFile(filepath.Join(dir, "Gemfile"), data, 0644)).To(Succeed())
// 	return dir
// }

// func AddDotProfileScriptToApp(dir string) {
// 	profilePath := filepath.Join(dir, ".profile")
// 	Expect(ioutil.WriteFile(profilePath, []byte(`#!/usr/bin/env bash
// echo PROFILE_SCRIPT_IS_PRESENT_AND_RAN
// `), 0755)).To(Succeed())
// }
