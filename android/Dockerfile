FROM golang:1.14.4

# CockroachDB

RUN wget -qO- https://binaries.cockroachdb.com/cockroach-v20.1.1.linux-amd64.tgz | tar  xvz
RUN cp -i cockroach-v20.1.1.linux-amd64/cockroach /usr/local/bin/

# Postgres

RUN curl https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ buster-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list
RUN curl -sL https://deb.nodesource.com/setup_13.x  | bash -

RUN apt-get update && apt-get install -y -qq postgresql-12 redis-server unzip libuv1-dev libjson-c-dev nettle-dev nodejs

RUN rm /etc/postgresql/12/main/pg_hba.conf; \
	echo 'local   all             all                                     trust' >> /etc/postgresql/12/main/pg_hba.conf; \
	echo 'host    all             all             127.0.0.1/8             trust' >> /etc/postgresql/12/main/pg_hba.conf; \
	echo 'host    all             all             ::1/128                 trust' >> /etc/postgresql/12/main/pg_hba.conf; \
	echo 'host    all             all             ::0/0                   trust' >> /etc/postgresql/12/main/pg_hba.conf;

RUN echo 'max_connections = 1000' >> /etc/postgresql/12/main/conf.d/connectionlimits.conf

# Tooling

# COPY ./scripts/install-awscli.sh /tmp/install-awscli.sh
# RUN bash /tmp/install-awscli.sh
ENV PATH "$PATH:/root/bin"

RUN curl -L https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip -o /tmp/protoc.zip && unzip /tmp/protoc.zip -d "$HOME"/protoc

# Android/Java binding tests
RUN apt-get install -y default-jre

# Duplicity backup tool for S3 gateway test scenarios
RUN apt-get install -y duplicity python-pip && pip install boto

# Duplicati backup tool for S3 gateway test scenarios
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 3FA7E0328081BFF6A14DA29AA6A19B38D3D831EF
RUN "deb http://download.mono-project.com/repo/debian stable-buster main" | tee /etc/apt/sources.list.d/mono-official.list
RUN apt-get update && apt-get -y install mono-devel
RUN curl -L https://updates.duplicati.com/beta/duplicati_2.0.5.1-1_all.deb -o /tmp/duplicati.deb
# installation from deb is failing but next step will fix missing deps
RUN dpkg -i /tmp/duplicati.deb; exit 0
RUN apt install -y -f

# Android SDK + NDK

ENV ANDROID_HOME /opt/android-sdk-linux
RUN apt-get update -qq

RUN dpkg --add-architecture i386
RUN apt-get update -qq
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y unzip rsync libc6:i386 libstdc++6:i386 libgcc1:i386 libncurses5:i386 libz1:i386

RUN cd /opt \
    && wget -q https://dl.google.com/android/repository/sdk-tools-linux-4333796.zip -O android-sdk-tools.zip \
    && unzip -q android-sdk-tools.zip -d ${ANDROID_HOME} \
    && rm android-sdk-tools.zip

# hack to make sdkmanager working with Java 11
RUN cd ${ANDROID_HOME}/tools/bin \
    && mkdir jaxb_lib \
    && wget https://repo1.maven.org/maven2/javax/activation/activation/1.1.1/activation-1.1.1.jar -O jaxb_lib/activation.jar \
    && wget https://repo1.maven.org/maven2/javax/xml/jaxb-impl/2.1/jaxb-impl-2.1.jar -O jaxb_lib/jaxb-impl.jar \
    && wget https://repo1.maven.org/maven2/org/glassfish/jaxb/jaxb-xjc/2.3.2/jaxb-xjc-2.3.2.jar -O jaxb_lib/jaxb-xjc.jar \
    && wget https://repo1.maven.org/maven2/org/glassfish/jaxb/jaxb-core/2.3.0.1/jaxb-core-2.3.0.1.jar -O jaxb_lib/jaxb-core.jar \
    && wget https://repo1.maven.org/maven2/org/glassfish/jaxb/jaxb-jxc/2.3.2/jaxb-jxc-2.3.2.jar -O jaxb_lib/jaxb-jxc.jar \
    && wget https://repo1.maven.org/maven2/javax/xml/bind/jaxb-api/2.3.1/jaxb-api-2.3.1.jar -O jaxb_lib/jaxb-api.jar
RUN export JAXB=${ANDROID_HOME}/tools/bin/jaxb_lib/activation.jar:${ANDROID_HOME}/tools/bin/jaxb_lib/jaxb-impl.jar:${ANDROID_HOME}/tools/bin/jaxb_lib/jaxb-xjc.jar:${ANDROID_HOME}/tools/bin/jaxb_lib/jaxb-core.jar:${ANDROID_HOME}/tools/bin/jaxb_lib/jaxb-jxc.jar:${ANDROID_HOME}/tools/bin/jaxb_lib/jaxb-api.jar \
    && sed -i '/^eval set -- $DEFAULT_JVM_OPTS.*/i CLASSPATH='$JAXB':$CLASSPATH' ${ANDROID_HOME}/tools/bin/sdkmanager \
	&& sed -i '/^eval set -- $DEFAULT_JVM_OPTS.*/i CLASSPATH='$JAXB':$CLASSPATH' ${ANDROID_HOME}/tools/bin/avdmanager

ENV PATH ${PATH}:${ANDROID_HOME}/tools:${ANDROID_HOME}/tools/bin:${ANDROID_HOME}/platform-tools

# accept all licenses
RUN yes | sdkmanager  --licenses
RUN touch /root/.android/repositories.cfg

# Platform tools
RUN yes | sdkmanager "platform-tools" "platforms;android-24" "tools" "emulator"

# The `yes` is for accepting all non-standard tool licenses.
RUN yes | sdkmanager --update --channel=3
RUN yes | sdkmanager \
    "ndk-bundle" \
    "system-images;android-24;default;x86_64"

# Linters

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ${GOPATH}/bin v1.28.3

RUN GO111MODULE=on go get \
    # Linters formatters \
    github.com/ckaznocha/protoc-gen-lint@v0.2.1 \
    github.com/nilslice/protolock/cmd/protolock@v0.15.0 \
    github.com/josephspurrier/goversioninfo@63e6d1acd3dd857ec6b8c54fbf52e10ce24a8786 \
    github.com/loov/leakcheck@83e415ebc9b993a8a0443bb788b0f737a50c4b62 \
    honnef.co/go/tools/cmd/staticcheck@v0.0.1-2020.1.4 \
    # Output formatters \
    github.com/mfridman/tparse@36f80740879e24ba6695649290a240c5908ffcbb \
    github.com/axw/gocov/gocov@v1.0.0 \
    github.com/AlekSi/gocov-xml@3a14fb1c4737b3995174c5f4d6d08a348b9b4180

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
