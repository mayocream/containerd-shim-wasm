FROM debian:10.2

RUN apt-get update && \
    apt-get install -y \
        curl \
        make \
        xz-utils \
        git \
        gcc

# Install golang
RUN cd /tmp && \
    curl -LO  https://dl.google.com/go/go1.14.linux-amd64.tar.gz && \
    tar -xvf go1.14.linux-amd64.tar.gz && \
    mv go /usr/local && \
    rm go1.14.linux-amd64.tar.gz
ENV GOROOT="/usr/local/go"
ENV GOPATH="/root/go"
ENV PATH="${GOPATH}/bin:${GOROOT}/bin:${PATH}"

# Install wasmer
ENV WASMER_VERSION 0.16.2
RUN cd /tmp && \
        curl -LO https://github.com/wasmerio/wasmer/releases/download/${WASMER_VERSION}/wasmer-linux-amd64.tar.gz && \
        tar -xzf wasmer-linux-amd64.tar.gz && \
        mv bin/wasmer /usr/local/bin/wasmer && \
        rm -r wasmer-linux-amd64.tar.gz bin

ENV HOME /root

WORKDIR /workspace
