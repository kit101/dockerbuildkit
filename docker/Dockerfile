FROM docker:27.4.1-dind

ARG TARGETOS
ARG TARGETARCH

ENV DOCKER_HOST=unix:///var/run/docker.sock
ENV BUILDKITD_HOME=/etc/buildkit

RUN mkdir -p /etc/buildkit

COPY release/$TARGETOS/$TARGETARCH/dockerbuildkit /bin/

ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint.sh", "/bin/dockerbuildkit"]