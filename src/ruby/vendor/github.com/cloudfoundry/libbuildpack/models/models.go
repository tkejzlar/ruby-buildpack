package models

import (
	"bytes"
	"os/exec"
)

type App interface {
	Name() string
	Path() string
	Stack() string
	Buildpacks() []string
	Memory() string
	Disk() string
	Stdout() *bytes.Buffer
	AppGUID() (string, error)
	Env() map[string]string

	SetStdout(stdout *bytes.Buffer)
	SetLogCmd(logCmd *exec.Cmd)
}
