FROM golang:1.20.6

WORKDIR /go/ci

# ci:slim

COPY images/slim.sh images/slim.sh
RUN bash images/slim.sh

# ci:deps

COPY images/deps.sh images/deps.sh
RUN bash images/deps.sh

# custom tools
COPY . /go/ci
RUN go install ./...
