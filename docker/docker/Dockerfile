FROM docker:27.3.1-dind

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV DOCKER_HOST=unix:///var/run/docker.sock

ADD release/linux/$TARGETARCH/drone-docker /bin/
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/drone-docker"]
