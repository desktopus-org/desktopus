#!/bin/bash
set -eu -o pipefail
echo "Insalling firefox..."
apt-get update && apt-get install -y firefox