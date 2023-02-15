FROM golang:1.20.1

WORKDIR /go/ci

# ci:slim

COPY images/slim.sh images/slim.sh
RUN bash images/slim.sh

# custom tools
COPY . /go/ci
RUN go install ./...
