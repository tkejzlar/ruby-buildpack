package integration_test

import (
	"testing"

	"github.com/cloudfoundry/libbuildpack/cfapi"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestPushSecondTime(t *testing.T) {
	t.Parallel()
	spec.Run(t, "pushing an app a second time", func(t *testing.T, when spec.G, it spec.S) {
		var app cfapi.App
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
			if app != nil {
				app.Destroy()
			}
		})
		it.Before(func() {
			app, err = cluster.NewApp(bpDir, "sinatra")
			Expect(err).ToNot(HaveOccurred())
		})

		RestoringVendorBundle := "Restoring vendor_bundle from cache"
		DownloadRegexp := `Download \[.*/bundler\-.*\.tgz\]`
		CopyRegexp := `Copy \[.*/bundler\-.*\.tgz\]`

		it("uses the cache and runs", func() {
			Expect(app.PushAndConfirm()).To(Succeed())
			Expect(app.Log()).ToNot(ContainSubstring(RestoringVendorBundle))
			if !cutlass.Cached {
				Expect(app.Log()).To(MatchRegexp(DownloadRegexp))
				Expect(app.Log()).ToNot(MatchRegexp(CopyRegexp))
			}
			Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))

			app.ResetLog()
			Expect(app.PushAndConfirm()).To(Succeed())
			Expect(app.Log()).To(ContainSubstring(RestoringVendorBundle))
			if !cutlass.Cached {
				Expect(app.Log()).To(MatchRegexp(CopyRegexp))
				Expect(app.Log()).ToNot(MatchRegexp(DownloadRegexp))
			}
			Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))

			app.ResetLog()
			Expect(app.PushAndConfirm()).To(Succeed())
			Expect(app.Log()).To(ContainSubstring(RestoringVendorBundle))
			if !cutlass.Cached {
				Expect(app.Log()).To(MatchRegexp(CopyRegexp))
				Expect(app.Log()).ToNot(MatchRegexp(DownloadRegexp))
			}
			Expect(app.GetBody("/")).To(ContainSubstring("Hello world!"))
		})
	}, spec.Parallel(), spec.Report(report.Terminal{}))
}
