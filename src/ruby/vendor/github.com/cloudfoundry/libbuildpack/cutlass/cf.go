package cutlass

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cloudfoundry/libbuildpack/cfcluster"
	"github.com/cloudfoundry/libbuildpack/models"
)

var DefaultMemory string = ""
var DefaultDisk string = ""
var Cached bool = false
var DefaultStdoutStderr io.Writer = nil

type CfCli interface {
	ApiVersion() (string, error)

	DeleteOrphanedRoutes() error
	CreateBuildpack(name, file string) error
	UpdateBuildpack(name, file string) error
	DeleteBuildpack(name string) error
	RunTask(name, command string) ([]byte, error)
	RestartApp(name string) error
	AppGUID(guid, name string) (string, error)
	InstanceStates(guid string) ([]string, error)
	PushApp(a models.App) error
	GetUrl(a models.App, path string) (string, error)
	Files(appName, path string) ([]string, error)
	DestroyApp(name string) error
}

var cfCli CfCli = cfcluster.New()

type App struct {
	Name       string
	Path       string
	Stack      string
	Buildpacks []string
	Memory     string
	Disk       string
	Stdout     *bytes.Buffer
	appGUID    string
	env        map[string]string
	logCmd     *exec.Cmd
}

func New(fixture string) *App {
	return &App{
		Name:       filepath.Base(fixture) + "-" + RandStringRunes(20),
		Path:       fixture,
		Stack:      "",
		Buildpacks: []string{},
		Memory:     DefaultMemory,
		Disk:       DefaultDisk,
		appGUID:    "",
		env:        map[string]string{},
		logCmd:     nil,
	}
}

func ApiVersion() (string, error) {
	return cfCli.ApiVersion()
}

func DeleteOrphanedRoutes() error {
	return cfCli.DeleteOrphanedRoutes()
}

func DeleteBuildpack(language string) error {
	return cfCli.DeleteBuildpack(fmt.Sprintf("%s_buildpack", language))
}

func UpdateBuildpack(language, file string) error {
	return cfCli.UpdateBuildpack(fmt.Sprintf("%s_buildpack", language), file)
}

func createBuildpack(language, file string) error {
	return cfCli.CreateBuildpack(fmt.Sprintf("%s_buildpack", language), file)
}

func CreateOrUpdateBuildpack(language, file string) error {
	createBuildpack(language, file)
	return UpdateBuildpack(language, file)
}

func (a *App) ConfirmBuildpack(version string) error {
	if !strings.Contains(a.Stdout.String(), fmt.Sprintf("Buildpack version %s\n", version)) {
		var versionLine string
		for _, line := range strings.Split(a.Stdout.String(), "\n") {
			if versionLine == "" && strings.Contains(line, " Buildpack version ") {
				versionLine = line
			}
		}
		return fmt.Errorf("Wrong buildpack version. Expected '%s', but this was logged: %s", version, versionLine)
	}
	return nil
}

func (a *App) RunTask(command string) ([]byte, error) {
	return cfCli.RunTask(a.Name, command)
}

func (a *App) Restart() error {
	return cfCli.RestartApp(a.Name)
}

func (a *App) SetEnv(key, value string) {
	a.env[key] = value
}

func (a *App) AppGUID() (string, error) {
	if a.appGUID != "" {
		return a.appGUID, nil
	}
	a.appGUID, err = cfCli.AppGUID(a.Name)
	return a.appGUID, nil
}

func (a *App) InstanceStates() ([]string, error) {
	if guid, err := a.AppGUID(); err != nil {
		return []string{}, err
	} else {
		return cfCli.InstanceStates(guid)
	}
}

func (a *App) Push() error {
	return cfCli.PushApp(a)
}

func (a *App) GetUrl(path string) (string, error) {
	return cfCli.GetUrl(a, path)
}

func (a *App) Get(path string, headers map[string]string) (string, map[string][]string, error) {
	url, err := a.GetUrl(path)
	if err != nil {
		return "", map[string][]string{}, err
	}
	client := &http.Client{}
	if headers["NoFollow"] == "true" {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		delete(headers, "NoFollow")
	}
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	if headers["user"] != "" && headers["password"] != "" {
		req.SetBasicAuth(headers["user"], headers["password"])
		delete(headers, "user")
		delete(headers, "password")
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", map[string][]string{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", map[string][]string{}, err
	}
	resp.Header["StatusCode"] = []string{strconv.Itoa(resp.StatusCode)}
	return string(data), resp.Header, err
}

func (a *App) GetBody(path string) (string, error) {
	body, _, err := a.Get(path, map[string]string{})
	// TODO: Non 200 ??
	// if !(len(headers["StatusCode"]) == 1 && headers["StatusCode"][0] == "200") {
	// 	return "", fmt.Errorf("non 200 status: %v", headers)
	// }
	return body, err
}

func (a *App) Files(path string) ([]string, error) {
	return cfCli.Files(a.Name, path)
}

func (a *App) Destroy() error {
	if a.logCmd != nil && a.logCmd.Process != nil {
		if err := a.logCmd.Process.Kill(); err != nil {
			return err
		}
	}

	return cfCli.DestroyApp(a.Name)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
