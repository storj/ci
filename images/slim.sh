set -xeuo pipefail

# Disable safe.directory https://git-scm.com/docs/git-config#Documentation/git-config.txt-safedirectory,
# because we are running in containerized environment anyways.
git config --global --add safe.directory '*'

# Tooling

apt-get update && apt-get install -y ca-certificates curl gnupg
mkdir -p /etc/apt/keyrings
curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg

NODE_MAJOR=24
echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_MAJOR.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list

apt-get update && DEBIAN_FRONTEND="noninteractive" apt-get install -y brotli unzip libuv1-dev libjson-c-dev nettle-dev nodejs git-restore-mtime

npm install -g pnpm@latest-10

curl -sfL https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip -o /tmp/protoc.zip && unzip /tmp/protoc.zip -d "$HOME"/protoc

# Linters

# Shellcheck for linting shell scripts
apt-get -y install shellcheck

apt-get install -yq clang-format
