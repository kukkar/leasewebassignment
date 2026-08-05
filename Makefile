BINARY := leasewebassignment
GO := go
CMD := ./cmd
CONFIG ?= config.yaml

all: build

build:
	$(GO) build -o bin/$(BINARY) $(CMD)

run-server:
	$(GO) run $(CMD) -config $(CONFIG)

test:
	$(GO) test ./...

clean:
	rm -rf bin
