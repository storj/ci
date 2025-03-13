# tools image that we can cache
FROM golang:1.24.1 AS tools

WORKDIR /go/ci

COPY . /go/ci
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go install ./...

FROM golang:1.24.1 AS linters
# Linters, formatters, build tools
ENV PATH=$PATH:/go/bin
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.7 && \
    go install github.com/ckaznocha/protoc-gen-lint@920e717f46849cad73a643be1c8f044cb4782101 && \
    go install github.com/nilslice/protolock/cmd/protolock@v0.17.0 && \
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@63e6d1acd3dd857ec6b8c54fbf52e10ce24a8786 && \
    go install honnef.co/go/tools/cmd/staticcheck@2025.1 && \
    go install github.com/mfridman/tparse@v0.15.0   && \
    go install github.com/axw/gocov/gocov@v1.1.0    && \
    go install github.com/AlekSi/gocov-xml@v1.1.0   && \
    go install github.com/google/go-licenses@v1.6.0 && \
    go install github.com/magefile/mage@v1.15.0

FROM golang:1.24.1 AS rclone
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    --mount=type=cache,target=/go/pkg/mod,id=gopkg \
    go install github.com/rclone/rclone@v1.65.0 && \
    go install github.com/rclone/rclone/fstest/test_all@v1.65.0

FROM golang:1.24.1 AS ci-slim

WORKDIR /go/ci

# ci:slim
COPY images/slim.sh images/slim.sh
RUN bash images/slim.sh

COPY --from=tools /go/ci    /go/ci
COPY --from=linters /go/bin/* /go/bin/
COPY --from=tools /go/bin/* /go/bin/

# do not remove go1.17.13 some uplink binary tests require an older Go version.
RUN go install golang.org/dl/go1.17.13@latest && \
    go1.17.13 download

# minimum version supported by our packages.
RUN go install golang.org/dl/go1.20.14@latest && \
    mv $(go env GOPATH)/bin/go1.20.14 $(go env GOPATH)/bin/go.min && \
    go.min download && \
    rm /root/sdk/go1.20.14/*.tar.gz

FROM ci-slim AS ci

# ci:deps

COPY images/deps.sh images/deps.sh
RUN bash images/deps.sh

# ci:full

COPY scripts scripts
COPY images/ci.sh images/ci.sh
# rclone and test tool for S3 gateway test scenarios

ENV PATH="$PATH:/root/bin:/usr/local/zig"
RUN bash images/ci.sh

# tools

COPY --from=tools /go/ci    /go/ci
COPY --from=tools /go/bin/* /go/bin/
COPY --from=rclone /go/bin/* /go/bin/
