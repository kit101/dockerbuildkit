//go:build windows
// +build windows

package dockerbuildkit

import (
	"os"
	"os/exec"
	"path/filepath"
)

const dockerExe = "docker.exe"
const dockerdExe = "dockerd"
const dockerHome = "/root/.docker/"
const BuildkitdHomeEnvName = "BUILDKITD_HOME"
const DefaultBuildkitdConfigPath = "/run/buildkit/buildkitd.toml"

func buildkitdCachePath() string {
	return filepath.Join(os.Getenv(BuildkitdHomeEnvName), "cache/%s")
}

func (p Plugin) startDaemon() {
}

// helper function to create the docker daemon command.
func commandDaemon(daemon Daemon) *exec.Cmd {
	return exec.Command(dockerdExe, []string{}...)
}
