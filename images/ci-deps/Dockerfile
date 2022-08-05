FROM golang:1.17.12

# ci:slim

COPY . /go/ci
WORKDIR /go/ci

RUN cd /go/ci && go install ./...
RUN bash images/slim.sh

# ci:deps

RUN bash images/deps.sh
