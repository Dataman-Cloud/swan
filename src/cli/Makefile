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
	go build -v -o swancfg .

install:
	install -v swancfg /usr/local/bin

fmt:
	go fmt ./...
