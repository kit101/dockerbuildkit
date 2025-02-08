package main

import (
	"os"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/drone-plugins/drone-plugin-lib/drone"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/kit101/dockerbuildkit"
)

var (
	version = "unknown"
)

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	app := cli.NewApp()
	app.Name = "dockerbuildkit"
	app.Usage = "dockerbuildkit"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "dry-run",
			Usage:  "dry run disables docker push",
			EnvVar: "PLUGIN_DRY_RUN",
		},
		cli.StringFlag{
			Name:   "remote.url",
			Usage:  "git remote url",
			EnvVar: "DRONE_REMOTE_URL",
		},
		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
			Value:  "00000000",
		},
		cli.StringFlag{
			Name:   "commit.ref",
			Usage:  "git commit ref",
			EnvVar: "DRONE_COMMIT_REF",
		},

		// buildx
		cli.StringFlag{
			Name:   "buildx.buildkitd-config",
			Usage:  "buildx buildkitd-config. docker buildx create --buildkitd-config {}. (default: /etc/buildkitd/buildkitd.toml)",
			EnvVar: "DRONE_BUILDX_BUILDKITD_CONFIG",
		},
		cli.StringFlag{
			Name:   "buildx.driver-opt.image",
			Usage:  "buildx driver-opt image. docker buildx create --driver-opt image={}",
			EnvVar: "PLUGIN_BUILDX_DRIVER_OPT_IMAGE",
		},
		cli.StringFlag{
			Name:   "buildx.driver-opt.network",
			Usage:  "buildx driver-opt network. docker buildx create --driver-opt network={}",
			EnvVar: "PLUGIN_BUILDX_DRIVER_OPT_NETWORK",
			//Value: "host",
		},
		cli.StringSliceFlag{
			Name:   "buildx.args",
			Usage:  "buildx args. docker buildx create {}",
			EnvVar: "PLUGIN_BUILDX_ARGS",
		},

		// daemon
		cli.StringFlag{
			Name:   "daemon.mirror",
			Usage:  "This flag is deprecated. Please use '--daemon.mirrors'",
			EnvVar: "PLUGIN_MIRROR,DOCKER_PLUGIN_MIRROR",
		},
		cli.StringSliceFlag{
			Name:   "daemon.mirrors",
			Usage:  "multiple docker daemon registry mirrors, separated by commas",
			EnvVar: "PLUGIN_MIRRORS,DOCKER_PLUGIN_MIRRORS",
		},
		cli.StringFlag{
			Name:   "daemon.storage-driver",
			Usage:  "docker daemon storage driver",
			EnvVar: "PLUGIN_STORAGE_DRIVER",
		},
		cli.StringFlag{
			Name:   "daemon.storage-path",
			Usage:  "docker daemon storage path",
			Value:  "/var/lib/docker",
			EnvVar: "PLUGIN_STORAGE_PATH",
		},
		cli.StringFlag{
			Name:   "daemon.bip",
			Usage:  "docker daemon bride ip address",
			EnvVar: "PLUGIN_BIP",
		},
		cli.StringFlag{
			Name:   "daemon.mtu",
			Usage:  "docker daemon custom mtu setting",
			EnvVar: "PLUGIN_MTU",
		},
		cli.StringSliceFlag{
			Name:   "daemon.dns",
			Usage:  "docker daemon dns server",
			EnvVar: "PLUGIN_CUSTOM_DNS",
		},
		cli.StringSliceFlag{
			Name:   "daemon.dns-search",
			Usage:  "docker daemon dns search domains",
			EnvVar: "PLUGIN_CUSTOM_DNS_SEARCH",
		},
		cli.BoolFlag{
			Name:   "daemon.insecure",
			Usage:  "docker daemon allows insecure registries",
			EnvVar: "PLUGIN_INSECURE",
		},
		cli.BoolFlag{
			Name:   "daemon.ipv6",
			Usage:  "docker daemon IPv6 networking",
			EnvVar: "PLUGIN_IPV6",
		},
		cli.BoolFlag{
			Name:   "daemon.experimental",
			Usage:  "docker daemon Experimental mode",
			EnvVar: "PLUGIN_EXPERIMENTAL",
		},
		cli.BoolFlag{
			Name:   "daemon.debug",
			Usage:  "docker daemon executes in debug mode",
			EnvVar: "PLUGIN_DEBUG,DOCKER_LAUNCH_DEBUG",
		},
		cli.BoolFlag{
			Name:   "daemon.off",
			Usage:  "don't start the docker daemon",
			EnvVar: "PLUGIN_DAEMON_OFF",
		},

		// build
		cli.StringFlag{
			Name:   "dockerfile",
			Usage:  "build dockerfile",
			Value:  "Dockerfile",
			EnvVar: "PLUGIN_DOCKERFILE",
		},
		cli.StringFlag{
			Name:   "context",
			Usage:  "build context",
			Value:  ".",
			EnvVar: "PLUGIN_CONTEXT",
		},
		cli.StringSliceFlag{
			Name:     "tags",
			Usage:    "build tags",
			Value:    &cli.StringSlice{"latest"},
			EnvVar:   "PLUGIN_TAG,PLUGIN_TAGS",
			FilePath: ".tags",
		},
		cli.BoolFlag{
			Name:   "tags.auto",
			Usage:  "default build tags",
			EnvVar: "PLUGIN_DEFAULT_TAGS,PLUGIN_AUTO_TAG",
		},
		cli.StringFlag{
			Name:   "tags.suffix",
			Usage:  "default build tags with suffix",
			EnvVar: "PLUGIN_DEFAULT_SUFFIX,PLUGIN_AUTO_TAG_SUFFIX",
		},
		cli.StringSliceFlag{
			Name:   "args",
			Usage:  "build args",
			EnvVar: "PLUGIN_BUILD_ARGS",
		},
		cli.StringSliceFlag{
			Name:   "args-from-env",
			Usage:  "build args",
			EnvVar: "PLUGIN_BUILD_ARGS_FROM_ENV",
		},
		cli.GenericFlag{
			Name:   "args-new",
			Usage:  "build args new",
			EnvVar: "PLUGIN_BUILD_ARGS_NEW",
			Value:  new(CustomStringSliceFlag),
		},
		cli.BoolFlag{
			Name:   "plugin-multiple-build-agrs",
			Usage:  "plugin multiple build agrs",
			EnvVar: "PLUGIN_MULTIPLE_BUILD_ARGS",
		},
		cli.BoolFlag{
			Name:   "quiet",
			Usage:  "quiet docker build",
			EnvVar: "PLUGIN_QUIET",
		},
		cli.StringFlag{
			Name:   "target",
			Usage:  "build target",
			EnvVar: "PLUGIN_TARGET",
		},
		cli.StringSliceFlag{
			Name:   "cache-from",
			Usage:  "images to consider as cache sources",
			EnvVar: "PLUGIN_CACHE_FROM",
		},
		cli.BoolFlag{
			Name:   "squash",
			Usage:  "squash the layers at build time",
			EnvVar: "PLUGIN_SQUASH",
		},
		cli.BoolTFlag{
			Name:   "pull-image",
			Usage:  "force pull base image at build time",
			EnvVar: "PLUGIN_PULL_IMAGE",
		},
		cli.BoolFlag{
			Name:   "compress",
			Usage:  "compress the build context using gzip",
			EnvVar: "PLUGIN_COMPRESS",
		},
		cli.StringFlag{
			Name:   "repo",
			Usage:  "docker repository",
			EnvVar: "PLUGIN_REPO",
		},
		cli.StringSliceFlag{
			Name:   "custom-labels",
			Usage:  "additional k=v labels",
			EnvVar: "PLUGIN_CUSTOM_LABELS",
		},
		cli.StringSliceFlag{
			Name:   "label-schema",
			Usage:  "label-schema labels",
			EnvVar: "PLUGIN_LABEL_SCHEMA",
		},
		cli.BoolTFlag{
			Name:   "auto-label",
			Usage:  "auto-label true|false",
			EnvVar: "PLUGIN_AUTO_LABEL",
		},
		cli.StringFlag{
			Name:   "link",
			Usage:  "link https://example.com/org/repo-name",
			EnvVar: "PLUGIN_REPO_LINK,DRONE_REPO_LINK",
		},

		// bake
		cli.StringSliceFlag{Name: "bake.file", EnvVar: "PLUGIN_BAKE_FILE", Usage: "Build definition file"},
		cli.StringSliceFlag{Name: "bake.target", EnvVar: "PLUGIN_BAKE_TARGET", Usage: "A target in a Bake file represents a build invocation"},
		cli.StringFlag{Name: "bake.provenance", EnvVar: "PLUGIN_BAKE_PROVENANCE", Usage: "Shorthand for \"--set=*.attest=type=provenance\""},
		cli.StringFlag{Name: "bake.sbom", EnvVar: "PLUGIN_BAKE_SBOM", Usage: "Shorthand for \"--set=*.attest=type=sbom\""},
		cli.StringSliceFlag{Name: "bake.set", EnvVar: "PLUGIN_BAKE_SET", Usage: "Override target value (e.g., \"targetpattern.key=value\")"},
		cli.StringFlag{Name: "bake.envfile", EnvVar: "PLUGIN_BAKE_ENVFILE", Usage: "will 'source ${bake.envfile}'"},
		cli.StringSliceFlag{Name: "bake.variable", EnvVar: "PLUGIN_BAKE_VARIABLE", Usage: "load env"},
		cli.StringFlag{Name: "bake.tags-variable-name", EnvVar: "PLUGIN_BAKE_TAGS_NAME", Usage: "Tags variable name generated after using tags or tags.auto. Default \"TAGS\"", Value: dockerbuildkit.DefaultTagsVariableName},

		// docker
		cli.StringFlag{
			Name:   "docker.registry",
			Usage:  "docker registry",
			Value:  "https://index.docker.io/v1/",
			EnvVar: "PLUGIN_REGISTRY,DOCKER_REGISTRY",
		},
		cli.StringFlag{
			Name:   "docker.username",
			Usage:  "docker username",
			EnvVar: "PLUGIN_USERNAME,DOCKER_USERNAME",
		},
		cli.StringFlag{
			Name:   "docker.password",
			Usage:  "docker password",
			EnvVar: "PLUGIN_PASSWORD,DOCKER_PASSWORD",
		},
		cli.StringFlag{
			Name:   "docker.baseimageusername",
			Usage:  "Docker username for base image registry",
			EnvVar: "PLUGIN_DOCKER_USERNAME,PLUGIN_BASE_IMAGE_USERNAME,DOCKER_BASE_IMAGE_USERNAME",
		},
		cli.StringFlag{
			Name:   "docker.baseimagepassword",
			Usage:  "Docker password for base image registry",
			EnvVar: "PLUGIN_DOCKER_PASSWORD,PLUGIN_BASE_IMAGE_PASSWORD,DOCKER_BASE_IMAGE_PASSWORD",
		},
		cli.StringFlag{
			Name:   "docker.baseimageregistry",
			Usage:  "Docker registry for base image registry",
			EnvVar: "PLUGIN_DOCKER_REGISTRY,PLUGIN_BASE_IMAGE_REGISTRY,DOCKER_BASE_IMAGE_REGISTRY",
		},
		cli.StringFlag{
			Name:   "docker.email",
			Usage:  "docker email",
			EnvVar: "PLUGIN_EMAIL,DOCKER_EMAIL",
		},
		cli.StringFlag{
			Name:   "docker.config",
			Usage:  "docker json dockerconfig content",
			EnvVar: "PLUGIN_CONFIG,DOCKER_PLUGIN_CONFIG",
		},
		cli.BoolTFlag{
			Name:   "docker.purge",
			Usage:  "docker should cleanup images",
			EnvVar: "PLUGIN_PURGE",
		},
		cli.StringFlag{
			Name:   "repo.branch",
			Usage:  "repository default branch",
			EnvVar: "DRONE_REPO_BRANCH",
		},
		cli.BoolFlag{
			Name:   "no-cache",
			Usage:  "do not use cached intermediate containers",
			EnvVar: "PLUGIN_NO_CACHE",
		},
		cli.StringSliceFlag{
			Name:   "add-host",
			Usage:  "additional host:IP mapping",
			EnvVar: "PLUGIN_ADD_HOST",
		},
		cli.StringFlag{
			Name:   "secret",
			Usage:  "secret key value pair eg id=MYSECRET",
			EnvVar: "PLUGIN_SECRET",
		},
		cli.StringSliceFlag{
			Name:   "secrets-from-env",
			Usage:  "secret key value pair eg secret_name=secret",
			EnvVar: "PLUGIN_SECRETS_FROM_ENV",
		},
		cli.StringSliceFlag{
			Name:   "secrets-from-file",
			Usage:  "secret key value pairs eg secret_name=/path/to/secret",
			EnvVar: "PLUGIN_SECRETS_FROM_FILE",
		},
		cli.StringFlag{
			Name:   "drone-card-path",
			Usage:  "card path location to write to",
			EnvVar: "DRONE_CARD_PATH",
		},
		cli.StringFlag{
			Name:   "platform",
			Usage:  "platform value to pass to docker",
			EnvVar: "PLUGIN_PLATFORM",
		},
		cli.StringFlag{
			Name:   "ssh-agent-key",
			Usage:  "ssh agent key to use",
			EnvVar: "PLUGIN_SSH_AGENT_KEY",
		},
		cli.StringFlag{
			Name:   "artifact-file",
			Usage:  "Artifact file location that will be generated by the plugin. This file will include information of docker images that are uploaded by the plugin.",
			EnvVar: "PLUGIN_ARTIFACT_FILE",
		},
		cli.StringFlag{
			Name:   "registry-type",
			Usage:  "registry type",
			EnvVar: "PLUGIN_REGISTRY_TYPE",
		},
		cli.StringFlag{
			Name:   "access-token",
			Usage:  "access token",
			EnvVar: "ACCESS_TOKEN",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	registryType := drone.Docker
	if c.String("registry-type") != "" {
		registryType = drone.RegistryType(c.String("registry-type"))
	}

	plugin := dockerbuildkit.Plugin{
		Dryrun:  c.Bool("dry-run"),
		Cleanup: c.BoolT("docker.purge"),
		Login: dockerbuildkit.Login{
			Registry:    c.String("docker.registry"),
			Username:    c.String("docker.username"),
			Password:    c.String("docker.password"),
			Email:       c.String("docker.email"),
			Config:      c.String("docker.config"),
			AccessToken: c.String("access-token"),
		},
		CardPath:     c.String("drone-card-path"),
		ArtifactFile: c.String("artifact-file"),
		Build: dockerbuildkit.Build{
			Remote:      c.String("remote.url"),
			Name:        c.String("commit.sha"),
			TempTag:     generateTempTag(),
			Dockerfile:  c.String("dockerfile"),
			Context:     c.String("context"),
			Tags:        c.StringSlice("tags"),
			Args:        c.StringSlice("args"),
			ArgsEnv:     c.StringSlice("args-from-env"),
			Target:      c.String("target"),
			Squash:      c.Bool("squash"),
			Pull:        c.BoolT("pull-image"),
			CacheFrom:   c.StringSlice("cache-from"),
			Compress:    c.Bool("compress"),
			Repo:        c.String("repo"),
			Labels:      c.StringSlice("custom-labels"),
			LabelSchema: c.StringSlice("label-schema"),
			AutoLabel:   c.BoolT("auto-label"),
			Link:        c.String("link"),
			NoCache:     c.Bool("no-cache"),
			Secret:      c.String("secret"),
			SecretEnvs:  c.StringSlice("secrets-from-env"),
			SecretFiles: c.StringSlice("secrets-from-file"),
			AddHost:     c.StringSlice("add-host"),
			Quiet:       c.Bool("quiet"),
			Platform:    c.String("platform"),
			SSHAgentKey: c.String("ssh-agent-key"),
		},
		Bake: dockerbuildkit.Bake{
			Files:            c.StringSlice("bake.file"),
			Targets:          c.StringSlice("bake.target"),
			Provenance:       c.String("bake.provenance"),
			Sbom:             c.String("bake.sbom"),
			Sets:             c.StringSlice("bake.set"),
			Variables:        c.StringSlice("bake.variable"),
			Envfile:          c.String("bake.envfile"),
			TagsVariableName: c.String("bake.tags-variable-name"),
		},
		Buildx: dockerbuildkit.Buildx{
			BuildkitdConfig:  c.String("buildx.buildkitd-config"),
			DriverOptImage:   c.String("buildx.driver-opt.image"),
			DriverOptNetwork: c.String("buildx.driver-opt.network"),
			Args:             c.StringSlice("buildx.args"),
		},
		Daemon: dockerbuildkit.Daemon{
			Registry:      c.String("docker.registry"),
			Mirrors:       mirrors(c),
			StorageDriver: c.String("daemon.storage-driver"),
			StoragePath:   c.String("daemon.storage-path"),
			Insecure:      c.Bool("daemon.insecure"),
			Disabled:      c.Bool("daemon.off"),
			IPv6:          c.Bool("daemon.ipv6"),
			Debug:         c.Bool("daemon.debug"),
			Bip:           c.String("daemon.bip"),
			DNS:           c.StringSlice("daemon.dns"),
			DNSSearch:     c.StringSlice("daemon.dns-search"),
			MTU:           c.String("daemon.mtu"),
			Experimental:  c.Bool("daemon.experimental"),
			RegistryType:  registryType,
		},
		BaseImageRegistry: c.String("docker.baseimageregistry"),
		BaseImageUsername: c.String("docker.baseimageusername"),
		BaseImagePassword: c.String("docker.baseimagepassword"),
	}

	if c.Bool("tags.auto") {
		if dockerbuildkit.UseDefaultTag( // return true if tag event or default branch
			c.String("commit.ref"),
			c.String("repo.branch"),
		) {
			tag, err := dockerbuildkit.DefaultTagSuffix(
				c.String("commit.ref"),
				c.String("tags.suffix"),
			)
			if err != nil {
				logrus.Printf("cannot build docker image for %s, invalid semantic version", c.String("commit.ref"))
				return err
			}
			plugin.Build.Tags = tag
		} else {
			logrus.Printf("skipping automated docker build for %s", c.String("commit.ref"))
			return nil
		}
	}

	err := plugin.Exec()
	return err
}

func generateTempTag() string {
	return strings.ToLower(uniuri.New())
}

func mirrors(c *cli.Context) []string {
	m := c.StringSlice("daemon.mirrors")
	if c.String("daemon.mirror") != "" {
		m = append(m, c.String("daemon.mirror"))
	}
	return slice.Unique(m)
}
