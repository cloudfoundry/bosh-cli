FROM ubuntu:14.04

RUN locale-gen en_US.UTF-8
RUN dpkg-reconfigure locales
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8

RUN apt-get update
RUN apt-get -y upgrade && apt-get clean

RUN apt-get install -y build-essential zlib1g-dev libssl-dev libxml2-dev libsqlite3-dev libxslt1-dev libpq-dev libmysqlclient-dev
RUN apt-get install -y git curl wget tar

RUN apt-get clean

# ...ruby
ADD install-ruby.sh /tmp/install-ruby.sh
RUN chmod a+x /tmp/install-ruby.sh
RUN cd tmp && ./install-ruby.sh && rm install-ruby.sh

# package manager provides 1.4.3, which is too old for vagrant-aws
RUN cd /tmp && wget -q https://releases.hashicorp.com/vagrant/1.8.6/vagrant_1.8.6_x86_64.deb \
 && echo "e6d83b6b43ad16475cb5cfcabe7dc798002147c1d048a7b6178032084c7070da vagrant_1.8.6_x86_64.deb" | sha256sum -c - \
 && dpkg -i vagrant_1.8.6_x86_64.deb
RUN vagrant plugin install vagrant-aws

# bosh-init dependencies
RUN apt-get install -y mercurial && apt-get clean

# golang {
ARG GO_VERSION=1.9.2
ARG GO_SHA=de874549d9a8d8d8062be05808509c09a88a248e77ec14eb77453530829ac02b

ENV GO_TAR=go$GO_VERSION.linux-amd64.tar.gz

WORKDIR /usr/local

ENV \
  PATH=/usr/local/go/bin:$PATH \
  GOPATH=/root/go \
  GOROOT=/usr/local/go

RUN curl -fSL https://storage.googleapis.com/golang/$GO_TAR -o $GO_TAR \
  && echo $GO_SHA $GO_TAR | sha256sum -c - \
  && tar -xzf $GO_TAR
# }

# lifecycle ssh test
RUN apt-get install -y sshpass && apt-get clean

# integration registry tests
RUN apt-get install -y openssh-server

# for hijack debugging
RUN apt-get install -y lsof psmisc strace
