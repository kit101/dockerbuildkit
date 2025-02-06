#!/bin/sh

docker run --rm -w /src -v ./:/src githubchangeloggenerator/github-changelog-generator \
  -u kit101 -p dockerbuildkit -t $GITHUB_TOKEN $@
