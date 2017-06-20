PACKAGES = $(shell go list ./...)
TEST_PACKAGES = $(shell go list ./src/... | grep -v vendor)

.PHONY: build fmt test test-cover-html test-cover-func collect-cover-data

export GO15VENDOREXPERIMENT=1

# Used to populate version variable in main package.
VERSION=$(shell git describe --always --tags --abbre=0)
BUILD_TIME=$(shell date -u +%Y-%m-%d:%H-%M-%S)
PKG=$(shell go list .)
gitCommit=$(shell git describe --tags)
gitDirty=$(shell git status --porcelain --untracked-files=no)
GIT_COMMIT=$(gitCommit)
ifneq ($(gitDirty),"")
GIT_COMMIT=$(gitCommit)-dirty
endif
GO_LDFLAGS=-X $(PKG)/src/version.version=$(VERSION) -X $(PKG)/src/version.gitCommit=$(GIT_COMMIT) -X $(PKG)/src/version.buildTime=$(BUILD_TIME) -w


default: build

docker-build:
	docker run --rm -w /go/src/github.com/Dataman-Cloud/swan -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64  -v $(shell pwd):/go/src/github.com/Dataman-Cloud/swan golang:1.8.1-alpine sh -c 'go build -ldflags "${GO_LDFLAGS}" -v -o bin/swan main.go'

build:
	go build -ldflags "${GO_LDFLAGS}" -v -o bin/swan main.go

image: build
	docker build --tag swan:$(shell git rev-parse --short HEAD) --rm .

clean:
	rm -rf bin/*

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
