#!/bin/sh
set -xe

export GOROOT=/usr/lib/golang
export GOPATH=/data/jenkins/workspace/go-jobs
export GOBIN=$GOPATH/bin
export PATH=$PATH:$HOME/bin:$GOROOT/bin:$GOBIN:/usr/local/bin
export GO15VENDOREXPERIMENT=1

cp /root/config.json ./config.json

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

docker run -p 9999:9999 -v $(pwd)/config.json:/go-jobs/config.json -v $(pwd)/data:/go-jobs/data --name=swan -d swan:v1.0

