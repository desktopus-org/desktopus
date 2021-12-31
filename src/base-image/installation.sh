#!/bin/bash
set -eu -o pipefail

echo "Installing necessary software..."
apt-get update

apt-get install -y \
	software-properties-common \
 	ca-certificates \
	openssh-client \
	vim \
	wget \
	curl \
	iptables \
	net-tools \
	locales \
	bzip2 \
	zip \
	git \
	supervisor \
    nano \
	geany \
	terminator \
	sudo \
	python3 \
	python3-setuptools \
	xdg-utils

# For audio
apt-get install -y \
	alsa-utils libpulse0 \
	--no-install-recommends

# For audio and gpu acceleration
#apt-get install -y \
#	mesa-utils \
#	mesa-utils-extra \
#	libgl1-mesa-dri \
#	libgl1-mesa-glx \
#	--no-install-recommends

# Docker installation
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

if ! wget -O docker.tgz "https://download.docker.com/linux/static/${DOCKER_CHANNEL}/${dockerArch}/docker-${DOCKER_VERSION}.tgz"; then
	echo >&2 "error: failed to download 'docker-${DOCKER_VERSION}' from '${DOCKER_CHANNEL}' for '${dockerArch}'"
	exit 1
fi

tar --extract \
	--file docker.tgz \
	--strip-components 1 \
	--directory /usr/local/bin/

rm docker.tgz

dockerd --version
docker --version

# Docker compose installation
curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
docker-compose version

echo "Installing XFCE4..."
echo "Install Xfce4 UI components..."
apt-get update && apt-get install -y xfce4 xfce4-goodies

echo "Installing themes..."
apt-get update && apt-get install -y arc-theme numix-icon-theme numix-gtk-theme

echo "Installing Tiger VNC..."
apt-get update && apt install -y tigervnc-standalone-server

echo "Installing noVNC version v${NOVNC_VERSION}..."
echo "noVNC step 1: Install web contents"
# Install noVNC
mkdir -p /opt/noVNC
curl -L "https://github.com/novnc/noVNC/zipball/v${NOVNC_VERSION}" -o /opt/noVNC/noVNC.zip
pushd /opt/noVNC
unzip noVNC.zip
rm noVNC.zip
mv novnc-noVNC-* serve
popd
chown -R userdocker:userdocker /opt/noVNC
# Install websockify
echo "noVNC step 2: Install websockify"
pushd /opt
git clone https://github.com/novnc/websockify
pushd websockify
python3 setup.py install
popd
popd

echo "Installing google chrome..."
wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
apt install -y ./google-chrome-stable_current_amd64.deb
rm google-chrome-stable_current_amd64.deb
### fix to start chromium in a Docker container, see https://github.com/ConSol/docker-headless-vnc-container/issues/2
echo "CHROMIUM_FLAGS='--no-sandbox --start-maximized --user-data-dir'" > "$HOME"/.chromium-browser.init

echo "Insalling firefox..."
apt-get update && apt-get install -y firefox

