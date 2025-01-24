package dockerbuildkit

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/dchest/uniuri"
	"github.com/joho/godotenv"
	resolverconfig "github.com/moby/buildkit/util/resolver/config"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (p Plugin) createBuildxInstance() error {
	if p.Buildx.BuildkitdConfig == "" && len(p.Daemon.Mirrors) > 0 {
		err := generateBuildkitdConfig(p.Daemon.Mirrors, DefaultBuildkitdConfigPath)
		if err != nil {
			return fmt.Errorf("generate buildkitd config[%s] error: %v", DefaultBuildkitdConfigPath, err)
		}
		err = traceRun(commandCatBuildkitdConfig(DefaultBuildkitdConfigPath), os.Stdout)
		if err != nil {
			return fmt.Errorf("print buildkitd config[%s] error: %v", DefaultBuildkitdConfigPath, err)
		}
		p.Buildx.BuildkitdConfig = DefaultBuildkitdConfigPath
	}
	err := traceRun(p.commandCreateBuildxInstance(), io.Discard)
	if err != nil {
		return fmt.Errorf("can't create buildx builder instance: %v", err)
	}
	return err
}

func (p Plugin) commandCreateBuildxInstance() *exec.Cmd {
	//buildx create --driver docker-container --use --platform linux/amd64,linux/arm64 --buildkitd-config xxx
	args := []string{
		"buildx",
		"create",
		"--driver", "docker-container",
		"--use",
	}
	if p.Build.Platform != "" {
		args = append(args, "--platform", p.Build.Platform)
	}
	if p.Buildx.DriverOptImage != "" {
		args = append(args, "--driver-opt", "image="+p.Buildx.DriverOptImage)
	}
	if p.Buildx.BuildkitdConfig != "" {
		args = append(args, "--buildkitd-config", p.Buildx.BuildkitdConfig)
	}
	if p.Buildx.Params != "" {
		args = append(args, p.Buildx.Params)
	}
	return exec.Command(dockerExe, args...)
}

func commandInspectBuildxInstance() *exec.Cmd {
	return exec.Command(dockerExe, "buildx", "inspect")
}

func generateBuildkitdConfig(mirrors []string, path string) error {
	plainHttp := false
	for i := range mirrors {
		m := mirrors[i]
		if strings.HasPrefix(m, "http://") {
			plainHttp = true
		}
		m, _ = strings.CutPrefix(m, "https://")
		m, _ = strings.CutPrefix(m, "http://")
		mirrors[i] = m
	}
	c := map[string]interface{}{
		"registry": map[string]resolverconfig.RegistryConfig{
			"docker.io": {
				Mirrors:   mirrors,
				PlainHTTP: &plainHttp,
				Insecure:  &plainHttp,
			},
		},
	}
	tomlBytes, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0644)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, tomlBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func commandCatBuildkitdConfig(path string) *exec.Cmd {
	return exec.Command("cat", path)
}

func (p Plugin) doBake() error {
	metadataFilePath := "/tmp/" + strings.ToLower(uniuri.New()) + "-metadata.json"
	cmds := p.preBuild()

	err := p.Bake.loadEnvfile()
	if err != nil {
		return errors.Join(fmt.Errorf("bake before load envfile"), err)
	}
	variables, err := p.Bake.loadVariables()
	if err != nil {
		return err
	}

	cmds = append(cmds, p.Bake.commandBakePrint(variables))
	cmds = append(cmds, p.Bake.commandBakePush(variables, metadataFilePath))

	for _, cmd := range cmds {
		err := traceRun(cmd, os.Stdout)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b Bake) loadEnvfile() error {
	if b.Envfile != "" {
		return godotenv.Load(b.Envfile)
	}
	return nil
}

func (b Bake) loadVariables() ([]string, error) {
	if len(b.Variables) <= 0 {
		return nil, nil
	}
	var vars []string
	for _, s := range b.Variables {
		partLen := len(strings.Split(s, "="))
		if !(partLen >= 1 && partLen <= 2) {
			return nil, fmt.Errorf("variable [%s] is incorrect. YOUR_VARIABLE_NAME=YOUR_VARIABLE_VALUE", s)
		}
		vars = append(vars, s)
	}
	return vars, nil
}

func (b Bake) commandBakePrint(variables []string) *exec.Cmd {
	args := []string{"buildx", "bake", "--print"}
	args = b.handleBakeParameters(args)
	cmd := exec.Command(dockerExe, args...)
	cmd.Env = append(cmd.Environ(), variables...)
	return cmd
}

func (b Bake) commandBakePush(variables []string, metadataFilePath string) *exec.Cmd {
	args := []string{"buildx", "bake", "--push"}
	args = b.handleBakeParameters(args)
	args = append(args, "--metadata-file", metadataFilePath)
	cmd := exec.Command(dockerExe, args...)
	cmd.Env = append(cmd.Environ(), variables...)
	return cmd
}

func (b Bake) handleBakeParameters(args []string) []string {
	for _, f := range b.Files {
		args = append(args, "--file", f)
	}
	for _, s := range b.Sets {
		args = append(args, "--set", s)
	}
	if b.Provenance != "" {
		args = append(args, "--provenance", b.Provenance)
	}
	if b.Sbom != "" {
		args = append(args, "--sbom", b.Sbom)
	}
	return args
}
