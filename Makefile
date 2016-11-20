PACKAGES = $(shell go list ./...)

.PHONY: build fmt test test-cover-html test-cover-func collect-cover-data

# Prepend our vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.
export GO15VENDOREXPERIMENT=1
# GOPATH := ${PWD}/vendor:${GOPATH}
# export GOPATH

default: build

build: fmt
	go build -v  -ldflags "-X github.com/Dataman-Cloud/swan/version.Version=0.0.1 -X github.com/Dataman-Cloud/swan/version.Commit=`git rev-parse --short HEAD`  -X github.com/Dataman-Cloud/swan/version.BuildTime=`date -u '+%Y-%m-%d_%I:%M:%S'` " -o swan .


rel: fmt
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -o ./rel/swan .

install:
	install -v swan /usr/local/bin

fmt:
	go fmt ./...

test:
	go test -cover=true ./...

collect-cover-data:
	@echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -v -coverprofile=coverage.out -covermode=count $(pkg) || exit $$?;\
		if [ -f coverage.out ]; then\
			tail -n +2 coverage.out >> coverage-all.out;\
		fi\
		;)
test-cover-html:
	go tool cover -html=coverage-all.out -o coverage.html

test-cover-func:
	go tool cover -func=coverage-all.out
