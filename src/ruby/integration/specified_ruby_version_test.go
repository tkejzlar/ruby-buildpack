package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Ruby Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "specified_ruby_version"))
	})

	It("", func() {
		PushAppAndConfirm(app)

		By("uses the specified ruby version", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby 2.2.7"))
		})
	})

	PIt("running a task", func() {
		// before { skip_if_no_run_task_support_on_targeted_cf }

		// it 'can find the specifed ruby in the container' do
		//   expect(@app).to be_running

		//   Open3.capture2e('cf','run-task', 'specified_ruby_version', 'echo "RUNNING A TASK: $(ruby --version)"')[1].success? or raise 'Could not create run task'
		//   expect(@app).to have_logged(/RUNNING A TASK: ruby 2\.2\.7/)
		// end
	})
})
