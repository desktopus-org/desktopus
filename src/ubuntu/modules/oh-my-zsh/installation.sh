#!/bin/bash
set -eu -o pipefail
apt-get install -y zsh
sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
usermod --shell /bin/zsh desktopus