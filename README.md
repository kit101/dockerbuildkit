# drone-docker

[![Build Status](http://cloud.drone.io/api/badges/drone-plugins/drone-docker/status.svg)](http://cloud.drone.io/drone-plugins/drone-docker)
[![Gitter chat](https://badges.gitter.im/drone/drone.png)](https://gitter.im/drone/drone)
[![Join the discussion at https://discourse.drone.io](https://img.shields.io/badge/discourse-forum-orange.svg)](https://discourse.drone.io)
[![Drone questions at https://stackoverflow.com](https://img.shields.io/badge/drone-stackoverflow-orange.svg)](https://stackoverflow.com/questions/tagged/drone.io)
[![](https://images.microbadger.com/badges/image/plugins/docker.svg)](https://microbadger.com/images/plugins/docker "Get your own image badge on microbadger.com")
[![Go Doc](https://godoc.org/github.com/drone-plugins/drone-docker?status.svg)](http://godoc.org/github.com/drone-plugins/drone-docker)
[![Go Report](https://goreportcard.com/badge/github.com/drone-plugins/drone-docker)](https://goreportcard.com/report/github.com/drone-plugins/drone-docker)

Drone plugin uses Docker-in-Docker to build and publish Docker images to a container registry. For the usage information and a listing of the available options please take a look at [the docs](http://plugins.drone.io/drone-plugins/drone-docker/).

### Git Leaks

Run the following script to install git-leaks support to this repo.
```
chmod +x ./git-hooks/install.sh
./git-hooks/install.sh
```

## Build

Build the binaries with the following commands:

```console
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GO111MODULE=on

go build -v -a -tags netgo -o release/linux/amd64/dockerbuildkit ./cmd
```

## Docker

Build the Docker images with the following commands:

```console
docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file docker/Dockerfile --tag kit101z/dockerbuildkit .

```

## Usage

> Notice: Be aware that the Docker plugin currently requires privileged capabilities, otherwise the integrated Docker daemon is not able to start.

### Help info

```console
/src # dockerbuildkit --help
NAME:
   docker plugin - docker plugin

USAGE:
   dockerbuildkit [global options] command [command options] [arguments...]

VERSION:
   d090da7975ecb0cd925aa4778b08d5e47d0c9e89

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --dry-run                         dry run disables docker push [$PLUGIN_DRY_RUN]
   --remote.url value                git remote url [$DRONE_REMOTE_URL]
   --commit.sha value                git commit sha (default: "00000000") [$DRONE_COMMIT_SHA]
   --commit.ref value                git commit ref [$DRONE_COMMIT_REF]
   --buildx.buildkitd-config value   buildx buildkitd-config. docker buildx create --buildkitd-config {}. (default: /etc/buildkitd/buildkitd.toml) [$DRONE_BUILDX_BUILDKITD_CONFIG]
   --buildx.driver-opt.image value   buildx driver-opt image. docker buildx create --driver-opt image={} [$DRONE_BUILDX_DRIVER_OPT_IMAGE]
   --buildx.params value             buildx params. docker buildx create {} [$DRONE_BUILDX_PARAMS]
   --daemon.mirror value             This flag is deprecated. Please use '--daemon.mirror' [$PLUGIN_MIRROR, $DOCKER_PLUGIN_MIRROR]
   --daemon.mirrors value            multiple docker daemon registry mirrors, separated by commas [$PLUGIN_MIRRORS, $DOCKER_PLUGIN_MIRRORS]
   --daemon.storage-driver value     docker daemon storage driver [$PLUGIN_STORAGE_DRIVER]
   --daemon.storage-path value       docker daemon storage path (default: "/var/lib/docker") [$PLUGIN_STORAGE_PATH]
   --daemon.bip value                docker daemon bride ip address [$PLUGIN_BIP]
   --daemon.mtu value                docker daemon custom mtu setting [$PLUGIN_MTU]
   --daemon.dns value                docker daemon dns server [$PLUGIN_CUSTOM_DNS]
   --daemon.dns-search value         docker daemon dns search domains [$PLUGIN_CUSTOM_DNS_SEARCH]
   --daemon.insecure                 docker daemon allows insecure registries [$PLUGIN_INSECURE]
   --daemon.ipv6                     docker daemon IPv6 networking [$PLUGIN_IPV6]
   --daemon.experimental             docker daemon Experimental mode [$PLUGIN_EXPERIMENTAL]
   --daemon.debug                    docker daemon executes in debug mode [$PLUGIN_DEBUG, $DOCKER_LAUNCH_DEBUG]
   --daemon.off                      don't start the docker daemon [$PLUGIN_DAEMON_OFF]
   --dockerfile value                build dockerfile (default: "Dockerfile") [$PLUGIN_DOCKERFILE]
   --context value                   build context (default: ".") [$PLUGIN_CONTEXT]
   --tags value                      build tags (default: "latest") [$PLUGIN_TAG, $PLUGIN_TAGS] [.tags]
   --tags.auto                       default build tags [$PLUGIN_DEFAULT_TAGS, $PLUGIN_AUTO_TAG]
   --tags.suffix value               default build tags with suffix [$PLUGIN_DEFAULT_SUFFIX, $PLUGIN_AUTO_TAG_SUFFIX]
   --args value                      build args [$PLUGIN_BUILD_ARGS]
   --args-from-env value             build args [$PLUGIN_BUILD_ARGS_FROM_ENV]
   --args-new value                  build args new [$PLUGIN_BUILD_ARGS_NEW]
   --plugin-multiple-build-agrs      plugin multiple build agrs [$PLUGIN_MULTIPLE_BUILD_ARGS]
   --quiet                           quiet docker build [$PLUGIN_QUIET]
   --target value                    build target [$PLUGIN_TARGET]
   --cache-from value                images to consider as cache sources [$PLUGIN_CACHE_FROM]
   --squash                          squash the layers at build time [$PLUGIN_SQUASH]
   --pull-image                      force pull base image at build time [$PLUGIN_PULL_IMAGE]
   --compress                        compress the build context using gzip [$PLUGIN_COMPRESS]
   --repo value                      docker repository [$PLUGIN_REPO]
   --custom-labels value             additional k=v labels [$PLUGIN_CUSTOM_LABELS]
   --label-schema value              label-schema labels [$PLUGIN_LABEL_SCHEMA]
   --auto-label                      auto-label true|false [$PLUGIN_AUTO_LABEL]
   --link value                      link https://example.com/org/repo-name [$PLUGIN_REPO_LINK, $DRONE_REPO_LINK]
   --bake.file value                 Build definition file [$PLUGIN_BAKE_FILE]
   --bake.target value               A target in a Bake file represents a build invocation [$PLUGIN_BAKE_TARGET]
   --bake.provenance value           Shorthand for "--set=*.attest=type=provenance" [$PLUGIN_BAKE_PROVENANCE]
   --bake.sbom value                 Shorthand for "--set=*.attest=type=sbom" [$PLUGIN_BAKE_SBOM]
   --bake.set value                  Override target value (e.g., "targetpattern.key=value") [$PLUGIN_BAKE_SET]
   --bake.envfile value              will 'source ${bake.envfile}' [$PLUGIN_BAKE_ENVFILE]
   --bake.variable value             load env [$PLUGIN_BAKE_VARIABLE]
   --bake.tags-variable-name value   Tags variable name generated after using tags or tags.auto. Default "TAGS" (default: "TAGS") [$PLUGIN_BAKE_TAGS_NAME]
   --docker.registry value           docker registry (default: "https://index.docker.io/v1/") [$PLUGIN_REGISTRY, $DOCKER_REGISTRY]
   --docker.username value           docker username [$PLUGIN_USERNAME, $DOCKER_USERNAME]
   --docker.password value           docker password [$PLUGIN_PASSWORD, $DOCKER_PASSWORD]
   --docker.baseimageusername value  Docker username for base image registry [$PLUGIN_DOCKER_USERNAME, $PLUGIN_BASE_IMAGE_USERNAME, $DOCKER_BASE_IMAGE_USERNAME]
   --docker.baseimagepassword value  Docker password for base image registry [$PLUGIN_DOCKER_PASSWORD, $PLUGIN_BASE_IMAGE_PASSWORD, $DOCKER_BASE_IMAGE_PASSWORD]
   --docker.baseimageregistry value  Docker registry for base image registry [$PLUGIN_DOCKER_REGISTRY, $PLUGIN_BASE_IMAGE_REGISTRY, $DOCKER_BASE_IMAGE_REGISTRY]
   --docker.email value              docker email [$PLUGIN_EMAIL, $DOCKER_EMAIL]
   --docker.config value             docker json dockerconfig content [$PLUGIN_CONFIG, $DOCKER_PLUGIN_CONFIG]
   --docker.purge                    docker should cleanup images [$PLUGIN_PURGE]
   --repo.branch value               repository default branch [$DRONE_REPO_BRANCH]
   --no-cache                        do not use cached intermediate containers [$PLUGIN_NO_CACHE]
   --add-host value                  additional host:IP mapping [$PLUGIN_ADD_HOST]
   --secret value                    secret key value pair eg id=MYSECRET [$PLUGIN_SECRET]
   --secrets-from-env value          secret key value pair eg secret_name=secret [$PLUGIN_SECRETS_FROM_ENV]
   --secrets-from-file value         secret key value pairs eg secret_name=/path/to/secret [$PLUGIN_SECRETS_FROM_FILE]
   --drone-card-path value           card path location to write to [$DRONE_CARD_PATH]
   --platform value                  platform value to pass to docker [$PLUGIN_PLATFORM]
   --ssh-agent-key value             ssh agent key to use [$PLUGIN_SSH_AGENT_KEY]
   --artifact-file value             Artifact file location that will be generated by the plugin. This file will include information of docker images that are uploaded by the plugin. [$PLUGIN_ARTIFACT_FILE]
   --registry-type value             registry type [$PLUGIN_REGISTRY_TYPE]
   --access-token value              access token [$ACCESS_TOKEN]
   --help, -h                        show help
   --version, -v                     print the version
```

### Using Docker buildkit Secrets

```yaml
kind: pipeline
name: default

steps:
- name: build dummy docker file and publish
  image: kit101z/dockerbuildkit
  pull: never
  settings:
    repo: kit101z/test
    tags: latest
    secret: id=mysecret,src=secret-file
    username:
      from_secret: docker_username
    password:
      from_secret: docker_password
```

Using a dockerfile that references the secret-file 

```bash
# syntax=docker/dockerfile:1.2

FROM alpine

# shows secret from default secret location:
RUN --mount=type=secret,id=mysecret cat /run/secrets/mysecret
```

and a secret file called secret-file

```
COOL BANANAS
```


### Running from the CLI

```console
docker run --rm \
  -e PLUGIN_TAG=latest \
  -e PLUGIN_REPO=octocat/hello-world \
  -e DRONE_COMMIT_SHA=d8dbe4d94f15fe89232e0402c6e8a0ddf21af3ab \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  --privileged \
  kit101z/dockerbuildkit --dry-run

# multi arch by bake file
docker run --rm \
  -e PLUGIN_TAG=latest \
  -e PLUGIN_REPO=octocat/hello-world \
  -e DRONE_COMMIT_SHA=d8dbe4d94f15fe89232e0402c6e8a0ddf21af3ab \
  -e PLUGIN_BAKE_FILE=docker-bake.hcl \
  -e PLUGIN_PLATFORM=linux/amd64,linux/arm64 \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  --privileged \
  kit101z/dockerbuildkit --dry-run
```


## Developer Notes

- When updating the base image, you will need to update for each architecture and OS.
- Arm32 base images are no longer being updated.

## Release procedure

Run the changelog generator.

```BASH
GITHUB_TOKEN=<secret token> scripts/changelog.sh
```

You can generate a token by logging into your GitHub account and going to Settings -> Personal access tokens.

Next we tag the PR's with the fixes or enhancements labels. If the PR does not fufil the requirements, do not add a label.

Run the changelog generator again with the future version according to semver.

```BASH
GITHUB_TOKEN=<secret token> scripts/changelog.sh --future-release v1.0.0
```

Create your pull request for the release. Get it merged then tag the release.

