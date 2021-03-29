FROM golang:1.14.2-alpine as builder

ENV GO111MODULE=on

# hadolint ignore=DL3018
RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make gar

FROM ubuntu:20.04

# Build ARGS
ARG RUNNER_VERSION
ARG DOCKER_VERSION

ARG DEBIAN_FRONTEND=noninteractive

# Environment vars
ENV GITHUB_TOKEN=""
ENV REG_URL=""
ENV RUNNER_PATH="/runner"
ENV RUNNER_WORKDIR="/home/runner/work"
ENV RUNNER_NAME=""
ENV RUNNER_LABELS=""

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# hadolint ignore=DL3008
RUN apt-get update \
 && apt-get install -y --no-install-recommends software-properties-common  \
                                               curl \
                                               gpg-agent \
 && add-apt-repository -y ppa:git-core/ppa \
 && add-apt-repository ppa:rmescandon/yq \
 && curl -sL https://deb.nodesource.com/setup_14.x | bash - \
 && curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - \
 && sh -c "echo 'deb https://dl.yarnpkg.com/debian/ stable main' >> /etc/apt/sources.list" \
 && apt-get update \
 && apt-get install -y --no-install-recommends build-essential \
                                               ca-certificates \
                                               dnsutils \
                                               ftp \
                                               git \
                                               iproute2 \
                                               iputils-ping \
                                               jq \
                                               libunwind8 \
                                               locales \
                                               netcat \
                                               openssh-client \
                                               parallel \
                                               rsync \
                                               shellcheck \
                                               sudo \
                                               telnet \
                                               time \
                                               tzdata \
                                               unzip \
                                               upx \
                                               wget \
                                               zip \
                                               zstd \
                                               gnupg \
                                               gnupg-agent \
                                               python3-venv \
                                               python3-pip \
                                               python3-setuptools \
                                               python3-wheel \
                                               yq \
                                               nodejs \
                                               yarn \
                                               xvfb \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

RUN curl -L -o docker.tgz https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz \
 && tar zxvf docker.tgz \
 && install -o root -g root -m 755 docker/docker /usr/local/bin/docker \
 && rm -rf docker docker.tgz \
 && curl -L -o /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 \
 && chmod +x /usr/local/bin/dumb-init \
 && adduser --disabled-password --gecos "" --uid 1000 runner \
 && usermod -aG sudo runner \
 && echo "%sudo   ALL=(ALL:ALL) NOPASSWD:ALL" > /etc/sudoers

WORKDIR /runner

RUN curl -L -o runner.tar.gz https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz \
 && tar xzf ./runner.tar.gz \
 && rm runner.tar.gz \
 && ./bin/installdependencies.sh \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/build/gar /runner/gar

USER runner:runner
ENTRYPOINT [ "/runner/gar"]
