package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rails 4 App", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("in an offline environment", func() {
		BeforeEach(func() {
			if !cutlass.Cached {
				Skip("cached tests")
			}

			app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "rails4"))
		})

		It("", func() {
			PushAppAndConfirm(app)

			Expect(app.GetBody("/")).To(ContainSubstring("The Kessel Run"))

			Expect(app.Stdout.String()).To(ContainSubstring("Downloaded [file://"))
		})

		PIt("Assert no internet traffic", func() {
			// expect(app).not_to have_internet_traffic
		})
	})

	Context("in an online environment", func() {
		BeforeEach(func() {
			if cutlass.Cached {
				Skip("uncached tests")
			}
		})
		Context("app has dependencies", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "rails4"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(MatchRegexp("Downloaded.*node-4\\."))

				Expect(app.GetBody("/")).To(ContainSubstring("The Kessel Run"))
				Expect(app.Stdout.String()).To(ContainSubstring("Downloaded [https://"))
			})
		})

		Context("app has non vendored dependencies", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "rails4_not_vendored"))
			})

			It("", func() {
				// TODO next line
				// expect(Dir.exists?("cf_spec/fixtures/#{app_name}/vendor")).to eql(false)
				PushAppAndConfirm(app)

				Expect(app.GetBody("/")).To(ContainSubstring("The Kessel Run"))
			})
			PIt("uses a proxy during staging if present", func() {
				// expect(app).to use_proxy_during_staging
			})
		})
	})
})
