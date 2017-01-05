#!/bin/bash

SERVER_PATH=localhost:9999
PATH_PREFIX=v1
APP_NAME=nginx0051-xcm-Unnamed
WAIT_SECOND=10

if ! command -v http &>/dev/null ; then
  echo "httpie not installed, 'apt-get install httpie', 'yum install -y httpie' or  'brew install httpie'"
  exit 1
fi


if ! command -v jq &>/dev/null ; then
  echo "httpie not installed, 'apt-get install jq', 'yum install -y jq' or  'brew install jq'"
  exit 1
fi


# color s
NORMAL='\033[0m'
RED='\033[31m'
GREEN="\033[0;32m"
LGREEN='\033[1;32m'
