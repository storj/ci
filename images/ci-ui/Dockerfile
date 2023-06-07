FROM golang:1.20.1

WORKDIR /go/ci

# ci:slim

COPY images/slim.sh images/slim.sh
RUN bash images/slim.sh

# ci:deps

COPY images/deps.sh images/deps.sh
RUN bash images/deps.sh

# ci:

# ci:ui

COPY images/ui.sh images/ui.sh
RUN bash images/ui.sh

# custom tools
COPY . /go/ci
RUN go install ./...
