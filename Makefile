BINARY := leasewebassignment
GO := go
CMD := ./cmd
CONFIG ?= config.yaml

.PHONY: all build run-server test vet lint verify bench docker clean

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
