#!/bin/sh
set -xe

export GOROOT=/usr/lib/golang
export GOPATH=/data/jenkins/workspace/go-jobs
export GOBIN=$GOPATH/bin
export PATH=$PATH:$HOME/bin:$GOROOT/bin:$GOBIN:/usr/local/bin
export GO15VENDOREXPERIMENT=1

cp /var/.swan.config.json ./config.json

make build-swan

if [ `docker ps -a|grep swan|wc -l` -gt 0 ]
then
docker rm -f `docker ps -a|grep swan|awk '{print $1}'`
fi

if [ `docker images|grep swan|wc -l` -gt 0 ]
then
docker rmi -f `docker images|grep swan|awk '{print $3}'`
fi

rm -rf ./data/*

# build new docker image
docker build -f dockerfiles/Dockerfile_runtime -t swan:v1.0 .

docker run -v $(pwd)/config.json:/go-swan/config.json -v $(pwd)/data:/go-swan/data --net host --name=swan -d swan:v1.0

