# tools image that we can cache
FROM golang:1.21.6 AS tools

WORKDIR /go/ci

COPY . /go/ci
RUN go install ./...

FROM golang:1.21.6

WORKDIR /go/ci

# ci:slim

COPY images/slim.sh images/slim.sh
RUN bash images/slim.sh

# ci:deps

COPY images/deps.sh images/deps.sh
RUN bash images/deps.sh

# ci:full

COPY scripts scripts
COPY images/ci.sh images/ci.sh
ENV PATH="$PATH:/root/bin"
RUN bash images/ci.sh

# tools

COPY --from=tools /go/ci    /go/ci
COPY --from=tools /go/bin/* /go/bin/