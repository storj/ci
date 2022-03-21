FROM golang:1.17.5

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

# Older versions of Go

RUN go install golang.org/dl/go1.14@latest && go1.14 download

# CockroachDB

RUN wget -qO- https://binaries.cockroachdb.com/cockroach-v21.2.2.linux-amd64.tgz | tar  xvz
RUN cp -i cockroach-v21.2.2.linux-amd64/cockroach /usr/local/bin/

# Postgres

RUN curl -sf https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ bullseye-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list
RUN curl -sfL https://deb.nodesource.com/setup_16.x  | bash -

RUN apt-get update && apt-get install -y -qq postgresql-13 redis-server unzip libuv1-dev libjson-c-dev nettle-dev nodejs

RUN rm /etc/postgresql/13/main/pg_hba.conf; \
	echo 'local   all             all                                     trust' >> /etc/postgresql/13/main/pg_hba.conf; \
	echo 'host    all             all             127.0.0.1/8             trust' >> /etc/postgresql/13/main/pg_hba.conf; \
	echo 'host    all             all             ::1/128                 trust' >> /etc/postgresql/13/main/pg_hba.conf; \
	echo 'host    all             all            ::0/0                   trust' >> /etc/postgresql/13/main/pg_hba.conf;

RUN echo 'max_connections = 1000' >> /etc/postgresql/13/main/conf.d/connectionlimits.conf; \
    echo 'fsync = off' >> /etc/postgresql/13/main/conf.d/nosync.conf;

# Tooling

COPY ./scripts/install-awscli.sh /tmp/install-awscli.sh
RUN bash /tmp/install-awscli.sh
ENV PATH "$PATH:/root/bin"

RUN curl -sfL https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip -o /tmp/protoc.zip && unzip /tmp/protoc.zip -d "$HOME"/protoc

# Android/Java binding tests
RUN apt-get install -y default-jre

# Duplicity backup tool for S3 gateway test scenarios
RUN apt-get install -y duplicity python3-pip && pip install boto3

# Duplicati backup tool for S3 gateway test scenarios
RUN apt-get -y install mono-devel
RUN curl -sfL https://github.com/duplicati/duplicati/releases/download/v2.0.5.114-2.0.5.114_canary_2021-03-10/duplicati_2.0.5.114-1_all.deb -o /tmp/duplicati.deb
RUN apt -y install /tmp/duplicati.deb

# Requirements for UI tests
RUN DEBIAN_FRONTEND="noninteractive" apt install -y brotli chromium xorg xvfb gtk2-engines-pixbuf dbus-x11 xfonts-base xfonts-100dpi xfonts-75dpi xfonts-cyrillic xfonts-scalable imagemagick x11-apps

# Linters

# Shellcheck for linting shell scripts
RUN apt-get -y install shellcheck

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ${GOPATH}/bin v1.45.0

# Linters, formatters, build tools
RUN go install github.com/ckaznocha/protoc-gen-lint@v0.2.4 && \
    go install github.com/nilslice/protolock/cmd/protolock@v0.15.2 && \
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@63e6d1acd3dd857ec6b8c54fbf52e10ce24a8786 && \
    go install honnef.co/go/tools/cmd/staticcheck@2021.1.2 && \
    # Output formatters \
    go install github.com/mfridman/tparse@36f80740879e24ba6695649290a240c5908ffcbb  && \
    go install github.com/axw/gocov/gocov@v1.0.0  && \
    go install github.com/AlekSi/gocov-xml@3a14fb1c4737b3995174c5f4d6d08a348b9b4180 && \
    go install github.com/google/go-licenses@ceb292363ec84358c9a276ef23aa0de893e59b84 && \
    go install github.com/magefile/mage@v1.11.0

RUN apt-get install -yq clang-format

# Tools in this repository
COPY . /go/ci
WORKDIR /go/ci
RUN go install ./...

# Reset to starting directory
WORKDIR /go

# Set our entrypoint to close after 28 minutes, and forcefully close at 30 minutes.
# This is to prevent Jenkins collecting cats.
ENTRYPOINT ["timeout", "-k30m", "28m"]
