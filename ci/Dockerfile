FROM ubuntu:14.04

RUN locale-gen en_US.UTF-8
RUN dpkg-reconfigure locales
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8

# create ubuntu user with no password
RUN useradd -mU ubuntu -s /bin/bash
RUN echo "ubuntu ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

RUN apt-get update
RUN apt-get -y upgrade; apt-get clean

RUN apt-get install -y build-essential; apt-get clean
RUN apt-get install -y git; apt-get clean
RUN apt-get install -y curl; apt-get clean
RUN apt-get install -y wget; apt-get clean
RUN apt-get install -y tar; apt-get clean

# bosh_cli dependencies
RUN apt-get install -y libssl-dev; apt-get clean
RUN apt-get install -y libreadline-dev; apt-get clean
RUN apt-get install -y libxml2-dev; apt-get clean
RUN apt-get install -y libxslt-dev; apt-get clean

# for go dependencies
RUN apt-get install -y mercurial; apt-get clean

# bosh-micro dependencies
ADD install-go.sh /tmp/install-go.sh
RUN chmod a+x /tmp/install-go.sh
RUN cd tmp; ./install-go.sh; rm install-go.sh

ADD install-ruby.sh /tmp/install-ruby.sh
RUN chmod a+x /tmp/install-ruby.sh
RUN cd /tmp; ./install-ruby.sh; rm install-ruby.sh

RUN gem install bosh_cli --no-rdoc --no-ri

USER ubuntu
ENV HOME /home/ubuntu
ENV PATH /usr/local/ruby/bin:/usr/local/go/bin:$PATH
RUN mkdir -p /home/ubuntu/go/src/github.com/cloudfoundry
