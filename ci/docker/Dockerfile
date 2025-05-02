FROM ghcr.io/cloudfoundry/golang-release

RUN \
  apt-get update \
  && apt-get -y install \
    default-libmysqlclient-dev \
    libpq-dev \
    libsqlite3-dev \
    lsof \
    ruby \
    ruby-bundler \
    openssh-server \
    psmisc \
    sshpass \
    strace \
    zlib1g-dev \
    libpcap-dev \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*
