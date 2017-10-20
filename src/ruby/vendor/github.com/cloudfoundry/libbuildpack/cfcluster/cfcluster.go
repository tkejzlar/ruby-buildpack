package cfcluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack/models"
	"github.com/tidwall/gjson"
)

var DefaultStdoutStderr io.Writer = nil

type cfConfig struct {
	SpaceFields struct {
		GUID string
	}
}
type cfApps struct {
	Resources []struct {
		Metadata struct {
			GUID string `json:"guid"`
		} `json:"metadata"`
	} `json:"resources"`
}
type cfInstance struct {
	State string `json:"state"`
}

type cfCluster struct{}

func New() *cfCluster {
	return &cfCluster{}
}

func (c *cfCluster) ApiVersion() (string, error) {
	cmd := exec.Command("cf", "curl", "/v2/info")
	cmd.Stderr = DefaultStdoutStderr
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	var info struct {
		ApiVersion string `json:"api_version"`
	}
	if err := json.Unmarshal(bytes, &info); err != nil {
		return "", err
	}
	return info.ApiVersion, nil
}

func (c *cfCluster) DeleteOrphanedRoutes() error {
	command := exec.Command("cf", "delete-orphaned-routes", "-f")
	command.Stdout = DefaultStdoutStderr
	command.Stderr = DefaultStdoutStderr
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

func (c *cfCluster) CreateBuildpack(name, file string) error {
	command := exec.Command("cf", "create-buildpack", name, file, "100", "--enable")
	if _, err := command.CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func (c *cfCluster) UpdateBuildpack(name, file string) error {
	command := exec.Command("cf", "update-buildpack", name, "-p", file, "--enable")
	if data, err := command.CombinedOutput(); err != nil {
		fmt.Println(string(data))
		return err
	}
	return nil
}

func (c *cfCluster) DeleteBuildpack(name string) error {
	command := exec.Command("cf", "delete-buildpack", "-f", name)
	if data, err := command.CombinedOutput(); err != nil {
		fmt.Println(string(data))
		return err
	}
	return nil
}

func (c *cfCluster) RunTask(name, command string) ([]byte, error) {
	cmd := exec.Command("cf", "run-task", name, command)
	cmd.Stderr = DefaultStdoutStderr
	bytes, err := cmd.Output()
	if err != nil {
		return bytes, err
	}
	return bytes, nil
}

func (c *cfCluster) RestartApp(name string) error {
	command := exec.Command("cf", "restart", name)
	command.Stdout = DefaultStdoutStderr
	command.Stderr = DefaultStdoutStderr
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

func (c *cfCluster) spaceGUID() (string, error) {
	cfHome := os.Getenv("CF_HOME")
	if cfHome == "" {
		cfHome = os.Getenv("HOME")
	}
	bytes, err := ioutil.ReadFile(filepath.Join(cfHome, ".cf", "config.json"))
	if err != nil {
		return "", err
	}
	var config cfConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		return "", err
	}
	return config.SpaceFields.GUID, nil
}

func (c *cfCluster) AppGUID(name string) (string, error) {

	// guid
	//
	cmd := exec.Command("cf", "curl", "/v2/apps?q=space_guid:"+guid+"&q=name:"+name)
	cmd.Stderr = DefaultStdoutStderr
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	var apps cfApps
	if err := json.Unmarshal(bytes, &apps); err != nil {
		return "", err
	}
	if len(apps.Resources) != 1 {
		return "", fmt.Errorf("Expected one app, found %d", len(apps.Resources))
	}
	return apps.Resources[0].Metadata.GUID, nil
}

func (c *cfCluster) InstanceStates(guid string) ([]string, error) {
	cmd := exec.Command("cf", "curl", "/v2/apps/"+guid+"/instances")
	cmd.Stderr = DefaultStdoutStderr
	bytes, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	var data map[string]cfInstance
	if err := json.Unmarshal(bytes, &data); err != nil {
		return []string{}, err
	}
	var states []string
	for _, value := range data {
		states = append(states, value.State)
	}
	return states, nil
}

func (c *cfCluster) PushApp(a models.App) error {
	args := []string{"push", a.Name(), "--no-start", "-p", a.Path()}
	if a.Stack() != "" {
		args = append(args, "-s", a.Stack())
	}
	var buildpacks = a.Buildpacks()
	if len(buildpacks) == 1 {
		args = append(args, "-b", buildpacks[len(buildpacks)-1])
	}
	if _, err := os.Stat(filepath.Join(a.Path(), "manifest.yml")); err == nil {
		args = append(args, "-f", filepath.Join(a.Path(), "manifest.yml"))
	}
	if a.Memory() != "" {
		args = append(args, "-m", a.Memory())
	}
	if a.Disk() != "" {
		args = append(args, "-k", a.Disk())
	}
	command := exec.Command("cf", args...)
	command.Stdout = DefaultStdoutStderr
	command.Stderr = DefaultStdoutStderr
	if err := command.Run(); err != nil {
		return err
	}

	for k, v := range a.Env() {
		command := exec.Command("cf", "set-env", a.Name(), k, v)
		command.Stdout = DefaultStdoutStderr
		command.Stderr = DefaultStdoutStderr
		if err := command.Run(); err != nil {
			return err
		}
	}

	logCmd := exec.Command("cf", "logs", a.Name())
	logCmd.Stderr = DefaultStdoutStderr
	stdout := bytes.NewBuffer(nil)
	logCmd.Stdout = stdout
	a.SetStdout(stdout)
	a.SetLogCmd(logCmd)
	if err := logCmd.Start(); err != nil {
		return err
	}

	if len(a.Buildpacks()) > 1 {
		args = []string{"v3-push", a.Name(), "-p", a.Path()}
		for _, buildpack := range a.Buildpacks() {
			args = append(args, "-b", buildpack)
		}
	} else {
		args = []string{"start", a.Name()}
	}
	command = exec.Command("cf", args...)
	command.Stdout = DefaultStdoutStderr
	command.Stderr = DefaultStdoutStderr
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

func (c *cfCluster) GetUrl(a models.App, path string) (string, error) {
	guid, err := a.AppGUID()
	if err != nil {
		return "", err
	}
	cmd := exec.Command("cf", "curl", "/v2/apps/"+guid+"/summary")
	cmd.Stderr = DefaultStdoutStderr
	data, err := cmd.Output()
	if err != nil {
		return "", err
	}
	host := gjson.Get(string(data), "routes.0.host").String()
	domain := gjson.Get(string(data), "routes.0.domain.name").String()
	return fmt.Sprintf("http://%s.%s%s", host, domain, path), nil
}

func (c *cfCluster) Files(appName, path string) ([]string, error) {
	cmd := exec.Command("cf", "ssh", appName, "-c", "find "+path)
	cmd.Stderr = DefaultStdoutStderr
	output, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	return strings.Split(string(output), "\n"), nil
}

func (c *cfCluster) DestroyApp(name string) error {
	command := exec.Command("cf", "delete", "-f", name)
	command.Stdout = DefaultStdoutStderr
	command.Stderr = DefaultStdoutStderr
	return command.Run()
}
