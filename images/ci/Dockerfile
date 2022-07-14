FROM golang:1.17.12

# ci:slim

COPY . /go/ci
WORKDIR /go/ci

RUN cd /go/ci && go install ./...
RUN bash images/slim.sh

# ci:deps

RUN bash images/deps.sh

# ci:full

ENV PATH="$PATH:/root/bin"
RUN bash images/ci.sh

# Set our entrypoint to close after 28 minutes, and forcefully close at 30 minutes.
# This is to prevent Jenkins collecting cats.
ENTRYPOINT ["timeout", "-k30m", "28m"]
