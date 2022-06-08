set -euo pipefail

# Older versions of Go

go install golang.org/dl/go1.14@latest && go1.14 download

# Tooling

curl -sfL https://deb.nodesource.com/setup_16.x  | bash -
apt-get update && DEBIAN_FRONTEND="noninteractive" apt-get install -y brotli unzip libuv1-dev libjson-c-dev nettle-dev nodejs

curl -sfL https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip -o /tmp/protoc.zip && unzip /tmp/protoc.zip -d "$HOME"/protoc

# Linters

# Shellcheck for linting shell scripts
apt-get -y install shellcheck

# Linters, formatters, build tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2 && \
    go install github.com/ckaznocha/protoc-gen-lint@v0.2.4 && \
    go install github.com/nilslice/protolock/cmd/protolock@v0.15.2 && \
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@63e6d1acd3dd857ec6b8c54fbf52e10ce24a8786 && \
    go install honnef.co/go/tools/cmd/staticcheck@2022.1.2 && \
    # Output formatters \
    go install github.com/mfridman/tparse@36f80740879e24ba6695649290a240c5908ffcbb  && \
    go install github.com/axw/gocov/gocov@v1.0.0  && \
    go install github.com/AlekSi/gocov-xml@3a14fb1c4737b3995174c5f4d6d08a348b9b4180 && \
    go install github.com/google/go-licenses@ceb292363ec84358c9a276ef23aa0de893e59b84 && \
    go install github.com/magefile/mage@v1.11.0

apt-get install -yq clang-format
