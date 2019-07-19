.PHONY: all clean build install test tf cover protofmt protoc protolint protodocs import-spec

# make sure we turn on go modules
export GO111MODULE := on

# for building out the dexd app
BUILDOUT ?= dexd
BUILD_VERSION ?= $(shell git describe --tags)
BUILD_FLAGS := -mod=readonly -ldflags "-X github.com/iov-one/weave.Version=$(BUILD_VERSION)"

# MODE=count records heat map in test coverage
# MODE=set just records which lines were hit by one test
MODE ?= set

# for dockerized prototool
USER := $(shell id -u):$(shell id -g)
DOCKER_BASE := docker run --rm --user=${USER} -v $(shell pwd):/work iov1/prototool:v0.2.2
PROTOTOOL := $(DOCKER_BASE) prototool
PROTOC := $(DOCKER_BASE) protoc
WEAVEDIR=$(shell go list -m -f '{{.Dir}}' github.com/iov-one/weave)

all: clean test import-spec install

clean:
	rm -f ${BUILDOUT}

build:
	go build $(BUILD_FLAGS) -o $(BUILDOUT) ./cmd/dexd

install:
	go install $(BUILD_FLAGS) ./cmd/dexd

test:
	go vet -mod=readonly ./...
	go test -mod=readonly -race ./...

# Test fast
tf:
	go test -short ./...

test-verbose:
	go vet ./...
	go test -v -race ./...

mod:
	go mod tidy

cover:
	go test -covermode=$(MODE) -coverprofile=coverage/coverage.txt ./...

protolint:
	$(PROTOTOOL) lint

protofmt:
	$(PROTOTOOL) format -w

protodocs:
	# TODO: fix compilation steps and add back to protoc
	./scripts/build_protodocs.sh docs/proto

protoc: protolint protofmt
	$(PROTOTOOL) generate

import-spec:
	@rm -rf ./spec
	@mkdir -p spec/github.com/iov-one/weave
	@cp -r ${WEAVEDIR}/spec/gogo/* spec/github.com/iov-one/weave
	@chmod -R +w spec
	