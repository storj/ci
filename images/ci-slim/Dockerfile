# tools image that we can cache
FROM golang:1.21.0 AS tools

WORKDIR /go/ci

COPY . /go/ci
RUN go install ./...

FROM golang:1.21.0

WORKDIR /go/ci

# ci:slim

COPY images/slim.sh images/slim.sh
RUN bash images/slim.sh

# tools

COPY --from=tools /go/ci    /go/ci
COPY --from=tools /go/bin/* /go/bin/