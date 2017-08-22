package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App with windows Gemfile", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "windows"))
		app.SetEnv("BP_DEBUG", "1")
	})

	It("warned the user about Windows line endings for windows Gemfile", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(ContainSubstring("Windows line endings detected in Gemfile. Your app may fail to stage. Please use UNIX line endings."))

		// TODO example for making sure supply does not change BuildDir
		// Expect(app.Stdout.String()).To(ContainSubstring("BuildDir Checksum Before Supply: 5a1298ab0ca3c11678b104bc98d8c730"))
		// Expect(app.Stdout.String()).To(ContainSubstring("BuildDir Checksum After Supply: 5a1298ab0ca3c11678b104bc98d8c730"))
	})
})
