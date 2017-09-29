
.PHONY: build image docker docker-centos clean

PACKAGES = $(shell go list ./... | grep -v vendor | grep -v integration-test)
PRJNAME := "swan"
Compose := "https://github.com/docker/compose/releases/download/1.14.0/docker-compose"
TmpDataDir := "/tmp/swan-tmp-data"

# Used to populate version variable in main package.
VERSION=$(shell git describe --always --tags --abbre=0)
BUILD_TIME=$(shell date -u +%Y-%m-%d:%H-%M-%S)
PKG := "github.com/Dataman-Cloud/swan"

gitCommit=$(shell git describe --tags --long)
gitDirty=$(shell git status --porcelain --untracked-files=no)
GIT_COMMIT=$(gitCommit)
ifneq ($(gitDirty),"")
GIT_COMMIT=$(gitCommit)-dirty
endif

GO_LDFLAGS=-X $(PKG)/version.version=$(VERSION) -X $(PKG)/version.gitCommit=$(GIT_COMMIT) -X $(PKG)/version.buildTime=$(BUILD_TIME) -w -s

UNAME=$(shell uname -s)
CGO_ENABLED=0
ifeq ($(UNAME),Darwin)
CGO_ENABLED=1
endif


default: build

build: clean
	CGO_ENABLED=${CGO_ENABLED} go build -v -a -ldflags "${GO_LDFLAGS}" -o bin/swan main.go

# multi-stage builds, require docker >= 17.05
docker:
	docker build --tag swan:$(shell git rev-parse --short HEAD) --rm .

# multi-stage builds, require docker >= 17.05
docker-centos:
	docker build --tag swan:$(shell git rev-parse --short HEAD) --rm -f ./Dockerfile.centos .

# compitable for legacy docker version
docker-build:
	docker run --name=buildswan --rm \
		-w /go/src/github.com/Dataman-Cloud/swan \
		-e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64  \
		-v $(shell pwd):/go/src/github.com/Dataman-Cloud/swan \
		golang:1.8.1-alpine \
		sh -c 'go build -ldflags "${GO_LDFLAGS}" -v -o bin/swan main.go'

# compitable for legacy docker version
docker-image:
	docker build --tag swan:$(shell git rev-parse --short HEAD) --rm -f ./Dockerfile.legacy .
	docker tag swan:$(shell git rev-parse --short HEAD) swan:latest

prepare-docker-compose:
	@if ! command -v docker-compose > /dev/null 2>&1; then \
        echo "docker-compose downloading ..."; \
        curl --progress-bar -L $(Compose)-$(shell uname -s)-$(shell uname -m) -o \
            /usr/local/bin/docker-compose; \
        chmod +x /usr/local/bin/docker-compose; \
        echo "docker-compose downloaded!"; \
    fi

build-binary:
	@if env | grep -q -w "JENKINS_HOME" >/dev/null 2>&1; then \
		$(MAKE) build; \
	else \
		$(MAKE) docker-build; \
	fi

tmpdir:
	mkdir -p $(TmpDataDir)

local-cluster: prepare-docker-compose build-binary docker-image tmpdir
	docker-compose -p ${PRJNAME} up -d
	docker-compose -p ${PRJNAME} ps

rm-local-cluster: prepare-docker-compose
	docker-compose -p ${PRJNAME} stop
	docker-compose -p ${PRJNAME} rm -f
	rm -rf $(TmpDataDir)

check-local-cluster:
	@./contrib/check-local-cluster.sh

integration-prepare: rm-local-cluster local-cluster check-local-cluster

integration-test: integration-prepare run-integration-test rm-local-cluster

run-integration-test:
	docker run --name=testswan --rm \
		-w /go/src/github.com/Dataman-Cloud/swan/integration-test \
		-e SWAN_HOST=$(shell docker inspect -f "{{.NetworkSettings.IPAddress}}" ${PRJNAME}_swan-master_1):9999 \
		-e "TESTON=" \
		-v $(shell pwd):/go/src/github.com/Dataman-Cloud/swan \
		golang:1.8.1-alpine \
		sh -c 'go test -check.v -test.timeout=10m github.com/Dataman-Cloud/swan/integration-test'
clean:
	rm -rfv bin/*

test:
	go test -cover=true ${PACKAGES}

collect-cover-data:
	@echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -v -coverprofile=coverage.out -covermode=count $(pkg) || exit $?;\
		if [ -f coverage.out ]; then\
			tail -n +2 coverage.out >> coverage-all.out;\
		fi\
		;)

test-cover-html:
	go tool cover -html=coverage-all.out -o coverage.html

test-cover-func:
	go tool cover -func=coverage-all.out
