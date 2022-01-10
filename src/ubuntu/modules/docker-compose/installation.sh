#!/bin/bash
set -eu -o pipefail
LATEST_DOCKER_COMPOSE_VERSION=$(curl -sSL https://api.github.com/repos/docker/compose/tags | grep name | head -n1 | cut -d'"' -f4)

OS="$(uname -s | awk '{print tolower($0)}')"
ARCH="$(uname -m)"

mkdir -p /usr/local/lib/docker/cli-plugins
rm -f /usr/local/lib/docker/cli-plugins/docker-compose
curl -SL "https://github.com/docker/compose/releases/download/${LATEST_DOCKER_COMPOSE_VERSION}/docker-compose-${OS}-${ARCH}" -o /usr/local/lib/docker/cli-plugins/docker-compose
chmod a+x /usr/local/lib/docker/cli-plugins/docker-compose
ln -s /usr/local/lib/docker/cli-plugins/docker-compose /usr/local/bin