CLUSTER_NAME = test
BIN_DIR = bin
BUILD_IMAGE = wasm
APP_IMAGE = dippynark/hello-wasm:v0.1.0
WASMTIME_VERSION=0.60.0

export DOCKER_BUILDKIT = 1
#export KUBECONFIG = $(CURDIR)/kubeconfig

reload: docker_containerd-shim-wasm-v1 delete create apply

apply:
	kubectl apply -f config/runtime-class.yaml
	kubectl apply -f config/job.yaml

create:
	kind create cluster --name $(CLUSTER_NAME) --config config/kind.yaml

delete:
	kind delete cluster --name $(CLUSTER_NAME)

buildx:
	docker build --platform=local -o . git://github.com/docker/buildx
	mv buildx $(HOME)/.docker/cli-plugins/docker-buildx

UNAME_S = $(shell uname -s)
wasm-cli-plugin:
ifeq ($(UNAME_S),Linux)
	cd /tmp &&
		curl
endif
ifeq ($(UNAME_S),Darwin)

endif

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

containerd-shim-wasm-v1: $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build \
		-o $(BIN_DIR)/containerd-shim-wasm-v1 \
		./cmd/containerd-shim-wasm-v1

wasmtime: $(BIN_DIR)
	cd /tmp && \
		curl -LO https://github.com/bytecodealliance/wasmtime/releases/download/cranelift-v${WASMTIME_VERSION}/wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux.tar.xz && \
		tar -xf wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux.tar.xz && \
		mv wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux/wasmtime $(CURDIR)/$(BIN_DIR)/wasmtime && \
		rm wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux.tar.xz && \
		rm -Rf wasmtime-cranelift-v${WASMTIME_VERSION}-x86_64-linux

build:
	# install plugin
	# https://github.com/tonistiigi/wasm-cli-plugin
	docker buildx build \
		-o $(BIN_DIR) \
		-t $(APP_IMAGE) \
		--platform=wasi/wasm \
		-f ./hello-wasm/Dockerfile .

run:
	wasmtime ./$(BIN_DIR)/hello-wasm

push:
	docker buildx build \
		-t $(APP_IMAGE) \
		--platform=wasi/wasm \
		-f ./hello-wasm/Dockerfile \
		--push .

docker_image:
	docker build -t $(BUILD_IMAGE) $(CURDIR)

docker_shell: docker_image
	@docker run -it \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(CURDIR)/cache/go-build:/root/.cache/go-build \
		-v $(CURDIR)/cache/go-home:/root/go \
		-v $(CURDIR):/workspace \
		$(BUILD_IMAGE)

docker_%: docker_image
	@docker run -it \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(CURDIR)/cache/go-build:/root/.cache/go-build \
		-v $(CURDIR)/cache/go-home:/root/go \
		-v $(CURDIR):/workspace \
		$(BUILD_IMAGE) \
		make $*
