PACKAGES = $(shell go list ./...)
TEST_PACKAGES = $(shell go list ./... | grep -v vendor)

.PHONY: build fmt test test-cover-html test-cover-func collect-cover-data

export GO15VENDOREXPERIMENT=1

default: build

build: fmt build-swan

build-swan:
	go build  -ldflags "-X github.com/Dataman-Cloud/swan/srv/version.BuildTime=`date -u +%Y-%m-%d:%H-%M-%S` -X github.com/Dataman-Cloud/swan/src/version.Version=0.01-`git rev-parse --short HEAD`"  -v -o bin/swan main.go

build-swancfg:
	go build -v -o bin/swancfg src/cli/cli.go

install:
	install -v bin/swan /usr/local/bin
	install -v bin/swancfg /usr/local/bin

generate:
	protoc --proto_path=./vendor/github.com/gogo/protobuf/:./src/manager/raft/types/:. --gogo_out=./src/manager/raft/types/ ./src/manager/raft/types/*.proto

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


docker-build:
	docker build --tag swan --rm .

docker-run-mixed:
	docker rm -f swan-mixed-1 2>&1 || echo 0
	docker run --interactive --tty --env-file ./contrib/envfiles/Envfile_mixed --name swan-mixed-1  --rm  -p 9999:9999 -p 2111:2111 -p 53:53/udp -p 80:80 -v `pwd`/data:/go/src/github.com/Dataman-Cloud/swan/data swan

docker-run-manager:
	docker rm -f swan-manager-1 2>&1 || echo 0
	docker run --interactive --tty --env-file ./contrib/envfiles/Envfile_manager --name swan-manager-1  --rm  -p 9999:9999 -p 2111:2111 -v `pwd`/data:/go/src/github.com/Dataman-Cloud/swan/data swan

docker-run-agent:
	docker rm -f swan-agent-1 2>&1 || echo 0
	docker run --interactive --tty --env-file ./contrib/envfiles/Envfile_agent --name swan-agent-1  --rm  -p 9998:9998 -p 53:53/udp -p 80:80  swan

docker-run-agent-2:
	docker rm -f swan-agent-2 2>&1 || echo 0
	docker run --interactive --tty --env-file ./contrib/envfiles/Envfile_agent_2 --name swan-agent-2  --rm  -p 9997:9998 -p 54:53/udp -p 81:80 swan

docker-run-mixed-detached:
	docker rm -f swan-mixed-1 2>&1 || echo 0
	docker run --interactive --tty --env-file ./contrib/envfiles/Envfile_mixed --name swan-mixed-1  -p 9999:9999 -p 2111:2111 -p 53:53/udp -p 80:80 -v `pwd`/data:/go/src/github.com/Dataman-Cloud/swan/data --detach swan

docker-run-mixed-cluster:
	docker rm -f swan-mixed-1 2>&1 || echo 0
	docker run --interactive --tty --env-file ./contrib/envfiles/Envfile_mixed --name swan-mixed-1  --rm  -p 9999:9999 -p 2111:2111 -p 53:53/udp -p 80:80 -v `pwd`/data:/go/src/github.com/Dataman-Cloud/swan/data swan
