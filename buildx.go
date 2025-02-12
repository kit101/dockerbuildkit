package dockerbuildkit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/dchest/uniuri"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/joho/godotenv"
	resolverconfig "github.com/moby/buildkit/util/resolver/config"
)

const DefaultTagsVariableName = "TAGS"

func (p *Plugin) createBuildxInstance() error {
	if p.Buildx.BuildkitdConfig == "" && len(p.Daemon.Mirrors) > 0 {
		err := generateBuildkitdConfig(p.Daemon.Mirrors, DefaultBuildkitdConfigPath)
		if err != nil {
			return fmt.Errorf("generate buildkitd config[%s] error: %v", DefaultBuildkitdConfigPath, err)
		}
		err = printBuildkitdConfig(DefaultBuildkitdConfigPath)
		if err != nil {
			return fmt.Errorf("print buildkitd config[%s] error: %v", DefaultBuildkitdConfigPath, err)
		}
		p.Buildx.BuildkitdConfig = DefaultBuildkitdConfigPath
	}

	var out bytes.Buffer
	err := traceRun(p.commandCreateBuildxInstance(), &out)
	if err != nil {
		return fmt.Errorf("can't create buildx builder instance: %v", err)
	}
	p.builder.name = strutil.Trim(out.String())
	fmt.Printf("[info] builder name: %s\n", p.builder.name)
	return err
}

func (p *Plugin) destroyBuildxInstance() {
	if p.builder.name != "" {
		//_ = traceRun(exec.Command(dockerExe, "buildx", "du", "--builder", p.builder.name), os.Stdout)
		_ = traceRun(exec.Command(dockerExe, "buildx", "prune", "-f", "-a", "--builder", p.builder.name), os.Stdout)
		_ = traceRun(exec.Command(dockerExe, "buildx", "rm", p.builder.name), os.Stdout)
	}
}

func (p *Plugin) commandCreateBuildxInstance() *exec.Cmd {
	args := []string{
		"buildx",
		"create",
		"--driver", "docker-container",
		"--use",
	}
	if p.Build.Platform != "" {
		args = append(args, "--platform", p.Build.Platform)
	}

	// --buildkit-config
	if p.Buildx.BuildkitdConfig != "" {
		args = append(args, "--buildkitd-config", p.Buildx.BuildkitdConfig)
	}
	// --buildkitd-flags
	if p.Buildx.BuildkitdFlags != "" {
		args = append(args, "--buildkitd-flags", p.Buildx.BuildkitdFlags)
	}

	// driver-opt
	var driverOpts []string
	if p.Buildx.NoDefaultNetwork {
		driverOpts = p.Buildx.DriverOpts
	} else {
		for _, opt := range p.Buildx.DriverOpts {
			if strings.HasPrefix(opt, "network=") {
				fmt.Printf("[warning] using default network mode: network=host")
			} else {
				driverOpts = append(driverOpts, opt)
			}
		}
		driverOpts = append(driverOpts, "network=host")
	}
	for _, opt := range driverOpts {
		args = append(args, "--driver-opt", opt)
	}

	// extra options
	args = append(args, p.Buildx.ExtraOptions...)

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

func printBuildkitdConfig(path string) error {
	fmt.Printf("[info] %s: \n", path)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(os.Stdout, f)
	if err != nil {
		return err
	}
	return nil
}

func (p *Plugin) doBake() error {
	metadataFilePath := "/tmp/" + strings.ToLower(uniuri.New()) + "-metadata.json"
	cmds := p.commandPreBuild()

	err := p.Bake.loadEnvfile()
	if err != nil {
		return errors.Join(fmt.Errorf("bake before load envfile"), err)
	}
	variables, err := p.Bake.loadVariables()
	if err != nil {
		return err
	}
	variables = p.Bake.loadTagsVariable(p.Build, variables)

	cmds = append(cmds, p.Bake.commandBakePrint(variables))
	cmds = append(cmds, p.Bake.commandBakePush(variables, metadataFilePath, p.Dryrun))

	for _, cmd := range cmds {
		err := traceRun(cmd, os.Stdout)
		if err != nil {
			return err
		}
	}
	p.Bake.printMetadataFile(metadataFilePath)
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

// loadTagsVariable 处理tags变量, variable > env > tags.auto/tags
func (b Bake) loadTagsVariable(build Build, variables []string) []string {
	for _, v := range b.Variables {
		if strings.HasPrefix(v, b.TagsVariableName+"=") {
			fmt.Printf("[info] tags from variable: %s\n", v)
			return variables
		}
	}
	if tags := os.Getenv(b.TagsVariableName); tags != "" {
		fmt.Printf("[info] tags from env: %s\n", tags)
		return variables
	}
	tags := fmt.Sprintf("%s=%s", b.TagsVariableName, strings.Join(build.Tags, ","))
	fmt.Printf("[info] tags from build.Tags: %s\n", tags)
	return append(variables, tags)
}

func (b Bake) commandBakePrint(variables []string) *exec.Cmd {
	args := []string{"buildx", "bake", "--print"}
	args = b.handleBakeParameters(args)
	cmd := exec.Command(dockerExe, args...)
	cmd.Env = append(cmd.Environ(), variables...)
	return cmd
}

func (b Bake) commandBakePush(variables []string, metadataFilePath string, dryRun bool) *exec.Cmd {
	args := []string{"buildx", "bake"}
	if !dryRun {
		args = append(args, "--push")
	}
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
	for _, t := range b.Targets {
		args = append(args, t)
	}
	return args
}

func (b Bake) printMetadataFile(metadataFilePath string) {
	if !fileutil.IsExist(metadataFilePath) {
		fmt.Printf("[warning] file `%s` is not exists", metadataFilePath)
		return
	}
	metadata, err := fileutil.ReadFileToString(metadataFilePath)
	if err != nil {
		fmt.Printf("[warning] read file `%s` to string error, %v", metadataFilePath, err)
		return
	}
	fmt.Printf("[info] metadata.json:\n%s\n", metadata)
}
