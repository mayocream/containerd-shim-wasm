# containerd-shim-wasm

containerd-shim-wasm is a fork of [containerd-wasm](https://github.com/dmcgowan/containerd-wasm)
that implements the [containerd shim API
v2](https://github.com/containerd/containerd/tree/master/runtime/v2) for running WebAssembly
modules using the [wasmtime](https://github.com/bytecodealliance/wasmtime) runtime.

> Warning: this project is a proof of concept and not suitable for production

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
