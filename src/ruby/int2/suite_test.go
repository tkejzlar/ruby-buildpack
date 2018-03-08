package integration_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cfapi"
	"github.com/cloudfoundry/libbuildpack/cfapi/setup"
	"github.com/cloudfoundry/libbuildpack/cutlass"
)

var bpDir string
var cluster cfapi.Cluster

func init() {
	var err error
	var buildpackName, buildpackFile, buildpackVersion, clusterType string
	flag.StringVar(&buildpackName, "bpName", "ruby_buildpack", "name of buildpack to use")
	flag.StringVar(&buildpackFile, "bpFile", "", "location of buildpack file. (must include version flag)")
	flag.StringVar(&buildpackVersion, "bpVersion", "", "version to use (builds if empty)")
	flag.StringVar(&clusterType, "clusterType", "foundation", "cluster type to run against [foundation,pack,cflocal]")
	flag.BoolVar(&cutlass.Cached, "cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "64M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "64M", "default disk for pushed apps")
	flag.Parse()

	// TODO remove
	clusterType = "pack"

	bpDir, cluster, err = setup.Suite(buildpackName, buildpackFile, buildpackVersion, clusterType)
	if err != nil {
		fmt.Printf("Error in SuiteSetup: %s\n", err)
		os.Exit(1)
	}
}

func By(_ string, f func()) { f() }

func DefaultVersion(name string) (string, error) {
	m := &libbuildpack.Manifest{}
	if err := (&libbuildpack.YAML{}).Load(filepath.Join(bpDir, "manifest.yml"), m); err != nil {
		return "", err
	}
	dep, err := m.DefaultVersion(name)
	if err != nil {
		return "", err
	}
	if dep.Version == "" {
		return "", fmt.Errorf("version was empty")
	}
	return dep.Version, nil
}
