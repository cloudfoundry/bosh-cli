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

# bosh-init dependencies
RUN apt-get install -y mercurial; apt-get clean
ADD install-go.sh /tmp/install-go.sh
RUN chmod a+x /tmp/install-go.sh
RUN cd tmp; ./install-go.sh; rm install-go.sh

# lifecycle ssh test
RUN apt-get install -y sshpass; apt-get clean

# package manager provides 1.4.3, which is too old for vagrant-aws
RUN cd /tmp && wget -q https://dl.bintray.com/mitchellh/vagrant/vagrant_1.6.5_x86_64.deb && dpkg -i vagrant_1.6.5_x86_64.deb

USER ubuntu
ENV HOME /home/ubuntu
ENV PATH /usr/local/ruby/bin:/usr/local/go/bin:$PATH
RUN mkdir -p /home/ubuntu/go/src/github.com/cloudfoundry
RUN vagrant plugin install vagrant-aws
