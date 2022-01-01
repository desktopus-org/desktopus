#!/bin/bash
arch="$(uname --m)"
case "$arch" in
	# amd64
		x86_64) dockerArch='x86_64' 
	;;
	# arm32v6
		armhf) dockerArch='armel'
	;;
	# arm32v7
		armv7) dockerArch='armhf'
	;;
	# arm64v8
		aarch64) dockerArch='aarch64'
	;;
	*) 
		echo >&2 "error: unsupported architecture ($arch)"
		exit 1 
	;;
esac

if ! wget -O docker.tgz "https://download.docker.com/linux/static/stable/${dockerArch}/docker-__version__.tgz"; then
	echo >&2 "error: failed to download 'docker-__version__' from 'stable' for '${dockerArch}'"
	exit 1
fi

tar --extract \
	--file docker.tgz \
	--strip-components 1 \
	--directory /usr/local/bin/

rm docker.tgz

dockerd --version
docker --version

echo "Creating docker group and adding desktopus to the group"

groupadd docker
usermod -aG docker desktopus
