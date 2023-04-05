FROM golang:1.20.3

WORKDIR /go/ci

# ci:slim

COPY images/slim.sh images/slim.sh
RUN bash images/slim.sh

# custom tools
COPY . /go/ci
RUN go install ./...
