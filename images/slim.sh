set -xeuo pipefail

# Disable safe.directory https://git-scm.com/docs/git-config#Documentation/git-config.txt-safedirectory,
# because we are running in containerized environment anyways.
git config --global --add safe.directory '*'

# Older versions of Go

# do not remove go1.17.13 some uplink binary tests require an older Go version.
go install golang.org/dl/go1.17.13@latest && go1.17.13 download
# minimum version supported by our packages.
go install golang.org/dl/go1.19.12@latest && \
    mv $(go env GOPATH)/bin/go1.19.12 $(go env GOPATH)/bin/go.min && \
    go.min download

# Tooling

curl -sfL https://deb.nodesource.com/setup_18.x  | bash -
apt-get update && DEBIAN_FRONTEND="noninteractive" apt-get install -y brotli unzip libuv1-dev libjson-c-dev nettle-dev nodejs
npm install -g npm@9.6.7
npm install -g pnpm@v8

curl -sfL https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip -o /tmp/protoc.zip && unzip /tmp/protoc.zip -d "$HOME"/protoc

# Linters

# Shellcheck for linting shell scripts
apt-get -y install shellcheck

# Linters, formatters, build tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2 && \
    go install github.com/ckaznocha/protoc-gen-lint@v0.3.0 && \
    go install github.com/nilslice/protolock/cmd/protolock@v0.16.0 && \
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@63e6d1acd3dd857ec6b8c54fbf52e10ce24a8786 && \
    go install honnef.co/go/tools/cmd/staticcheck@2023.1.5 && \
    # Output formatters \
    go install github.com/mfridman/tparse@v0.12.1   && \
    go install github.com/axw/gocov/gocov@v1.1.0    && \
    go install github.com/AlekSi/gocov-xml@v1.1.0   && \
    go install github.com/google/go-licenses@v1.6.0 && \
    go install github.com/magefile/mage@v1.11.0

apt-get install -yq clang-format
