.PHONY: all clean build install test tf cover protofmt protoc protolint protodocs

# make sure we turn on go modules
export GO111MODULE := on

# for building out the mycoind app
BUILDOUT ?= mycoind
BUILD_VERSION ?= $(shell git describe --tags)
BUILD_FLAGS := -ldflags "-X github.com/iov-one/weave.Version=$(BUILD_VERSION)"

# MODE=count records heat map in test coverage
# MODE=set just records which lines were hit by one test
MODE ?= set

# for dockerized prototool
USER := $(shell id -u):$(shell id -g)
DOCKER_BASE := docker run --rm --user $(USER) --mount type=bind,source="$(shell pwd)",target=/work --tmpfs /tmp:exec iov1/prototool:v0.2.0
PROTOTOOL := $(DOCKER_BASE) prototool
PROTOC := $(DOCKER_BASE) protoc

all: clean build test

clean:
	rm -f ${BUILDOUT}

build:
	go build $(BUILD_FLAGS) -o $(BUILDOUT) .

install:
	go install $(BUILD_FLAGS) .

test:
	go vet ./...
	go test -race ./...

# Test fast
tf:
	go test -short ./...

cover:
	go test -covermode=$(MODE) -coverprofile=coverage/coverage.txt ./...

protolint:
	$(PROTOTOOL) lint

protofmt:
	$(PROTOTOOL) format -w

protodocs:
	./scripts/build_protodocs.sh docs/proto

protoc: protolint protofmt protodocs
	$(PROTOTOOL) generate