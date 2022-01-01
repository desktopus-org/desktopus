#!/bin/bash
docker build --pull --no-cache --rm=true . -t __workspace_name__:latest
docker tag __workspace_name__:latest __workspace_name__:__distro_and_version__