FROM golang:1

RUN \
  apt-get update \
  && apt-get -y install \
    default-libmysqlclient-dev \
    libpq-dev \
    libsqlite3-dev \
    lsof \
    ruby \
    openssh-server \
    psmisc \
    sshpass \
    strace \
    zlib1g-dev \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*
