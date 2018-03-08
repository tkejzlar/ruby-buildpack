package models

type Cluster interface {
	UploadBuildpack(name, version, file string) error
	DeleteBuildpack(name string) error
	NewApp(bpDir, fixtureName string) (App, error)

	HasMultiBuildpack() bool
	HasTask() bool
}

type App interface {
	Name() string
	SetEnv(key, value string)
	Buildpacks([]string)
	ConfirmBuildpack(version string) error
	Push() error
	PushAndConfirm() error
	Destroy() error
	// GetUrl(path string) (string, error)
	// Get(path string, headers map[string]string) (string, map[string][]string, error)
	GetBody(path string) (string, error)
	Log() string
	ResetLog()
	Files(string) ([]string, error)
	RunTask(command string) ([]byte, error)
}
