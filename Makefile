CLUSTER_NAME = wasm
BIN_DIR = bin
BUILD_IMAGE = wasm
APP_IMAGE = dippynark/hello-wasm:v0.1.0

WASMER_VERSION=0.16.2
DOCKER_WASM_VERSION = v0.2.0

export DOCKER_BUILDKIT = 1

deploy: docker_shim docker_wasmer create apply

reload: docker_shim
	kubectl delete pod --all

apply:
	kubectl apply -f config/runtime-class.yaml
	kubectl apply -f config/deployment.yaml

create:
	kind create cluster --name $(CLUSTER_NAME) --config config/kind.yaml

delete:
	kind delete cluster --name $(CLUSTER_NAME)

init:
	mkdir -p $(BIN_DIR)

plugin_buildx:
	docker build --platform=local -o $(BIN_DIR) git://github.com/docker/buildx
	cp -a $(BIN_DIR)/buildx $(HOME)/.docker/cli-plugins/docker-buildx

plugin_wasm:
	curl -Ls https://github.com/tonistiigi/wasm-cli-plugin/releases/download/$(DOCKER_WASM_VERSION)/docker-wasm-$(DOCKER_WASM_VERSION).$(shell uname -s | tr '[:upper:]' '[:lower:]')-amd64.tar.gz \
		| tar -xzOf - docker-wasm > $(BIN_DIR)/docker-wasm
	chmod +x $(BIN_DIR)/docker-wasm
	cp -a $(BIN_DIR)/docker-wasm $(HOME)/.docker/cli-plugins/docker-wasm

wasmer:
	cp -a `which wasmer` $(BIN_DIR)/wasmer

shim:
	GOOS=linux GOARCH=amd64 go build \
		-o $(BIN_DIR)/containerd-shim-wasm-v1 \
		./cmd/containerd-shim-wasm-v1

hello_wasm: build push

run:
	wasmer ./$(BIN_DIR)/hello-wasm

build:
	# install plugin
	# https://github.com/tonistiigi/wasm-cli-plugin
	docker buildx build \
		-o $(BIN_DIR) \
		-t $(APP_IMAGE) \
		--platform=wasi/wasm \
		-f ./hello-wasm/Dockerfile .

push:
	docker buildx build \
		-t $(APP_IMAGE) \
		--platform=wasi/wasm \
		-f ./hello-wasm/Dockerfile \
		--push .

docker_image:
	docker build -t $(BUILD_IMAGE) \
		--build-arg WASMER_VERSION=$(WASMER_VERSION) \
		$(CURDIR)

docker_shell: docker_image
	docker run -it \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(CURDIR)/cache/go-build:/root/.cache/go-build \
		-v $(CURDIR)/cache/go-home:/root/go \
		-v $(CURDIR):/workspace \
		$(BUILD_IMAGE)

docker_%: docker_image
	docker run -it \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(CURDIR)/cache/go-build:/root/.cache/go-build \
		-v $(CURDIR)/cache/go-home:/root/go \
		-v $(CURDIR):/workspace \
		$(BUILD_IMAGE) \
		make $*
