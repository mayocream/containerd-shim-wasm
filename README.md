# containerd-shim-wasm

containerd-shim-wasm is a fork of [containerd-wasm](https://github.com/dmcgowan/containerd-wasm)
that implements the [containerd shim API
v2](https://github.com/containerd/containerd/tree/master/runtime/v2) for running WebAssembly
modules using the [wasmtime](https://github.com/bytecodealliance/wasmtime) runtime.

> Warning: this project is a proof of concept and not suitable for production

## Alternatives

One difficulty with this shim implementation is that the shim API assumes a container runtime (as
that's what it was designed for), but this doens't align as well with running WebAssmebly modules
(for example currently you can't exec into a WebAssmebly module as you would a container). The
[Krustlet](https://github.com/deislabs/krustlet) project implements a Kubelet replacement that
treats wasi/wasm modules as first class citizens.

## Quickstart

`make all` will do the following:

- Build or download the following binaries
  - [buildx](https://github.com/docker/buildx) Docker plugin
  - [docker-wasm](https://github.com/tonistiigi/wasm-cli-plugin) Docker plugin
  - hello-wasm WebAssembly module
  - containerd-shim-wasm-v1
  - wasmtime
- Create a local [kind](https://github.com/kubernetes-sigs/kind) cluster
- Deploy the WebAssembly module to the cluster
