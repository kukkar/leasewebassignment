### Build stage
# BUILDPLATFORM/TARGETOS/TARGETARCH are populated automatically by
# BuildKit based on whatever platform is actually being built for (set by
# `docker buildx build --platform ...`, or by Render's own build target) -
# a hardcoded --platform literal here would fight that mechanism instead
# of using it (BuildKit's linter flags this: FromPlatformFlagConstDisallowed).
# Building the compiler stage with --platform=$BUILDPLATFORM runs the Go
# toolchain natively on whatever host is doing the build and cross-compiles
# via GOOS/GOARCH, which is correct for any target platform and avoids
# paying for QEMU emulation of the compiler itself.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /leasewebassignment ./cmd

### Final image
# No --platform override here, deliberately: it should resolve to whatever
# the overall build's target platform actually is, decided externally by
# whoever invokes the build - not fixed inside the Dockerfile.
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /leasewebassignment /usr/local/bin/leasewebassignment
COPY config.yaml ./config.yaml
COPY data ./data
COPY web ./web
COPY docs/openapi.yaml ./docs/openapi.yaml
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/leasewebassignment", "-config=config.yaml"]
