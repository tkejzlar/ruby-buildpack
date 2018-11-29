package integration_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Rails 5.1 (Webpack/Yarn) App", func() {
	var app *cutlass.App
	//AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "rails51_with_precompiled_assets"))
		app.SetEnv("BP_DEBUG", "1")
	})

	It("Installs node6 and runs", func() {
		// cause assets to compile 4 times
		assetPath := filepath.Join(app.Path, "app", "assets", "javascripts", "javaScriptAsset.js")
		var randString string
		for i := 0; i < 3; i++ {
			fmt.Println("pushing app and confirming")
			randString = cutlass.RandStringRunes(20)
			Expect(ioutil.WriteFile(assetPath, []byte(fmt.Sprintf("console.log(%s);", randString)), 0666)).To(Succeed())
			//Expect(app.PushNoStart()).To(Succeed())
			PushAppAndConfirm(app)

		}
		randString = cutlass.RandStringRunes(20)
		Expect(ioutil.WriteFile(assetPath, []byte(fmt.Sprintf("console.log(%salskjfghkljdsh)", randString)), 0666)).To(Succeed())
		PushAppAndConfirm(app)

		Expect(app.GetBody("/")).To(ContainSubstring("javaScriptAsset Count 3"))
		Eventually(app.Stdout.String).Should(ContainSubstring(`Started GET "/" for`))

		//By("Make sure supply does not change BuildDir", func() {
		//	Expect(app).To(HaveUnchangedAppdir("BuildDir Checksum Before Supply", "BuildDir Checksum After Supply"))
		//})
		//
		//By("Make sure binstubs work", func() {
		//	command := exec.Command("cf", "ssh", app.Name, "-c", "/tmp/lifecycle/launcher /home/vcap/app 'rails about' ''")
		//	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		//	Expect(err).ToNot(HaveOccurred())
		//	Eventually(session, 10, 0.25).Should(gexec.Exit(0))
		//})
	})
})
