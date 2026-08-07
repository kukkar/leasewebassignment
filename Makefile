BINARY := leasewebassignment
GO := go
CMD := ./cmd
CONFIG ?= config.yaml

.PHONY: all build run-server test vet lint verify bench cover docker clean

all: build

build:
	$(GO) build -o bin/$(BINARY) $(CMD)

run-server:
	$(GO) run $(CMD) -config $(CONFIG)

test:
	$(GO) test ./... -race

vet:
	$(GO) vet ./...

lint:
	golangci-lint run ./...

bench:
	$(GO) test ./internal/store/... -run '^$$' -bench . -benchmem

# -coverpkg=./... attributes coverage to the package that owns the code, not
# just the package the test file lives in - without it, internal/server and
# internal/server/handlers show as 0% even though tests/ (the e2e suite)
# exercises them thoroughly through the real HTTP stack rather than in
# isolation.
cover:
	$(GO) test ./... -coverpkg=./... -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out | tail -1

# verify is the single command CI and local pre-submit checks both run.
# It exists specifically so "does it compile" is caught before "do the tests
# pass" ever runs - a plain `go test ./...` will happily report failures from
# unrelated packages while masking a build break in another one. lint is
# deliberately not part of verify: it requires golangci-lint to be installed
# locally, and verify should always work with nothing but the Go toolchain.
# CI runs both verify and lint as separate steps - see .github/workflows/ci.yml.
verify: build vet test

docker: build
	docker build -t $(BINARY):local .

clean:
	rm -rf bin
