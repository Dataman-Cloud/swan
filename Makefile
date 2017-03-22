PACKAGES = $(shell go list ./...)
TEST_PACKAGES = $(shell go list ./src/... | grep -v vendor)

.PHONY: build fmt test test-cover-html test-cover-func collect-cover-data

export GO15VENDOREXPERIMENT=1

# Used to populate version variable in main package.
VERSION=$(shell git describe --always --tags --abbre=0)
BUILD_TIME=$(shell date -u +%Y-%m-%d:%H-%M-%S)
GO_LDFLAGS=-ldflags "-X `go list ./src/version`.Version=$(VERSION) -X `go list ./src/version`.BuildTime=$(BUILD_TIME)"

default: build

docker-build:
	docker run --rm -w /go/src/github.com/Dataman-Cloud/swan -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64  -v $(shell pwd):/go/src/github.com/Dataman-Cloud/swan golang:1.6.3-alpine sh -c "apk update && apk add make && apk add git && make"

build: fmt
	go build ${GO_LDFLAGS} -v -o bin/swan main.go

install:
	install -v bin/swan /usr/local/bin
	install -v bin/swancfg /usr/local/bin

generate:
	protoc --proto_path=./vendor/github.com/gogo/protobuf/:./src/manager/raft/types/:. --gogo_out=./src/manager/raft/types/ ./src/manager/raft/types/*.proto
	go generate ./src/manager/framework/state/constraints.go

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

release: list-authors

list-authors:
	./contrib/list-authors.sh


build-docker-image:
	docker build --tag swan:$(VERSION) --rm .

docker-run-manager-init:
	docker rm -f swan-manager 2>&1 || echo 0
	docker run --interactive --tty --env-file ./contrib/envfiles/Envfile_manager_init --name swan-manager  --rm  -p 9999:9999 -p 2111:2111 -v `pwd`/data:/data swan:$(VERSION) manager init

docker-run-manager-join:
	docker rm -f swan-manager 2>&1 || echo 0
	docker run --interactive --tty --env-file ./contrib/envfiles/Envfile_manager_join --name swan-manager  --rm  -p 9999:9999 -p 2111:2111 -v `pwd`/data:/data swan:$(VERSION) manager join

docker-run-agent:
	docker rm -f swan-agent 2>&1 || echo 0
	docker run --interactive --tty --env-file ./contrib/envfiles/Envfile_agent --name swan-agent  --rm  -p 9998:9998 -p 53:53/udp -p 80:80  swan:$(VERSION) agent join



