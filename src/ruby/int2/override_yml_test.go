package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestOverrideYml(t *testing.T) {
	t.Parallel()
	spec.Run(t, "override yml", func(t *testing.T, when spec.G, it spec.S) {
		var app cfapi.App
		var buildpackName string
		var err error
		var g *GomegaWithT
		var Expect func(actual interface{}, extra ...interface{}) GomegaAssertion
		var Eventually func(actual interface{}, intervals ...interface{}) GomegaAsyncAssertion
		it.Before(func() {
			g = NewGomegaWithT(t)
			Expect = g.Expect
			Eventually = g.Eventually
		})
		it.After(func() {
			if buildpackName != "" {
				cutlass.DeleteBuildpack(buildpackName)
			}
			if app != nil {
				app.Destroy()
			}
		})
		it.Before(func() {
			if !cluster.HasMultiBuildpack() {
				t.Skip("Multi buildpack support is required")
			}

			buildpackName = "override_yml_" + cutlass.RandStringRunes(5)
			Expect(cutlass.CreateOrUpdateBuildpack(buildpackName, filepath.Join(bpDir, "fixtures", "overrideyml_bp"))).To(Succeed())
			app, err = cluster.NewApp(bpDir, "with_execjs")
			Expect(err).ToNot(HaveOccurred())
			app.Buildpacks([]string{buildpackName + "_buildpack", "ruby_buildpack"})
		})

		it("Forces nodejs from override buildpack, installs ruby from ruby buildpack", func() {
			Expect(app.Push()).ToNot(Succeed())
			Expect(app.Log()).To(ContainSubstring("-----> OverrideYML Buildpack"))
			Expect(app.ConfirmBuildpack("")).To(Succeed())

			Eventually(app.Log).Should(ContainSubstring("-----> Installing ruby"))

			Eventually(app.Log).Should(ContainSubstring("-----> Installing node"))
			Eventually(app.Log).Should(MatchRegexp("Copy .*/node.tgz"))
			Eventually(app.Log).Should(ContainSubstring("Unable to install node: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc"))
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
