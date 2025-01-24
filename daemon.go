//go:build !windows
// +build !windows

package dockerbuildkit

import (
	"io"
	"os"
	"path/filepath"
)

const dockerExe = "/usr/local/bin/docker"
const dockerdExe = "/usr/local/bin/dockerd"
const dockerHome = "/root/.docker/"
const BuildkitdHomeEnvName = "BUILDKITD_HOME"
const DefaultBuildkitdConfigPath = "/run/buildkit/buildkitd.toml"

func buildkitdCachePath() string {
	return filepath.Join(os.Getenv(BuildkitdHomeEnvName), "cache/%s")
}

func (p Plugin) startDaemon() {
	cmd := commandDaemon(p.Daemon)
	if p.Daemon.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
	go func() {
		trace(cmd)
		cmd.Run()
	}()
}
