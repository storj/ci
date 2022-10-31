FROM golang:1.18.7

# ci:slim

COPY . /go/ci
WORKDIR /go/ci

RUN cd /go/ci && go install ./...
RUN bash images/slim.sh

# ci:deps

RUN bash images/deps.sh
