#!/usr/bin/env bash
set -euo pipefail
set -x

mkdir -p "$HOME/awscli"
pushd "$HOME/awscli"

curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install -b ~/bin

popd
rm -r "$HOME/awscli"

set +x