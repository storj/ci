FROM golang:1.15.2

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

# CockroachDB

RUN wget -qO- https://binaries.cockroachdb.com/cockroach-v20.1.1.linux-amd64.tgz | tar  xvz
RUN cp -i cockroach-v20.1.1.linux-amd64/cockroach /usr/local/bin/

# Postgres

RUN curl -sf https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ buster-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list
RUN curl -sfL https://deb.nodesource.com/setup_14.x  | bash -

RUN apt-get update && apt-get install -y -qq postgresql-12 redis-server unzip libuv1-dev libjson-c-dev nettle-dev nodejs

RUN rm /etc/postgresql/12/main/pg_hba.conf; \
	echo 'local   all             all                                     trust' >> /etc/postgresql/12/main/pg_hba.conf; \
	echo 'host    all             all             127.0.0.1/8             trust' >> /etc/postgresql/12/main/pg_hba.conf; \
	echo 'host    all             all             ::1/128                 trust' >> /etc/postgresql/12/main/pg_hba.conf; \
	echo 'host    all             all             ::0/0                   trust' >> /etc/postgresql/12/main/pg_hba.conf;

RUN echo 'max_connections = 1000' >> /etc/postgresql/12/main/conf.d/connectionlimits.conf

# Tooling

COPY ./scripts/install-awscli.sh /tmp/install-awscli.sh
RUN bash /tmp/install-awscli.sh
ENV PATH "$PATH:/root/bin"

RUN curl -sfL https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip -o /tmp/protoc.zip && unzip /tmp/protoc.zip -d "$HOME"/protoc

# Android/Java binding tests
RUN apt-get install -y default-jre

# Duplicity backup tool for S3 gateway test scenarios
RUN apt-get install -y duplicity python-pip && pip install boto

# Duplicati backup tool for S3 gateway test scenarios
RUN apt-get -y install mono-devel
RUN curl -sfL https://updates.duplicati.com/beta/duplicati_2.0.5.1-1_all.deb -o /tmp/duplicati.deb
RUN apt -y install /tmp/duplicati.deb

# Linters

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ${GOPATH}/bin v1.30.0

RUN GO111MODULE=on go get \
    # Linters formatters \
    github.com/ckaznocha/protoc-gen-lint@v0.2.1 \
    github.com/nilslice/protolock/cmd/protolock@v0.15.0 \
    github.com/josephspurrier/goversioninfo@63e6d1acd3dd857ec6b8c54fbf52e10ce24a8786 \
    github.com/loov/leakcheck@83e415ebc9b993a8a0443bb788b0f737a50c4b62 \
    honnef.co/go/tools/cmd/staticcheck@2020.1.5 \
    # Output formatters \
    github.com/mfridman/tparse@36f80740879e24ba6695649290a240c5908ffcbb \
    github.com/axw/gocov/gocov@v1.0.0 \
    github.com/AlekSi/gocov-xml@3a14fb1c4737b3995174c5f4d6d08a348b9b4180

RUN apt-get install -yq clang-format

# Install go-licenses
#
# NOTE: It requires its own go path because it uses db files from the licenses
# go module.
RUN mkdir -p /ci/go-licenses && \
    GO111MODULE=on GOPATH=/ci/go-licenses go get \
    github.com/google/go-licenses@2ee7a02f6ae4f78b6b2d6ef421cedadbeabe2a89
ENV PATH "$PATH:/ci/go-licenses/bin"

# Tools in this repository
COPY . /go/ci
WORKDIR /go/ci
RUN go install ...

# Reset to starting directory
WORKDIR /go

# Set our entrypoint to close after 28 minutes, and forcefully close at 30 minutes.
# This is to prevent Jenkins collecting cats.
ENTRYPOINT ["timeout", "-k30m", "28m"]
