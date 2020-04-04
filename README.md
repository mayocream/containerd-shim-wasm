# containerd-shim-wasm

containerd-shim-wasm is a fork of [containerd-wasm](https://github.com/dmcgowan/containerd-wasm)
that implements the [containerd shim API
v2](https://github.com/containerd/containerd/tree/master/runtime/v2) for running WebAssembly
modules. The current implementation uses the [wasmer](https://github.com/wasmerio/wasmer) runtime.

> Warning: this project is a proof of concept and not suitable for production

## Quickstart

We can demonstrate the capabilities of the shim by running the [hello-wasm](hello-wasm) module on a
local [kind](https://github.com/kubernetes-sigs/kind) cluster. To do this we must first install
Docker, kind and kubectl.

### Deploy

Once the prerequisites are installed we can deploy our cluster using `make deploy`. This will do the
following:

- Build `containerd-shim-wasm-v1` and download the `wasmer` runtime
- Create a local kind cluster and bind mount the binaries
- Deploy the `hello-wasm` WebAssembly module to the cluster

### Logs

```
$ kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
wasm-5959bbb595-4vdw5   1/1     Running   0          57s
$ kubectl logs wasm-5959bbb595-4vdw5 -f
OS name: wasi
Hardware identifier: wasm32

Arguments:
argv[0]: /run/containerd/io.containerd.runtime.v2.task/k8s.io/2f5ca10471accc20520d38050283963077df4ff380f8b634659b85e31d8fa35b/rootfs/hello-wasm

Environment:
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=wasm-5959bbb595-4vdw5
KUBERNETES_SERVICE_PORT_HTTPS=443
KUBERNETES_PORT=tcp://10.96.0.1:443
KUBERNETES_PORT_443_TCP=tcp://10.96.0.1:443
KUBERNETES_PORT_443_TCP_PROTO=tcp
KUBERNETES_PORT_443_TCP_PORT=443
KUBERNETES_PORT_443_TCP_ADDR=10.96.0.1
KUBERNETES_SERVICE_HOST=10.96.0.1
KUBERNETES_SERVICE_PORT=443

Waiting 10 seconds (1)...
Waiting 10 seconds (2)...
Waiting 10 seconds (3)...
...
```

### Cleanup

```sh
make delete
```

## Alternatives

One difficulty with this shim implementation is that the shim API assumes a container runtime (as
that's what it was designed for), but this doens't align as well with running WebAssmebly modules
(for example currently you can't exec into a WebAssmebly module as you would a container). The
[Krustlet](https://github.com/deislabs/krustlet) project implements a Kubelet replacement that
treats wasi/wasm modules as first class citizens.
