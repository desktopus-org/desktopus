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
