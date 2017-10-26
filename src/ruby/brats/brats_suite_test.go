package brats_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var bpDir string
var buildpacks struct {
	Unbuilt    string
	Cached     string
	CachedFile string
}

func init() {
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	bpVersion := cutlass.RandStringRunes(6)
	buildpacks.Cached = "brats_ruby_cached_" + bpVersion
	buildpacks.Unbuilt = "brats_ruby_unbuilt_" + bpVersion

	bpDir, err := cutlass.FindRoot()
	Expect(err).NotTo(HaveOccurred())

	// Build cached buildpack
	cachedBuildpack, err := cutlass.PackageUniquelyVersionedBuildpackExtra(buildpacks.Cached, bpVersion, true)
	Expect(err).NotTo(HaveOccurred())
	buildpacks.CachedFile = cachedBuildpack.File

	cmd := exec.Command("git", "archive", "-o", filepath.Join("/tmp", buildpacks.Unbuilt+".zip"), "HEAD")
	cmd.Dir = bpDir
	Expect(cmd.Run()).To(Succeed())
	Expect(cutlass.CreateOrUpdateBuildpack(buildpacks.Unbuilt, filepath.Join("/tmp", buildpacks.Unbuilt+".zip"))).To(Succeed())
	Expect(os.Remove(filepath.Join("/tmp", buildpacks.Unbuilt+".zip"))).To(Succeed())

	buildpacks.Cached = buildpacks.Cached + "_buildpack"
	buildpacks.Unbuilt = buildpacks.Unbuilt + "_buildpack"

	// Marshall for run all nodes
	data, err := json.Marshal(buildpacks)
	Expect(err).NotTo(HaveOccurred())

	return data
}, func(data []byte) {
	// Run on all nodes
	err := json.Unmarshal(data, &buildpacks)
	Expect(err).NotTo(HaveOccurred())

	bpDir, err = cutlass.FindRoot()
	Expect(err).NotTo(HaveOccurred())

	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(buildpacks.Unbuilt, "_buildpack", "", 1))).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(buildpacks.Cached, "_buildpack", "", 1))).To(Succeed())
	Expect(os.Remove(buildpacks.CachedFile)).To(Succeed())
})

func TestBrats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brats Suite")
}

func PushApp(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(app.InstanceStates, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}

func DestroyApp(app *cutlass.App) *cutlass.App {
	if app != nil {
		app.Destroy()
	}
	return nil
}

func CopySimpleBrats(rubyVersion string) string {
	dir, err := cutlass.CopyFixture(filepath.Join(bpDir, "fixtures", "simple_brats"))
	Expect(err).ToNot(HaveOccurred())
	data, err := ioutil.ReadFile(filepath.Join(dir, "Gemfile"))
	Expect(err).ToNot(HaveOccurred())
	data = bytes.Replace(data, []byte("<%= ruby_version %>"), []byte(rubyVersion), -1)
	Expect(ioutil.WriteFile(filepath.Join(dir, "Gemfile"), data, 0644)).To(Succeed())
	return dir
}

func AddDotProfileScriptToApp(dir string) {
	profilePath := filepath.Join(dir, ".profile")
	Expect(ioutil.WriteFile(profilePath, []byte(`#!/usr/bin/env bash
echo PROFILE_SCRIPT_IS_PRESENT_AND_RAN
`), 0755)).To(Succeed())
}
