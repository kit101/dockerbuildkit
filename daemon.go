//go:build !windows
// +build !windows

package docker

import (
	"io"
	"os"
)

const dockerExe = "/usr/local/bin/docker"
const dockerdExe = "/usr/local/bin/dockerd"
const dockerHome = "/root/.docker/"
const buildxExe = "/usr/local/bin/buildx"
const builderName = "default_builder"
const buildkitImageTarPath = "/tmp/moby-buildkit-buildx-stable.tar"

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
