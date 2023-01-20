FROM golang:1.19.5

WORKDIR /go/ci

# ci:slim

COPY images/slim.sh images/slim.sh
RUN bash images/slim.sh

# custom tools
COPY . /go/ci
RUN go install ./...
