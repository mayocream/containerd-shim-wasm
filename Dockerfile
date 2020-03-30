FROM debian:10.2

RUN apt-get update && \
    apt-get install -y \
        curl \
        make \
        xz-utils \
        git \
        gcc \
        golang

# Install golang
#RUN cd /tmp && \
#    curl -LO  https://dl.google.com/go/go1.14.linux-amd64.tar.gz && \
#    tar -xvf go1.14.linux-amd64.tar.gz && \
#    mv go /usr/local && \
#    rm go1.14.linux-amd64.tar.gz
#ENV GOROOT="/usr/local/go"
#ENV GOPATH="/root/go"
#ENV PATH="${GOPATH}/bin:${GOROOT}/bin:${PATH}"

# Install wasmtime
ENV WASMTIME_VERSION 0.60.0
RUN cd /tmp && \
        curl -LO https://github.com/bytecodealliance/wasmtime/releases/download/cranelift-v${WASMTIME_VERSION}/wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux.tar.xz && \
        tar -xf wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux.tar.xz && \
        mv wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux/wasmtime /usr/local/bin/wasmtime && \
        rm wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux.tar.xz && \
        rm -Rf wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux

ENV HOME /root

RUN apt-get install -y build-essential

WORKDIR /workspace
