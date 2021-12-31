#!/bin/bash
# Build image
pushd src/base-image || exit 1
docker build --pull --no-cache --rm=true . -t desktopus/ubuntu-base-xfce:latest
docker tag desktopus/ubuntu-base-xfce:latest desktopus/ubuntu-base-xfce:20.04
popd || exit 1