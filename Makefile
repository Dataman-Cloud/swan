PACKAGES = $(shell go list ./...)
TEST_PACKAGES = $(shell go list ./... | grep -v scheduler | grep -v vendor)

.PHONY: build fmt test test-cover-html test-cover-func collect-cover-data

# Prepend our vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
export GO15VENDOREXPERIMENT=1
# GOPATH := ${PWD}/vendor:${GOPATH}
# export GOPATH

default: build

build: fmt build-swancfg build-swan

build-swan:
	go build -v -o bin/swan main.go node.go

build-swancfg:
	go build -v -o bin/swancfg src/cli/cli.go

install:
	install -v bin/swan /usr/local/bin
	install -v bin/swancfg /usr/local/bin

generate:
	protoc --proto_path=./vendor/github.com/gogo/protobuf/:./src/manager/raft/tyeps/:. --gogo_out=./src/manager/raft/tyeps/ ./src/manager/raft/tyeps/*.proto

clean:
	rm -rf bin/*

fmt:
	go fmt ./src/...

test:
	go test -cover=true ${TEST_PACKAGES}

collect-cover-data:
	@echo "mode: count" > coverage-all.out
	$(foreach pkg,$(TEST_PACKAGES),\
		go test -v -coverprofile=coverage.out -covermode=count $(pkg) || exit $?;\
		if [ -f coverage.out ]; then\
			tail -n +2 coverage.out >> coverage-all.out;\
		fi\
		;)
test-cover-html:
	go tool cover -html=coverage-all.out -o coverage.html

test-cover-func:
	go tool cover -func=coverage-all.out
