#!/bin/bash
set -e

echo "Installing necessary software..."
apt-get update

apt-get install -y vim wget net-tools locales bzip2 \
    python-numpy nano geany terminator sudo

# For audio and gpu acceleration
apt-get install -y alsa-utils \
	libgl1-mesa-dri \
	libgl1-mesa-glx \
	libpulse0 \
	xdg-utils \
	--no-install-recommends

echo "Installing XFCE4..."
echo "Install Xfce4 UI components..."
apt-get install -y supervisor xfce4 xfce4-terminal xfce4-whiskermenu-plugin xterm

echo "Installing themes..."
apt-get install -y arc-theme numix-icon-theme numix-gtk-theme

echo "Installing Tiger VNC..."
apt install -y tigervnc-standalone-server tigervnc-xorg-extension tigervnc-viewer

echo "Installing noVNC..."
apt -y install novnc websockify python-numpy

echo "Installing chromium..."
apt-get install -y chromium-browser chromium-browser-l10n chromium-codecs-ffmpeg
ln -s /usr/bin/chromium-browser /usr/bin/google-chrome
### fix to start chromium in a Docker container, see https://github.com/ConSol/docker-headless-vnc-container/issues/2
echo "CHROMIUM_FLAGS='--no-sandbox --start-maximized --user-data-dir'" > $HOME/.chromium-browser.init

echo "Insalling firefox..."
apt-get install -y firefox

