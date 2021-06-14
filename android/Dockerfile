FROM golang:rc

# Postgres
RUN curl https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ buster-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list

RUN apt-get update && apt-get install -y -qq postgresql-12 redis-server unzip libuv1-dev libjson-c-dev nettle-dev

RUN rm /etc/postgresql/12/main/pg_hba.conf; \
	echo 'local   all             all                                     trust' >> /etc/postgresql/12/main/pg_hba.conf; \
	echo 'host    all             all             127.0.0.1/8             trust' >> /etc/postgresql/12/main/pg_hba.conf; \
	echo 'host    all             all             ::1/128                 trust' >> /etc/postgresql/12/main/pg_hba.conf; \
	echo 'host    all             all             ::0/0                   trust' >> /etc/postgresql/12/main/pg_hba.conf;

RUN echo 'max_connections = 1000' >> /etc/postgresql/12/main/conf.d/connectionlimits.conf

# Android/Java binding tests
RUN apt-get install -y default-jre

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
