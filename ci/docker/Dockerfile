FROM bosh/main-ruby-go

RUN \
  apt-get update \
  && apt-get -y install \
    libmysqlclient-dev \
    libpq-dev \
    libsqlite3-dev \
    lsof \
    openssh-server \
    psmisc \
    sshpass \
    strace \
    zlib1g-dev \
  && apt-get clean
