FROM --platform=$BUILDPLATFORM tonistiigi/xx:llvm AS builder
ARG TARGETPLATFORM
WORKDIR /src
COPY hello-wasm/main.c .
RUN clang -static -o /hello-wasm main.c

FROM scratch
COPY --from=builder /hello-wasm .
CMD ["/hello-wasm"]
