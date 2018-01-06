FROM ubuntu:16.04

ARG go_version=1.8

# Install Go
RUN apt-get -q update -y
RUN apt-get -q install -y build-essential
RUN apt-get -q install -y curl
RUN apt-get -q install -y git
RUN apt-get -q install -y zip
RUN curl https://storage.googleapis.com/golang/go${go_version}.linux-amd64.tar.gz | tar -xz -C /usr/local

# Configure go environment
ENV GOROOT=/usr/local/go
ENV GOPATH=/opt/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin
RUN mkdir -p $GOPATH/bin $GOPATH/src $GOPATH/pkg
RUN go get -u github.com/kardianos/govendor

CMD /bin/bash