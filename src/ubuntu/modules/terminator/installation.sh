#!/bin/bash
set -eu -o pipefail
echo "Insalling terminator..."
apt-get update && apt-get install -y terminator